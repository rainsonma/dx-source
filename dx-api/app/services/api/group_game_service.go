package api

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// GroupGameSearchItem represents a game in group game search results.
type GroupGameSearchItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Mode         string `json:"mode"`
	CategoryName string `json:"category_name"`
}

// SearchGamesForGroup searches published active games by name for group game selection.
func SearchGamesForGroup(query string, limit int) ([]GroupGameSearchItem, error) {
	if limit <= 0 {
		limit = 20
	}

	q := facades.Orm().Query().
		Where("status", consts.GameStatusPublished).
		Where("is_active", true).
		Order("created_at DESC").
		Limit(limit)

	if query != "" {
		q = q.Where("name ILIKE ?", "%"+query+"%")
	}

	var games []models.Game
	if err := q.Get(&games); err != nil {
		return nil, fmt.Errorf("failed to search games for group: %w", err)
	}

	// Batch load categories
	categoryIDs := make([]string, 0, len(games))
	for _, g := range games {
		if g.GameCategoryID != nil && *g.GameCategoryID != "" {
			categoryIDs = append(categoryIDs, *g.GameCategoryID)
		}
	}
	categoryMap := make(map[string]string)
	if len(categoryIDs) > 0 {
		var categories []models.GameCategory
		if err := facades.Orm().Query().Where("id IN ?", categoryIDs).Get(&categories); err == nil {
			for _, cat := range categories {
				categoryMap[cat.ID] = cat.Name
			}
		}
	}

	result := make([]GroupGameSearchItem, 0, len(games))
	for _, g := range games {
		item := GroupGameSearchItem{
			ID:   g.ID,
			Name: g.Name,
			Mode: g.Mode,
		}
		if g.GameCategoryID != nil {
			if name, ok := categoryMap[*g.GameCategoryID]; ok {
				item.CategoryName = name
			}
		}
		result = append(result, item)
	}

	return result, nil
}

// SetGroupGame sets the current game and game mode for a group.
func SetGroupGame(userID, groupID, gameID, gameMode string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if group.IsPlaying {
		return ErrGroupIsPlaying
	}

	var game models.Game
	if err := facades.Orm().Query().Where("id", gameID).First(&game); err != nil || game.ID == "" {
		return ErrGameNotFound
	}
	if game.Status != consts.GameStatusPublished {
		return ErrGameNotPublished
	}

	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update(map[string]any{
		"current_game_id": gameID,
		"game_mode":       gameMode,
	}); err != nil {
		return fmt.Errorf("failed to set group game: %w", err)
	}
	return nil
}

// ClearGroupGame clears the current game and game mode for a group.
func ClearGroupGame(userID, groupID string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if group.IsPlaying {
		return ErrGroupIsPlaying
	}

	if _, err := facades.Orm().Query().Exec(
		"UPDATE game_groups SET current_game_id = NULL, game_mode = NULL WHERE id = ?",
		groupID,
	); err != nil {
		return fmt.Errorf("failed to clear group game: %w", err)
	}
	return nil
}

// GroupGameStartEvent is the SSE payload for group_game_start.
type GroupGameStartEvent struct {
	GameGroupID     string  `json:"game_group_id"`
	GameID          string  `json:"game_id"`
	GameName        string  `json:"game_name"`
	GameMode        string  `json:"game_mode"`
	Degree          string  `json:"degree"`
	Pattern         *string `json:"pattern"`
	AnswerTimeLimit int     `json:"answer_time_limit"`
}

// StartGroupGame validates and initiates a group game round, broadcasting via SSE.
func StartGroupGame(userID, groupID, degree string, pattern *string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}
	if group.IsPlaying {
		return ErrGroupIsPlaying
	}
	if group.CurrentGameID == nil || *group.CurrentGameID == "" {
		return ErrNoGameSet
	}
	if group.GameMode == nil || *group.GameMode == "" {
		return ErrNoGameModeSet
	}

	// Fetch game name for SSE payload
	var game models.Game
	if err := facades.Orm().Query().Where("id", *group.CurrentGameID).First(&game); err != nil || game.ID == "" {
		return ErrGameNotFound
	}

	// Set is_playing = true
	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).
		Update("is_playing", true); err != nil {
		return fmt.Errorf("failed to set is_playing: %w", err)
	}

	// Broadcast SSE event
	helpers.GroupSSEHub.Broadcast(groupID, "group_game_start", GroupGameStartEvent{
		GameGroupID:     groupID,
		GameID:          *group.CurrentGameID,
		GameName:        game.Name,
		GameMode:        *group.GameMode,
		Degree:          degree,
		Pattern:         pattern,
		AnswerTimeLimit: group.AnswerTimeLimit,
	})

	return nil
}
