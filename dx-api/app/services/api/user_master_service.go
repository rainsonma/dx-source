package api

import (
	"fmt"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
)

// MasterStatsData holds mastered word statistics.
type MasterStatsData struct {
	Total     int64 `json:"total"`
	ThisWeek  int64 `json:"thisWeek"`
	ThisMonth int64 `json:"thisMonth"`
}

// MarkAsMastered upserts a mastered entry and removes from unknown.
func MarkAsMastered(userID, contentItemID, gameID, gameLevelID string) error {
	allowed, err := helpers.CheckRateLimit(
		fmt.Sprintf(rateLimitMasterKey, userID), rateLimitTracking, rateLimitTrackingSec,
	)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	now := carbon.Now()
	_, err = facades.Orm().Query().Exec(
		`INSERT INTO user_masters (id, user_id, content_item_id, game_id, game_level_id, mastered_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, now(), now())
		 ON CONFLICT (user_id, content_item_id) DO NOTHING`,
		newID(), userID, contentItemID, gameID, gameLevelID, now,
	)
	if err != nil {
		return fmt.Errorf("failed to mark as mastered: %w", err)
	}

	// Remove from unknown if exists
	_, _ = facades.Orm().Query().Exec(
		"DELETE FROM user_unknowns WHERE user_id = ? AND content_item_id = ?",
		userID, contentItemID,
	)

	return nil
}

// ListMastered returns paginated mastered items with content details.
func ListMastered(userID, cursor string, limit int) ([]TrackingItemData, string, bool, error) {
	query := facades.Orm().Query()

	var masters []models.UserMaster
	q := query.Where("user_id", userID).Order("mastered_at desc").Limit(limit + 1)
	if cursor != "" {
		var cursorItem models.UserMaster
		if err := query.Where("id", cursor).First(&cursorItem); err == nil && cursorItem.ID != "" {
			q = q.Where("mastered_at <= ?", cursorItem.MasteredAt).Where("id != ?", cursor)
		}
	}
	if err := q.Get(&masters); err != nil {
		return nil, "", false, fmt.Errorf("failed to list mastered: %w", err)
	}

	hasMore := len(masters) > limit
	if hasMore {
		masters = masters[:limit]
	}

	nextCursor := ""
	if hasMore && len(masters) > 0 {
		nextCursor = masters[len(masters)-1].ID
	}

	return enrichTrackingItems(masters, nil), nextCursor, hasMore, nil
}

// GetMasterStats returns mastered word count statistics.
func GetMasterStats(userID string) (*MasterStatsData, error) {
	query := facades.Orm().Query()
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	total, _ := query.Model(&models.UserMaster{}).Where("user_id", userID).Count()
	thisWeek, _ := query.Model(&models.UserMaster{}).Where("user_id", userID).Where("mastered_at >= ?", startOfWeek).Count()
	thisMonth, _ := query.Model(&models.UserMaster{}).Where("user_id", userID).Where("mastered_at >= ?", startOfMonth).Count()

	return &MasterStatsData{Total: total, ThisWeek: thisWeek, ThisMonth: thisMonth}, nil
}

// DeleteMastered removes a single mastered entry owned by the user.
func DeleteMastered(userID, id string) error {
	_, err := facades.Orm().Query().Exec(
		"DELETE FROM user_masters WHERE id = ? AND user_id = ?", id, userID,
	)
	return err
}

// BulkDeleteMastered removes multiple mastered entries owned by the user.
func BulkDeleteMastered(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := facades.Orm().Query().Exec(
		"DELETE FROM user_masters WHERE user_id = ? AND id IN ?", userID, ids,
	)
	return err
}
