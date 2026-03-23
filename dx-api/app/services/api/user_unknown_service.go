package api

import (
	"fmt"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// UnknownStatsData holds unknown word statistics.
type UnknownStatsData struct {
	Total         int64 `json:"total"`
	Today         int64 `json:"today"`
	LastThreeDays int64 `json:"lastThreeDays"`
}

// MarkAsUnknown upserts an unknown entry.
func MarkAsUnknown(userID, contentItemID, gameID, gameLevelID string) error {
	allowed, err := helpers.CheckRateLimit(
		fmt.Sprintf(rateLimitUnknownKey, userID), rateLimitTracking, rateLimitTrackingSec,
	)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	_, err = facades.Orm().Query().Exec(
		`INSERT INTO user_unknowns (id, user_id, content_item_id, game_id, game_level_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, now(), now())
		 ON CONFLICT (user_id, content_item_id) DO NOTHING`,
		newID(), userID, contentItemID, gameID, gameLevelID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark as unknown: %w", err)
	}
	return nil
}

// ListUnknown returns paginated unknown items with content details.
func ListUnknown(userID, cursor string, limit int) ([]TrackingItemData, string, bool, error) {
	query := facades.Orm().Query()

	var unknowns []models.UserUnknown
	q := query.Where("user_id", userID).Order("created_at desc").Limit(limit + 1)
	if cursor != "" {
		var cursorItem models.UserUnknown
		if err := query.Where("id", cursor).First(&cursorItem); err == nil && cursorItem.ID != "" {
			q = q.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursorItem.CreatedAt, cursorItem.CreatedAt, cursor)
		}
	}
	if err := q.Get(&unknowns); err != nil {
		return nil, "", false, fmt.Errorf("failed to list unknown: %w", err)
	}

	hasMore := len(unknowns) > limit
	if hasMore {
		unknowns = unknowns[:limit]
	}

	nextCursor := ""
	if hasMore && len(unknowns) > 0 {
		nextCursor = unknowns[len(unknowns)-1].ID
	}

	return enrichTrackingItems(nil, unknowns), nextCursor, hasMore, nil
}

// GetUnknownStats returns unknown word count statistics.
func GetUnknownStats(userID string) (*UnknownStatsData, error) {
	query := facades.Orm().Query()
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	threeDaysAgo := startOfToday.AddDate(0, 0, -3)

	total, _ := query.Model(&models.UserUnknown{}).Where("user_id", userID).Count()
	today, _ := query.Model(&models.UserUnknown{}).Where("user_id", userID).Where("created_at >= ?", startOfToday).Count()
	lastThree, _ := query.Model(&models.UserUnknown{}).Where("user_id", userID).Where("created_at >= ?", threeDaysAgo).Count()

	return &UnknownStatsData{Total: total, Today: today, LastThreeDays: lastThree}, nil
}

// DeleteUnknown removes a single unknown entry owned by the user.
func DeleteUnknown(userID, id string) error {
	_, err := facades.Orm().Query().Exec(
		"DELETE FROM user_unknowns WHERE id = ? AND user_id = ?", id, userID,
	)
	return err
}

// BulkDeleteUnknown removes multiple unknown entries owned by the user.
func BulkDeleteUnknown(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := facades.Orm().Query().Exec(
		"DELETE FROM user_unknowns WHERE user_id = ? AND id IN ?", userID, ids,
	)
	return err
}
