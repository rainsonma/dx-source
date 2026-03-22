package api

import (
	"fmt"

	"github.com/goravel/framework/facades"
	"dx-api/app/models"
)

// InviteData contains invite URL, stats, and first page of referrals.
type InviteData struct {
	InviteCode     string          `json:"inviteCode"`
	Stats          InviteStats     `json:"stats"`
	Referrals      []ReferralItem  `json:"referrals"`
	TotalReferrals int64           `json:"totalReferrals"`
}

// InviteStats holds aggregated referral statistics.
type InviteStats struct {
	Total    int64 `json:"total"`
	Pending  int64 `json:"pending"`
	Paid     int64 `json:"paid"`
	Rewarded int64 `json:"rewarded"`
}

// ReferralItem represents a single referral record with invitee info.
type ReferralItem struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	RewardAmount float64 `json:"rewardAmount"`
	RewardedAt   any     `json:"rewardedAt"`
	CreatedAt    any     `json:"createdAt"`
	Invitee      *ReferralInvitee `json:"invitee"`
}

// ReferralInvitee contains the invited user's basic info.
type ReferralInvitee struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Nickname *string `json:"nickname"`
	Grade    string  `json:"grade"`
}

// GetInviteData returns the user's invite code, referral stats, and first page of referrals.
func GetInviteData(userID string) (*InviteData, error) {
	var user models.User
	if err := facades.Orm().Query().Select("id", "invite_code").Where("id", userID).First(&user); err != nil || user.ID == "" {
		return nil, ErrUserNotFound
	}

	// Get total count
	total, err := facades.Orm().Query().Model(&models.UserReferral{}).Where("referrer_id", userID).Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count referrals: %w", err)
	}

	// Get stats by status
	type statusCount struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	var counts []statusCount
	if err := facades.Orm().Query().Raw(`
		SELECT status, COUNT(*) AS count
		FROM user_referrals
		WHERE referrer_id = ?
		GROUP BY status
	`, userID).Scan(&counts); err != nil {
		return nil, fmt.Errorf("failed to query referral stats: %w", err)
	}

	stats := InviteStats{Total: total}
	for _, c := range counts {
		switch c.Status {
		case "pending":
			stats.Pending = c.Count
		case "paid":
			stats.Paid = c.Count
		case "rewarded":
			stats.Rewarded = c.Count
		}
	}

	// Get first page of referrals
	referrals, err := GetReferrals(userID, 1, 15)
	if err != nil {
		return nil, err
	}

	return &InviteData{
		InviteCode:     user.InviteCode,
		Stats:          stats,
		Referrals:      referrals,
		TotalReferrals: total,
	}, nil
}

// GetReferrals returns paginated referral records for a user.
func GetReferrals(userID string, page, pageSize int) ([]ReferralItem, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 15
	}
	offset := (page - 1) * pageSize

	type referralRow struct {
		ID              string  `gorm:"column:id"`
		Status          string  `gorm:"column:status"`
		RewardAmount    float64 `gorm:"column:reward_amount"`
		RewardedAt      any     `gorm:"column:rewarded_at"`
		CreatedAt       any     `gorm:"column:created_at"`
		InviteeID       *string `gorm:"column:invitee_id"`
		InviteeUsername *string `gorm:"column:invitee_username"`
		InviteeNickname *string `gorm:"column:invitee_nickname"`
		InviteeGrade    *string `gorm:"column:invitee_grade"`
	}

	var rows []referralRow
	if err := facades.Orm().Query().Raw(`
		SELECT r.id, r.status, r.reward_amount, r.rewarded_at, r.created_at,
		       u.id AS invitee_id, u.username AS invitee_username,
		       u.nickname AS invitee_nickname, u.grade AS invitee_grade
		FROM user_referrals r
		LEFT JOIN users u ON u.id = r.invitee_id
		WHERE r.referrer_id = ?
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, pageSize, offset).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query referrals: %w", err)
	}

	results := make([]ReferralItem, 0, len(rows))
	for _, r := range rows {
		item := ReferralItem{
			ID:           r.ID,
			Status:       r.Status,
			RewardAmount: r.RewardAmount,
			RewardedAt:   r.RewardedAt,
			CreatedAt:    r.CreatedAt,
		}
		if r.InviteeID != nil {
			item.Invitee = &ReferralInvitee{
				ID:       *r.InviteeID,
				Username: derefStr(r.InviteeUsername),
				Nickname: r.InviteeNickname,
				Grade:    derefStr(r.InviteeGrade),
			}
		}
		results = append(results, item)
	}

	return results, nil
}

// CountReferrals returns the total number of referrals for a user.
func CountReferrals(userID string) (int64, error) {
	total, err := facades.Orm().Query().Model(&models.UserReferral{}).Where("referrer_id", userID).Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count referrals: %w", err)
	}
	return total, nil
}

// derefStr safely dereferences a *string, returning "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
