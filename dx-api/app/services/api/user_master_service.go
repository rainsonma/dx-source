package api

import (
	"fmt"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
)

const (
	rateLimitMasterKey   = "ratelimit:mark-mastered:%s"
	rateLimitUnknownKey  = "ratelimit:mark-unknown:%s"
	rateLimitReviewKey   = "ratelimit:mark-review:%s"
	rateLimitTracking    = 30
	rateLimitTrackingSec = 60
)

// TrackingItemData represents a mastered/unknown/review item with content details.
type TrackingItemData struct {
	ID          string               `json:"id"`
	ContentItem *TrackingContentData `json:"contentItem"`
	GameName    *string              `json:"gameName"`
	MasteredAt  any                  `json:"masteredAt,omitempty"`
	CreatedAt   any                  `json:"createdAt"`
}

// TrackingContentData holds content item details for tracking lists.
type TrackingContentData struct {
	Content     string  `json:"content"`
	Translation *string `json:"translation"`
	ContentType string  `json:"contentType"`
}

// enrichTrackingItems enriches master or unknown items with content and game details.
func enrichTrackingItems(masters []models.UserMaster, unknowns []models.UserUnknown) []TrackingItemData {
	var contentIDs, gameIDs []string
	var items []TrackingItemData

	if masters != nil {
		contentIDs = make([]string, 0, len(masters))
		gameIDs = make([]string, 0, len(masters))
		for _, m := range masters {
			contentIDs = append(contentIDs, m.ContentItemID)
			gameIDs = append(gameIDs, m.GameID)
		}
	} else if unknowns != nil {
		contentIDs = make([]string, 0, len(unknowns))
		gameIDs = make([]string, 0, len(unknowns))
		for _, u := range unknowns {
			contentIDs = append(contentIDs, u.ContentItemID)
			gameIDs = append(gameIDs, u.GameID)
		}
	}

	contentMap := batchLoadContentItems(contentIDs)
	gameMap := batchLoadGameNames(gameIDs)

	if masters != nil {
		items = make([]TrackingItemData, 0, len(masters))
		for _, m := range masters {
			item := TrackingItemData{
				ID:         m.ID,
				MasteredAt: m.MasteredAt,
				CreatedAt:  m.CreatedAt,
			}
			if ci, ok := contentMap[m.ContentItemID]; ok {
				item.ContentItem = ci
			}
			if name, ok := gameMap[m.GameID]; ok {
				item.GameName = &name
			}
			items = append(items, item)
		}
	} else if unknowns != nil {
		items = make([]TrackingItemData, 0, len(unknowns))
		for _, u := range unknowns {
			item := TrackingItemData{
				ID:        u.ID,
				CreatedAt: u.CreatedAt,
			}
			if ci, ok := contentMap[u.ContentItemID]; ok {
				item.ContentItem = ci
			}
			if name, ok := gameMap[u.GameID]; ok {
				item.GameName = &name
			}
			items = append(items, item)
		}
	}

	return items
}

// batchLoadContentItems loads content items by IDs and returns a map.
func batchLoadContentItems(ids []string) map[string]*TrackingContentData {
	result := make(map[string]*TrackingContentData)
	if len(ids) == 0 {
		return result
	}
	var items []models.ContentItem
	facades.Orm().Query().Where("id IN ?", ids).Get(&items)
	for _, ci := range items {
		result[ci.ID] = &TrackingContentData{
			Content:     ci.Content,
			Translation: ci.Translation,
			ContentType: ci.ContentType,
		}
	}
	return result
}

// batchLoadGameNames loads game names by IDs and returns a map.
func batchLoadGameNames(ids []string) map[string]string {
	result := make(map[string]string)
	if len(ids) == 0 {
		return result
	}
	var games []models.Game
	facades.Orm().Query().Where("id IN ?", ids).Get(&games)
	for _, g := range games {
		result[g.ID] = g.Name
	}
	return result
}

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
