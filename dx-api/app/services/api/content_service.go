package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// ContentItemData represents a content item returned to the client.
// Shape preserved exactly for dx-mini compatibility.
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

// GetLevelContent returns content for a game level, mode-branched.
// Word-sentence: reads content_items, filtered by degree's allowed content_type set.
// Vocab modes: reads content_vocabs via game_vocabs, synthesizes ContentItemData envelopes.
func GetLevelContent(userID, gameLevelID string, degree string) ([]ContentItemData, error) {
	// VIP guard: non-first levels require active VIP
	var level models.GameLevel
	if err := facades.Orm().Query().Select("id", "game_id").Where("id", gameLevelID).First(&level); err != nil || level.ID == "" {
		return nil, ErrLevelNotFound
	}
	if err := requireVipForLevel(userID, level.GameID, gameLevelID); err != nil {
		return nil, err
	}

	// Fetch the game to determine mode
	var game models.Game
	if err := facades.Orm().Query().Select("id", "mode").Where("id", level.GameID).First(&game); err != nil || game.ID == "" {
		return nil, ErrGameNotFound
	}

	if consts.IsVocabMode(game.Mode) {
		return getLevelVocabContent(gameLevelID)
	}
	return getLevelItemContent(gameLevelID, degree)
}

func getLevelItemContent(gameLevelID, degree string) ([]ContentItemData, error) {
	allowedTypes, hasDegree := consts.DegreeContentTypes[degree]

	query := facades.Orm().Query().Model(&models.ContentItem{}).
		Where("game_level_id", gameLevelID)

	if hasDegree && allowedTypes != nil {
		query = query.Where("content_type IN ?", allowedTypes)
	}

	var items []models.ContentItem
	if err := query.Order(`"order" ASC`).Get(&items); err != nil {
		return nil, fmt.Errorf("failed to get level content: %w", err)
	}

	result := make([]ContentItemData, 0, len(items))
	for _, item := range items {
		result = append(result, ContentItemData{
			ID:          item.ID,
			Content:     item.Content,
			ContentType: item.ContentType,
			Translation: item.Translation,
			Definition:  item.Definition,
			Explanation: item.Explanation,
			Items:       item.Items,
			Structure:   item.Structure,
			UkAudioURL:  item.UkAudioURL,
			UsAudioURL:  item.UsAudioURL,
		})
	}
	return result, nil
}

// getLevelVocabContent loads game_vocabs joined with content_vocabs and
// synthesizes ContentItemData envelopes so dx-mini sees the same shape.
func getLevelVocabContent(gameLevelID string) ([]ContentItemData, error) {
	type joinedRow struct {
		GvID        string  `gorm:"column:gv_id"`
		Order       float64 `gorm:"column:gv_order"`
		Content     string  `gorm:"column:content"`
		Definition  *string `gorm:"column:definition"`
		Explanation *string `gorm:"column:explanation"`
		UkPhonetic  *string `gorm:"column:uk_phonetic"`
		UsPhonetic  *string `gorm:"column:us_phonetic"`
		UkAudioURL  *string `gorm:"column:uk_audio_url"`
		UsAudioURL  *string `gorm:"column:us_audio_url"`
	}
	var rows []joinedRow
	if err := facades.Orm().Query().Raw(
		`SELECT gv.id AS gv_id, gv."order" AS gv_order,
		        cv.content, cv.definition, cv.explanation,
		        cv.uk_phonetic, cv.us_phonetic, cv.uk_audio_url, cv.us_audio_url
		   FROM game_vocabs gv
		   JOIN content_vocabs cv
		     ON cv.id = gv.content_vocab_id AND cv.deleted_at IS NULL
		  WHERE gv.game_level_id = ? AND gv.deleted_at IS NULL
		  ORDER BY gv."order" ASC`,
		gameLevelID,
	).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to load level vocabs: %w", err)
	}

	result := make([]ContentItemData, 0, len(rows))
	for _, r := range rows {
		// Synthesize: definition becomes joined Chinese gloss; items has one element.
		joinedDef := joinDefinitionGloss(r.Definition)
		itemsJSON := buildSyntheticVocabItems(r.Content, r.UkPhonetic, r.UsPhonetic, r.Definition)

		row := ContentItemData{
			ID:          r.GvID,
			Content:     r.Content,
			ContentType: "vocab",
			UkAudioURL:  r.UkAudioURL,
			UsAudioURL:  r.UsAudioURL,
			Explanation: r.Explanation,
		}
		if joinedDef != "" {
			row.Definition = &joinedDef
		}
		if itemsJSON != "" {
			row.Items = &itemsJSON
		}
		result = append(result, row)
	}
	return result, nil
}

// joinDefinitionGloss flattens the [{pos: gloss}] JSON into "gloss; gloss; gloss"
// for dx-mini's display. Returns "" if input is null or unparseable.
func joinDefinitionGloss(defJSON *string) string {
	if defJSON == nil || *defJSON == "" {
		return ""
	}
	var entries []map[string]string
	if err := json.Unmarshal([]byte(*defJSON), &entries); err != nil {
		return ""
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		for _, gloss := range entry {
			parts = append(parts, gloss)
		}
	}
	return strings.Join(parts, "; ")
}

// buildSyntheticVocabItems produces the items JSON dx-mini's MCQ builder expects.
func buildSyntheticVocabItems(content string, ukP, usP, defJSON *string) string {
	uk := ""
	us := ""
	if ukP != nil {
		uk = *ukP
	}
	if usP != nil {
		us = *usP
	}
	posLabel := ""
	defLabel := ""
	if defJSON != nil && *defJSON != "" {
		var entries []map[string]string
		if err := json.Unmarshal([]byte(*defJSON), &entries); err == nil {
			posKeys := make([]string, 0, len(entries))
			defParts := make([]string, 0, len(entries))
			for _, e := range entries {
				for k, v := range e {
					posKeys = append(posKeys, k)
					defParts = append(defParts, v)
				}
			}
			posLabel = strings.Join(posKeys, "/")
			defLabel = strings.Join(defParts, "; ")
		}
	}

	item := map[string]any{
		"position":   1,
		"item":       content,
		"phonetic":   map[string]string{"uk": uk, "us": us},
		"pos":        posLabel,
		"definition": defLabel,
		"answer":     true,
	}
	out, err := json.Marshal([]map[string]any{item})
	if err != nil {
		return ""
	}
	return string(out)
}
