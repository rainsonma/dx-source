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

// MarkAsReview is polymorphic: pass exactly one of contentItemID or contentVocabID.
// Creates a review entry with spaced repetition schedule if it doesn't exist.
func MarkAsReview(userID string, contentItemID, contentVocabID *string, gameID, gameLevelID string) error {
	allowed, err := helpers.CheckRateLimit(
		fmt.Sprintf(rateLimitReviewKey, userID), rateLimitTracking, rateLimitTrackingSec,
	)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}
	if (contentItemID == nil) == (contentVocabID == nil) {
		return fmt.Errorf("must specify exactly one of contentItemID / contentVocabID")
	}

	nextReview := consts.GetNextReviewAt(0)
	nextReviewCarbon := carbon.FromStdTime(nextReview)

	if contentItemID != nil {
		if _, err := facades.Orm().Query().Exec(
			`INSERT INTO user_reviews
			   (id, user_id, content_item_id, game_id, game_level_id, next_review_at, review_count, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, 0, now(), now())
			 ON CONFLICT (user_id, content_item_id) WHERE content_item_id IS NOT NULL AND deleted_at IS NULL DO NOTHING`,
			newID(), userID, *contentItemID, gameID, gameLevelID, nextReviewCarbon,
		); err != nil {
			return fmt.Errorf("failed to mark for review (item): %w", err)
		}
	} else {
		if _, err := facades.Orm().Query().Exec(
			`INSERT INTO user_reviews
			   (id, user_id, content_vocab_id, game_id, game_level_id, next_review_at, review_count, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, 0, now(), now())
			 ON CONFLICT (user_id, content_vocab_id) WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL DO NOTHING`,
			newID(), userID, *contentVocabID, gameID, gameLevelID, nextReviewCarbon,
		); err != nil {
			return fmt.Errorf("failed to mark for review (vocab): %w", err)
		}
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

	items := make([]ReviewItemData, 0, len(reviews))
	if len(reviews) == 0 {
		return items, nextCursor, hasMore, nil
	}

	// Collect IDs for batch loading
	itemIDs := []string{}
	vocabIDs := []string{}
	gameIDs := []string{}
	for _, r := range reviews {
		if r.ContentItemID != nil {
			itemIDs = append(itemIDs, *r.ContentItemID)
		}
		if r.ContentVocabID != nil {
			vocabIDs = append(vocabIDs, *r.ContentVocabID)
		}
		if r.GameID != "" {
			gameIDs = append(gameIDs, r.GameID)
		}
	}

	contentMap := batchLoadContentItems(itemIDs)
	vocabMap := batchLoadContentVocabs(vocabIDs)
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
		if r.ContentItemID != nil {
			item.ContentItem = contentMap[*r.ContentItemID]
		} else if r.ContentVocabID != nil {
			item.ContentItem = vocabMap[*r.ContentVocabID]
		}
		item.GameName = gameMapPtr(gameMap, r.GameID)
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

// DeleteReview soft-deletes a single review entry owned by the user.
func DeleteReview(userID, id string) error {
	_, err := facades.Orm().Query().Exec(
		`UPDATE user_reviews SET deleted_at = NOW()
		   WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	)
	return err
}

// BulkDeleteReviews soft-deletes multiple review entries owned by the user.
func BulkDeleteReviews(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := facades.Orm().Query().Exec(
		`UPDATE user_reviews SET deleted_at = NOW()
		   WHERE user_id = ? AND id IN ? AND deleted_at IS NULL`,
		userID, ids,
	)
	return err
}
