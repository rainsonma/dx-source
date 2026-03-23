package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
)

// ReviewStatsData holds review statistics.
type ReviewStatsData struct {
	Pending       int64 `json:"pending"`
	Overdue       int64 `json:"overdue"`
	ReviewedToday int64 `json:"reviewedToday"`
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

	nextReview := consts.GetNextReviewAt(0)
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
