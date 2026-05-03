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

type TrackingItemData struct {
	ID          string               `json:"id"`
	ContentItem *TrackingContentData `json:"contentItem"`
	GameName    *string              `json:"gameName"`
	MasteredAt  any                  `json:"masteredAt,omitempty"`
	CreatedAt   any                  `json:"createdAt"`
}

type TrackingContentData struct {
	Content     string  `json:"content"`
	Translation *string `json:"translation"`
	ContentType string  `json:"contentType"`
}

// MarkAsMastered is now polymorphic: pass exactly one of contentItemID or contentVocabID.
func MarkAsMastered(userID string, contentItemID, contentVocabID *string, gameID, gameLevelID string) error {
	allowed, err := helpers.CheckRateLimit(
		fmt.Sprintf(rateLimitMasterKey, userID), rateLimitTracking, rateLimitTrackingSec,
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

	now := carbon.Now()
	id := newID()

	// Plain insert; partial uniques (with `WHERE deleted_at IS NULL`) keep us
	// idempotent against double-clicks. Conflicts on live rows are silent.
	if contentItemID != nil {
		if _, err := facades.Orm().Query().Exec(
			`INSERT INTO user_masters
			   (id, user_id, content_item_id, game_id, game_level_id, mastered_at, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, now(), now())
			 ON CONFLICT (user_id, content_item_id) WHERE content_item_id IS NOT NULL AND deleted_at IS NULL DO NOTHING`,
			id, userID, *contentItemID, gameID, gameLevelID, now,
		); err != nil {
			return fmt.Errorf("failed to mark as mastered (item): %w", err)
		}
		// Soft-delete from unknown if exists
		_, _ = facades.Orm().Query().Exec(
			`UPDATE user_unknowns SET deleted_at = NOW()
			   WHERE user_id = ? AND content_item_id = ? AND deleted_at IS NULL`,
			userID, *contentItemID,
		)
	} else {
		if _, err := facades.Orm().Query().Exec(
			`INSERT INTO user_masters
			   (id, user_id, content_vocab_id, game_id, game_level_id, mastered_at, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, now(), now())
			 ON CONFLICT (user_id, content_vocab_id) WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL DO NOTHING`,
			id, userID, *contentVocabID, gameID, gameLevelID, now,
		); err != nil {
			return fmt.Errorf("failed to mark as mastered (vocab): %w", err)
		}
		_, _ = facades.Orm().Query().Exec(
			`UPDATE user_unknowns SET deleted_at = NOW()
			   WHERE user_id = ? AND content_vocab_id = ? AND deleted_at IS NULL`,
			userID, *contentVocabID,
		)
	}

	return nil
}

// ListMastered returns paginated mastered items with content details (mixed item+vocab).
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

	return enrichTrackingItems(masters, nil, nil), nextCursor, hasMore, nil
}

