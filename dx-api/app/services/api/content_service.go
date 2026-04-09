package api

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// ContentItemData represents a content item returned to the client.
type ContentItemData struct {
	ID          string  `json:"id"`
	Content     string  `json:"content"`
	ContentType string  `json:"contentType"`
	Translation *string `json:"translation"`
	Definition  *string `json:"definition"`
	Explanation *string `json:"explanation"`
	Items       *string `json:"items"`
	Structure   *string `json:"structure"`
	UkAudioURL  *string `json:"ukAudioUrl"`
	UsAudioURL  *string `json:"usAudioUrl"`
}

// GetLevelContent returns content items for a game level, filtered by degree.
func GetLevelContent(userID, gameLevelID string, degree string) ([]ContentItemData, error) {
	// VIP guard: non-first levels require active VIP
	var level models.GameLevel
	if err := facades.Orm().Query().Select("id", "game_id").Where("id", gameLevelID).First(&level); err != nil || level.ID == "" {
		return nil, ErrLevelNotFound
	}
	if err := requireVipForLevel(userID, level.GameID, gameLevelID); err != nil {
		return nil, err
	}

	// Determine allowed content types from degree
	allowedTypes, hasDegree := consts.DegreeContentTypes[degree]

	query := facades.Orm().Query().
		Where("content_items.game_level_id", gameLevelID).
		Where("content_items.is_active", true)

	// If degree is defined and has specific content type restrictions, filter by them
	if hasDegree && allowedTypes != nil {
		query = query.Where("content_items.content_type IN ?", allowedTypes)
	}

	var items []models.ContentItem
	if err := query.Order("content_items.\"order\" ASC").Get(&items); err != nil {
		return nil, fmt.Errorf("failed to get level content: %w", err)
	}

	// Collect audio IDs for batch lookup
	audioIDs := make([]string, 0, len(items)*2)
	for _, item := range items {
		if item.UkAudioID != nil && *item.UkAudioID != "" {
			audioIDs = append(audioIDs, *item.UkAudioID)
		}
		if item.UsAudioID != nil && *item.UsAudioID != "" {
			audioIDs = append(audioIDs, *item.UsAudioID)
		}
	}

	// Batch load audio URLs
	audioMap := make(map[string]string)
	if len(audioIDs) > 0 {
		var audios []models.Image
		if err := facades.Orm().Query().Where("id IN ?", audioIDs).Get(&audios); err == nil {
			for _, a := range audios {
				audioMap[a.ID] = a.Url
			}
		}
	}

	result := make([]ContentItemData, 0, len(items))
	for _, item := range items {
		data := ContentItemData{
			ID:          item.ID,
			Content:     item.Content,
			ContentType: item.ContentType,
			Translation: item.Translation,
			Definition:  item.Definition,
			Explanation: item.Explanation,
			Items:       item.Items,
			Structure:   item.Structure,
		}

		if item.UkAudioID != nil {
			if url, ok := audioMap[*item.UkAudioID]; ok {
				data.UkAudioURL = &url
			}
		}
		if item.UsAudioID != nil {
			if url, ok := audioMap[*item.UsAudioID]; ok {
				data.UsAudioURL = &url
			}
		}

		result = append(result, data)
	}

	return result, nil
}
