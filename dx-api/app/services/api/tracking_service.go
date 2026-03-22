package api

import (
	"fmt"
	"time"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/support/carbon"
)

const (
	rateLimitMasterKey  = "ratelimit:mark-mastered:%s"
	rateLimitUnknownKey = "ratelimit:mark-unknown:%s"
	rateLimitReviewKey  = "ratelimit:mark-review:%s"
	rateLimitTracking   = 30
	rateLimitTrackingSec = 60
)

// --- DTOs ---

// TrackingItemData represents a mastered/unknown/review item with content details.
type TrackingItemData struct {
	ID          string  `json:"id"`
	ContentItem *TrackingContentData `json:"contentItem"`
	GameName    *string `json:"gameName"`
	MasteredAt  any     `json:"masteredAt,omitempty"`
	CreatedAt   any     `json:"createdAt"`
}

// TrackingContentData holds content item details for tracking lists.
type TrackingContentData struct {
	Content     string  `json:"content"`
	Translation *string `json:"translation"`
	ContentType string  `json:"contentType"`
}

// ReviewItemData extends TrackingItemData with review-specific fields.
type ReviewItemData struct {
	ID           string               `json:"id"`
	ContentItem  *TrackingContentData `json:"contentItem"`
	GameID       string               `json:"gameId"`
	GameName     *string              `json:"gameName"`
	LastReviewAt any                  `json:"lastReviewAt"`
	NextReviewAt any                  `json:"nextReviewAt"`
	ReviewCount  int                  `json:"reviewCount"`
	CreatedAt    any                  `json:"createdAt"`
}

// MasterStatsData holds mastered word statistics.
type MasterStatsData struct {
	Total     int64 `json:"total"`
	ThisWeek  int64 `json:"thisWeek"`
	ThisMonth int64 `json:"thisMonth"`
}

// UnknownStatsData holds unknown word statistics.
type UnknownStatsData struct {
	Total         int64 `json:"total"`
	Today         int64 `json:"today"`
	LastThreeDays int64 `json:"lastThreeDays"`
}

// ReviewStatsData holds review statistics.
type ReviewStatsData struct {
	Pending       int64 `json:"pending"`
	Overdue       int64 `json:"overdue"`
	ReviewedToday int64 `json:"reviewedToday"`
}

// --- Mastered ---

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

// --- Unknown ---

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
			q = q.Where("created_at <= ?", cursorItem.CreatedAt).Where("id != ?", cursor)
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

// --- Review ---

// MarkAsReview creates a review entry if it doesn't exist, with spaced repetition schedule.
func MarkAsReview(userID, contentItemID, gameID, gameLevelID string) error {
	allowed, err := helpers.CheckRateLimit(
		fmt.Sprintf(rateLimitReviewKey, userID), rateLimitTracking, rateLimitTrackingSec,
	)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	nextReview := constants.GetNextReviewAt(0)
	nextReviewCarbon := carbon.FromStdTime(nextReview)

	_, err = facades.Orm().Query().Exec(
		`INSERT INTO user_reviews (id, user_id, content_item_id, game_id, game_level_id, next_review_at, review_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 0, now(), now())
		 ON CONFLICT (user_id, content_item_id) DO NOTHING`,
		newID(), userID, contentItemID, gameID, gameLevelID, nextReviewCarbon,
	)
	if err != nil {
		return fmt.Errorf("failed to mark for review: %w", err)
	}
	return nil
}

// ListReviews returns paginated review items ordered by urgency (nextReviewAt ASC).
func ListReviews(userID, cursor string, limit int) ([]ReviewItemData, string, bool, error) {
	query := facades.Orm().Query()

	var reviews []models.UserReview
	q := query.Where("user_id", userID).Order("next_review_at asc").Limit(limit + 1)
	if cursor != "" {
		var cursorItem models.UserReview
		if err := query.Where("id", cursor).First(&cursorItem); err == nil && cursorItem.ID != "" {
			q = q.Where("next_review_at >= ?", cursorItem.NextReviewAt).Where("id != ?", cursor)
		}
	}
	if err := q.Get(&reviews); err != nil {
		return nil, "", false, fmt.Errorf("failed to list reviews: %w", err)
	}

	hasMore := len(reviews) > limit
	if hasMore {
		reviews = reviews[:limit]
	}

	nextCursor := ""
	if hasMore && len(reviews) > 0 {
		nextCursor = reviews[len(reviews)-1].ID
	}

	// Enrich with content item details
	items := make([]ReviewItemData, 0, len(reviews))
	if len(reviews) == 0 {
		return items, nextCursor, hasMore, nil
	}

	contentIDs := make([]string, 0, len(reviews))
	gameIDs := make([]string, 0, len(reviews))
	for _, r := range reviews {
		contentIDs = append(contentIDs, r.ContentItemID)
		gameIDs = append(gameIDs, r.GameID)
	}

	contentMap := batchLoadContentItems(contentIDs)
	gameMap := batchLoadGameNames(gameIDs)

	for _, r := range reviews {
		item := ReviewItemData{
			ID:           r.ID,
			GameID:       r.GameID,
			LastReviewAt: r.LastReviewAt,
			NextReviewAt: r.NextReviewAt,
			ReviewCount:  r.ReviewCount,
			CreatedAt:    r.CreatedAt,
		}
		if ci, ok := contentMap[r.ContentItemID]; ok {
			item.ContentItem = ci
		}
		if name, ok := gameMap[r.GameID]; ok {
			item.GameName = &name
		}
		items = append(items, item)
	}

	return items, nextCursor, hasMore, nil
}

// GetReviewStats returns review statistics: pending, overdue, reviewedToday.
func GetReviewStats(userID string) (*ReviewStatsData, error) {
	query := facades.Orm().Query()
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	pending, _ := query.Model(&models.UserReview{}).Where("user_id", userID).Where("next_review_at <= ?", now).Count()
	overdue, _ := query.Model(&models.UserReview{}).Where("user_id", userID).Where("next_review_at < ?", startOfToday).Count()
	reviewedToday, _ := query.Model(&models.UserReview{}).Where("user_id", userID).Where("last_review_at >= ?", startOfToday).Count()

	return &ReviewStatsData{Pending: pending, Overdue: overdue, ReviewedToday: reviewedToday}, nil
}

// DeleteReview removes a single review entry owned by the user.
func DeleteReview(userID, id string) error {
	_, err := facades.Orm().Query().Exec(
		"DELETE FROM user_reviews WHERE id = ? AND user_id = ?", id, userID,
	)
	return err
}

// BulkDeleteReviews removes multiple review entries owned by the user.
func BulkDeleteReviews(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := facades.Orm().Query().Exec(
		"DELETE FROM user_reviews WHERE user_id = ? AND id IN ?", userID, ids,
	)
	return err
}

// --- Helpers ---

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
