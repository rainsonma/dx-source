package api

import (
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
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
