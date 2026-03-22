package api

import (
	"fmt"
	"time"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/models"

	"github.com/goravel/framework/support/carbon"
)

// RedeemItem represents a user's redemption record.
type RedeemItem struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Grade      string `json:"grade"`
	RedeemedAt any    `json:"redeemedAt"`
}

// RedeemCodeResult contains the result of a code redemption.
type RedeemCodeResult struct {
	Grade string `json:"grade"`
}

// GetRedeems returns a user's redemption records (paginated).
func GetRedeems(userID string, page, pageSize int) ([]RedeemItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 15
	}
	offset := (page - 1) * pageSize

	total, err := facades.Orm().Query().Model(&models.UserRedeem{}).
		Where("user_id", userID).Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count redeems: %w", err)
	}

	var redeems []models.UserRedeem
	if err := facades.Orm().Query().
		Where("user_id", userID).
		Order("redeemed_at desc").
		Offset(offset).Limit(pageSize).
		Get(&redeems); err != nil {
		return nil, 0, fmt.Errorf("failed to query redeems: %w", err)
	}

	items := make([]RedeemItem, 0, len(redeems))
	for _, r := range redeems {
		items = append(items, RedeemItem{
			ID:         r.ID,
			Code:       r.Code,
			Grade:      r.Grade,
			RedeemedAt: r.RedeemedAt,
		})
	}

	return items, total, nil
}

// RedeemCode processes a redemption code transactionally.
// Steps: verify unused → mark redeemed → update user VIP → grant beans.
func RedeemCode(userID, code string) (*RedeemCodeResult, error) {
	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// Find the redeem code
	var redeem models.UserRedeem
	if err := tx.Where("code", code).LockForUpdate().First(&redeem); err != nil || redeem.ID == "" {
		_ = tx.Rollback()
		return nil, ErrRedeemNotFound
	}

	// Check if already used
	if redeem.UserID != nil && *redeem.UserID != "" {
		_ = tx.Rollback()
		return nil, ErrRedeemAlreadyUsed
	}

	// Mark as redeemed
	now := carbon.Now()
	if _, err := tx.Exec(
		"UPDATE user_redeems SET user_id = ?, redeemed_at = ?, updated_at = NOW() WHERE id = ?",
		userID, now, redeem.ID,
	); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to mark redeem as used: %w", err)
	}

	// Get user for VIP calculation
	var user models.User
	if err := tx.Where("id", userID).LockForUpdate().First(&user); err != nil || user.ID == "" {
		_ = tx.Rollback()
		return nil, ErrUserNotFound
	}

	// Calculate new VIP due date
	newVipDueAt := calcVipDueAt(redeem.Grade, user.VipDueAt)

	// Update user grade and VIP
	if _, err := tx.Exec(
		"UPDATE users SET grade = ?, vip_due_at = ?, updated_at = NOW() WHERE id = ?",
		redeem.Grade, newVipDueAt, userID,
	); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to update user VIP: %w", err)
	}

	// Grant beans
	beanAmount := 10000
	if redeem.Grade == constants.UserGradeLifetime {
		beanAmount = 15000
	}

	newBalance := user.Beans + beanAmount
	if _, err := tx.Model(&models.User{}).Where("id", userID).
		Update("beans", newBalance); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to grant beans: %w", err)
	}

	ledger := models.UserBean{
		ID:     newID(),
		UserID: userID,
		Beans:  beanAmount,
		Origin: user.Beans,
		Result: newBalance,
		Slug:   constants.BeanSlugMembershipGrant,
		Reason: constants.BeanReasonMembershipGrant,
	}
	if err := tx.Create(&ledger); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to create bean ledger: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit redeem transaction: %w", err)
	}

	return &RedeemCodeResult{Grade: redeem.Grade}, nil
}

// calcVipDueAt computes the new VIP expiration date.
func calcVipDueAt(grade string, currentDueAt *carbon.DateTime) *time.Time {
	if grade == constants.UserGradeLifetime {
		return nil
	}

	months := constants.UserGradeMonths[grade]
	if months == 0 {
		return nil
	}

	now := time.Now()
	var base time.Time

	if currentDueAt != nil {
		dueTime := currentDueAt.StdTime()
		if dueTime.After(now) {
			base = dueTime
		} else {
			base = now
		}
	} else {
		base = now
	}

	// Add months and subtract 1 day
	result := base.AddDate(0, months, -1)

	// Handle month overflow: if day changed unexpectedly, clamp to last day of month
	expectedMonth := (int(base.Month()) + months - 1) % 12
	if expectedMonth == 0 {
		expectedMonth = 12
	}
	if int(result.Month()) != expectedMonth {
		// Went past the end of month, clamp
		result = time.Date(result.Year(), result.Month(), 0, 0, 0, 0, 0, result.Location())
	}

	return &result
}