// enrichTrackingItems assembles {content, translation, contentType} for each row,
// loading from content_items or content_vocabs based on which FK is non-nil.
func enrichTrackingItems(masters []models.UserMaster, unknowns []models.UserUnknown, reviews []models.UserReview) []TrackingItemData {
	itemIDs := []string{}
	vocabIDs := []string{}
	gameIDs := []string{}

	collect := func(itemID, vocabID *string, gameID string) {
		if itemID != nil {
			itemIDs = append(itemIDs, *itemID)
		}
		if vocabID != nil {
			vocabIDs = append(vocabIDs, *vocabID)
		}
		if gameID != "" {
			gameIDs = append(gameIDs, gameID)
		}
	}

	for _, m := range masters {
		collect(m.ContentItemID, m.ContentVocabID, m.GameID)
	}
	for _, u := range unknowns {
		collect(u.ContentItemID, u.ContentVocabID, u.GameID)
	}
	for _, r := range reviews {
		collect(r.ContentItemID, r.ContentVocabID, r.GameID)
	}

	itemMap := batchLoadContentItems(itemIDs)
	vocabMap := batchLoadContentVocabs(vocabIDs)
	gameMap := batchLoadGameNames(gameIDs)

	resolve := func(itemID, vocabID *string) *TrackingContentData {
		if itemID != nil {
			return itemMap[*itemID]
		}
		if vocabID != nil {
			return vocabMap[*vocabID]
		}
		return nil
	}

	out := make([]TrackingItemData, 0, len(masters)+len(unknowns)+len(reviews))
	for _, m := range masters {
		out = append(out, TrackingItemData{
			ID:          m.ID,
			ContentItem: resolve(m.ContentItemID, m.ContentVocabID),
			GameName:    gameMapPtr(gameMap, m.GameID),
			MasteredAt:  m.MasteredAt,
			CreatedAt:   m.CreatedAt,
		})
	}
	for _, u := range unknowns {
		out = append(out, TrackingItemData{
			ID:          u.ID,
			ContentItem: resolve(u.ContentItemID, u.ContentVocabID),
			GameName:    gameMapPtr(gameMap, u.GameID),
			CreatedAt:   u.CreatedAt,
		})
	}
	for _, r := range reviews {
		out = append(out, TrackingItemData{
			ID:          r.ID,
			ContentItem: resolve(r.ContentItemID, r.ContentVocabID),
			GameName:    gameMapPtr(gameMap, r.GameID),
			CreatedAt:   r.CreatedAt,
		})
	}
	return out
}

func gameMapPtr(m map[string]string, id string) *string {
	if v, ok := m[id]; ok {
		return &v
	}
	return nil
}

func batchLoadContentItems(ids []string) map[string]*TrackingContentData {
	result := make(map[string]*TrackingContentData)
	if len(ids) == 0 {
		return result
	}
	var items []models.ContentItem
	facades.Orm().Query().WithTrashed().Where("id IN ?", ids).Get(&items)
	for _, ci := range items {
		result[ci.ID] = &TrackingContentData{
			Content:     ci.Content,
			Translation: ci.Translation,
			ContentType: ci.ContentType,
		}
	}
	return result
}

func batchLoadContentVocabs(ids []string) map[string]*TrackingContentData {
	result := make(map[string]*TrackingContentData)
	if len(ids) == 0 {
		return result
	}
	var vocabs []models.ContentVocab
	facades.Orm().Query().WithTrashed().Where("id IN ?", ids).Get(&vocabs)
	for _, cv := range vocabs {
		// translation is the joined gloss from definition (or nil if empty)
		joined := joinDefinitionGloss(cv.Definition)
		var translation *string
		if joined != "" {
			translation = &joined
		}
		result[cv.ID] = &TrackingContentData{
			Content:     cv.Content,
			Translation: translation,
			ContentType: "vocab",
		}
	}
	return result
}

func batchLoadGameNames(ids []string) map[string]string {
	result := make(map[string]string)
	if len(ids) == 0 {
		return result
	}
	var games []models.Game
	facades.Orm().Query().WithTrashed().Where("id IN ?", ids).Get(&games)
	for _, g := range games {
		result[g.ID] = g.Name
	}
	return result
}

// MasterStatsData / GetMasterStats / DeleteMastered / BulkDeleteMastered: same shape.

type MasterStatsData struct {
	Total     int64 `json:"total"`
	ThisWeek  int64 `json:"thisWeek"`
	ThisMonth int64 `json:"thisMonth"`
}

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

// DeleteMastered soft-deletes a single mastered entry owned by the user.
func DeleteMastered(userID, id string) error {
	_, err := facades.Orm().Query().Exec(
		`UPDATE user_masters SET deleted_at = NOW()
		   WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	)
	return err
}

// BulkDeleteMastered soft-deletes multiple mastered entries owned by the user.
func BulkDeleteMastered(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := facades.Orm().Query().Exec(
		`UPDATE user_masters SET deleted_at = NOW()
		   WHERE user_id = ? AND id IN ? AND deleted_at IS NULL`,
		userID, ids,
	)
	return err
}
