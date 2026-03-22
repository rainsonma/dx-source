package adm

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/goravel/framework/facades"
	"dx-api/app/models"

	"github.com/oklog/ulid/v2"
)

// AdminRedeemItem represents a redeem code in the admin list.
type AdminRedeemItem struct {
	ID         string  `json:"id"`
	Code       string  `json:"code"`
	Grade      string  `json:"grade"`
	UserID     *string `json:"userId"`
	Username   *string `json:"username"`
	Nickname   *string `json:"nickname"`
	RedeemedAt any     `json:"redeemedAt"`
	CreatedAt  any     `json:"createdAt"`
}

// GenerateCodes creates a batch of redeem codes for a given grade.
func GenerateCodes(grade string, count int) (int, error) {
	codes := make(map[string]bool, count)
	var redeems []models.UserRedeem

	for len(codes) < count {
		code := generateRedeemCode()
		if codes[code] {
			continue
		}
		codes[code] = true
		redeems = append(redeems, models.UserRedeem{
			ID:    ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
			Code:  code,
			Grade: grade,
		})
	}

	for _, r := range redeems {
		if err := facades.Orm().Query().Create(&r); err != nil {
			return 0, fmt.Errorf("failed to create redeem code: %w", err)
		}
	}

	return len(redeems), nil
}

// GetAllRedeems returns all redeem codes with user info (admin, paginated).
func GetAllRedeems(page, pageSize int) ([]AdminRedeemItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 15
	}
	offset := (page - 1) * pageSize

	total, err := facades.Orm().Query().Model(&models.UserRedeem{}).Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count redeems: %w", err)
	}

	type redeemRow struct {
		ID         string  `gorm:"column:id"`
		Code       string  `gorm:"column:code"`
		Grade      string  `gorm:"column:grade"`
		UserID     *string `gorm:"column:user_id"`
		Username   *string `gorm:"column:username"`
		Nickname   *string `gorm:"column:nickname"`
		RedeemedAt any     `gorm:"column:redeemed_at"`
		CreatedAt  any     `gorm:"column:created_at"`
	}

	var rows []redeemRow
	if err := facades.Orm().Query().Raw(`
		SELECT r.id, r.code, r.grade, r.user_id, r.redeemed_at, r.created_at,
		       u.username, u.nickname
		FROM user_redeems r
		LEFT JOIN users u ON u.id = r.user_id
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset).Scan(&rows); err != nil {
		return nil, 0, fmt.Errorf("failed to query redeems: %w", err)
	}

	items := make([]AdminRedeemItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, AdminRedeemItem{
			ID:         r.ID,
			Code:       r.Code,
			Grade:      r.Grade,
			UserID:     r.UserID,
			Username:   r.Username,
			Nickname:   r.Nickname,
			RedeemedAt: r.RedeemedAt,
			CreatedAt:  r.CreatedAt,
		})
	}

	return items, total, nil
}

const redeemCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateRedeemCode creates a code in XXXX-XXXX-XXXX-XXXX format.
func generateRedeemCode() string {
	groups := make([]string, 4)
	for g := 0; g < 4; g++ {
		var sb strings.Builder
		sb.Grow(4)
		for i := 0; i < 4; i++ {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(redeemCharset))))
			sb.WriteByte(redeemCharset[n.Int64()])
		}
		groups[g] = sb.String()
	}
	return strings.Join(groups, "-")
}
