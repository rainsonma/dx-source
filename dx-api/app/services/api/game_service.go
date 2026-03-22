package api

import (
	"fmt"

	"dx-api/app/consts"
	"github.com/goravel/framework/facades"
	"dx-api/app/models"
)

// GameCardData represents a published game in list views.
type GameCardData struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Mode         string  `json:"mode"`
	CreatedAt    any     `json:"createdAt"`
	CoverURL     *string `json:"coverUrl"`
	Author       *string `json:"author"`
	CategoryName *string `json:"categoryName"`
	LevelCount   int     `json:"levelCount"`
}

// GameDetailData represents a full game detail with levels.
type GameDetailData struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  *string           `json:"description"`
	Mode         string            `json:"mode"`
	CreatedAt    any               `json:"createdAt"`
	CoverURL     *string           `json:"coverUrl"`
	Author       *string           `json:"author"`
	CategoryName *string           `json:"categoryName"`
	PressName    *string           `json:"pressName"`
	Levels       []GameLevelData   `json:"levels"`
	LevelCount   int               `json:"levelCount"`
}

// GameLevelData represents a game level in detail views.
type GameLevelData struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Order float64 `json:"order"`
}

// GameSearchResultData represents a game in search results.
type GameSearchResultData struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Mode         string  `json:"mode"`
	CategoryName *string `json:"categoryName"`
}

// PlayedGameData represents a game the user has played.
type PlayedGameData struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Mode         string  `json:"mode"`
	CategoryName *string `json:"categoryName"`
}

// CategoryData represents a game category with hierarchy info.
type CategoryData struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Depth  int    `json:"depth"`
	IsLeaf bool   `json:"isLeaf"`
}

// PressData represents a game publisher.
type PressData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListPublishedGames returns published games with cursor pagination and optional filters.
func ListPublishedGames(cursor string, limit int, categoryIDs []string, pressID string, mode string) ([]GameCardData, string, bool, error) {
	if limit <= 0 {
		limit = 12
	}

	query := facades.Orm().Query().
		Where("status", consts.GameStatusPublished).
		Where("is_active", true)

	if len(categoryIDs) > 0 {
		query = query.Where("game_category_id IN ?", categoryIDs)
	}
	if pressID != "" {
		query = query.Where("game_press_id", pressID)
	}
	if mode != "" {
		query = query.Where("mode", mode)
	}

	if cursor != "" {
		// Cursor-based pagination: skip past the cursor game by excluding games
		// with created_at newer than or equal to the cursor, unless their id differs.
		var cursorGame models.Game
		if err := facades.Orm().Query().Where("id", cursor).First(&cursorGame); err == nil && cursorGame.ID != "" {
			query = query.Where("created_at <= ?", cursorGame.CreatedAt).
				Where("id != ?", cursor)
		}
	}

	var games []models.Game
	if err := query.Order("created_at DESC").Order("id DESC").Limit(limit + 1).Get(&games); err != nil {
		return nil, "", false, fmt.Errorf("failed to list games: %w", err)
	}

	hasMore := len(games) > limit
	if hasMore {
		games = games[:limit]
	}

	nextCursor := ""
	if hasMore && len(games) > 0 {
		nextCursor = games[len(games)-1].ID
	}

	// Collect IDs for batch lookups
	coverIDs := make([]string, 0, len(games))
	categoryIDs2 := make([]string, 0, len(games))
	userIDs := make([]string, 0, len(games))
	gameIDs := make([]string, 0, len(games))

	for _, g := range games {
		gameIDs = append(gameIDs, g.ID)
		if g.CoverID != nil && *g.CoverID != "" {
			coverIDs = append(coverIDs, *g.CoverID)
		}
		if g.GameCategoryID != nil && *g.GameCategoryID != "" {
			categoryIDs2 = append(categoryIDs2, *g.GameCategoryID)
		}
		if g.UserID != nil && *g.UserID != "" {
			userIDs = append(userIDs, *g.UserID)
		}
	}

	// Batch load cover images
	coverMap := make(map[string]string)
	if len(coverIDs) > 0 {
		var images []models.Image
		if err := facades.Orm().Query().Where("id IN ?", coverIDs).Get(&images); err == nil {
			for _, img := range images {
				coverMap[img.ID] = img.Url
			}
		}
	}

	// Batch load categories
	categoryMap := make(map[string]string)
	if len(categoryIDs2) > 0 {
		var categories []models.GameCategory
		if err := facades.Orm().Query().Where("id IN ?", categoryIDs2).Get(&categories); err == nil {
			for _, cat := range categories {
				categoryMap[cat.ID] = cat.Name
			}
		}
	}

	// Batch load users (for author nickname/username)
	authorMap := make(map[string]string)
	if len(userIDs) > 0 {
		var users []models.User
		if err := facades.Orm().Query().Where("id IN ?", userIDs).Select("id", "username").Get(&users); err == nil {
			for _, u := range users {
				authorMap[u.ID] = u.Username
			}
		}
	}

	// Count levels per game
	levelCountMap := make(map[string]int)
	if len(gameIDs) > 0 {
		var levels []models.GameLevel
		if err := facades.Orm().Query().Where("game_id IN ?", gameIDs).Where("is_active", true).Get(&levels); err == nil {
			for _, l := range levels {
				levelCountMap[l.GameID]++
			}
		}
	}

	// Build result
	result := make([]GameCardData, 0, len(games))
	for _, g := range games {
		card := GameCardData{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Mode:        g.Mode,
			CreatedAt:   g.CreatedAt,
			LevelCount:  levelCountMap[g.ID],
		}

		if g.CoverID != nil {
			if url, ok := coverMap[*g.CoverID]; ok {
				card.CoverURL = &url
			}
		}
		if g.GameCategoryID != nil {
			if name, ok := categoryMap[*g.GameCategoryID]; ok {
				card.CategoryName = &name
			}
		}
		if g.UserID != nil {
			if name, ok := authorMap[*g.UserID]; ok {
				card.Author = &name
			}
		}

		result = append(result, card)
	}

	return result, nextCursor, hasMore, nil
}

// SearchGames searches published games by name (case-insensitive).
func SearchGames(queryStr string, limit int) ([]GameSearchResultData, error) {
	if limit <= 0 {
		limit = 10
	}

	var games []models.Game
	if err := facades.Orm().Query().
		Where("status", consts.GameStatusPublished).
		Where("is_active", true).
		Where("name ILIKE ?", "%"+queryStr+"%").
		Order("created_at DESC").
		Limit(limit).
		Get(&games); err != nil {
		return nil, fmt.Errorf("failed to search games: %w", err)
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

	result := make([]GameSearchResultData, 0, len(games))
	for _, g := range games {
		item := GameSearchResultData{
			ID:   g.ID,
			Name: g.Name,
			Mode: g.Mode,
		}
		if g.GameCategoryID != nil {
			if name, ok := categoryMap[*g.GameCategoryID]; ok {
				item.CategoryName = &name
			}
		}
		result = append(result, item)
	}

	return result, nil
}

// GetPlayedGames returns all games the user has played.
func GetPlayedGames(userID string) ([]PlayedGameData, error) {
	var stats []models.GameStatsTotal
	if err := facades.Orm().Query().
		Where("user_id", userID).
		Order("last_played_at DESC").
		Limit(10).
		Get(&stats); err != nil {
		return nil, fmt.Errorf("failed to get recent games: %w", err)
	}

	if len(stats) == 0 {
		return []PlayedGameData{}, nil
	}

	// Collect game IDs
	gameIDs := make([]string, 0, len(stats))
	for _, s := range stats {
		gameIDs = append(gameIDs, s.GameID)
	}

	// Load games that are published and active
	var games []models.Game
	if err := facades.Orm().Query().
		Where("id IN ?", gameIDs).
		Where("status", consts.GameStatusPublished).
		Where("is_active", true).
		Get(&games); err != nil {
		return nil, fmt.Errorf("failed to load games: %w", err)
	}

	gameMap := make(map[string]models.Game)
	categoryIDs := make([]string, 0, len(games))
	for _, g := range games {
		gameMap[g.ID] = g
		if g.GameCategoryID != nil && *g.GameCategoryID != "" {
			categoryIDs = append(categoryIDs, *g.GameCategoryID)
		}
	}

	// Batch load categories
	categoryMap := make(map[string]string)
	if len(categoryIDs) > 0 {
		var categories []models.GameCategory
		if err := facades.Orm().Query().Where("id IN ?", categoryIDs).Get(&categories); err == nil {
			for _, cat := range categories {
				categoryMap[cat.ID] = cat.Name
			}
		}
	}

	// Build result preserving order from stats (last_played_at DESC)
	result := make([]PlayedGameData, 0, len(stats))
	for _, s := range stats {
		g, ok := gameMap[s.GameID]
		if !ok {
			continue
		}
		item := PlayedGameData{
			ID:   g.ID,
			Name: g.Name,
			Mode: g.Mode,
		}
		if g.GameCategoryID != nil {
			if name, ok := categoryMap[*g.GameCategoryID]; ok {
				item.CategoryName = &name
			}
		}
		result = append(result, item)
	}

	return result, nil
}

// GetGameDetail returns full game detail with levels.
func GetGameDetail(gameID string) (*GameDetailData, error) {
	var game models.Game
	if err := facades.Orm().Query().
		Where("id", gameID).
		Where("status", consts.GameStatusPublished).
		Where("is_active", true).
		First(&game); err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}
	if game.ID == "" {
		return nil, ErrGameNotFound
	}

	// Load active levels
	var levels []models.GameLevel
	if err := facades.Orm().Query().
		Where("game_id", gameID).
		Where("is_active", true).
		Order("\"order\" ASC").
		Get(&levels); err != nil {
		return nil, fmt.Errorf("failed to load levels: %w", err)
	}

	// Load cover image
	var coverURL *string
	if game.CoverID != nil && *game.CoverID != "" {
		var image models.Image
		if err := facades.Orm().Query().Where("id", *game.CoverID).First(&image); err == nil && image.ID != "" {
			coverURL = &image.Url
		}
	}

	// Load category name
	var categoryName *string
	if game.GameCategoryID != nil && *game.GameCategoryID != "" {
		var cat models.GameCategory
		if err := facades.Orm().Query().Where("id", *game.GameCategoryID).First(&cat); err == nil && cat.ID != "" {
			categoryName = &cat.Name
		}
	}

	// Load press name
	var pressName *string
	if game.GamePressID != nil && *game.GamePressID != "" {
		var press models.GamePress
		if err := facades.Orm().Query().Where("id", *game.GamePressID).First(&press); err == nil && press.ID != "" {
			pressName = &press.Name
		}
	}

	// Load author
	var author *string
	if game.UserID != nil && *game.UserID != "" {
		var user models.User
		if err := facades.Orm().Query().Where("id", *game.UserID).Select("id", "username").First(&user); err == nil && user.ID != "" {
			author = &user.Username
		}
	}

	// Build level data
	levelData := make([]GameLevelData, 0, len(levels))
	for _, l := range levels {
		levelData = append(levelData, GameLevelData{
			ID:    l.ID,
			Name:  l.Name,
			Order: l.Order,
		})
	}

	detail := &GameDetailData{
		ID:           game.ID,
		Name:         game.Name,
		Description:  game.Description,
		Mode:         game.Mode,
		CreatedAt:    game.CreatedAt,
		CoverURL:     coverURL,
		Author:       author,
		CategoryName: categoryName,
		PressName:    pressName,
		Levels:       levelData,
		LevelCount:   len(levels),
	}

	return detail, nil
}

// ListCategories returns all enabled categories in hierarchical order.
func ListCategories() ([]CategoryData, error) {
	var categories []models.GameCategory
	if err := facades.Orm().Query().
		Where("is_enabled", true).
		Order("\"order\" ASC").
		Get(&categories); err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// Build parent-children map
	parentMap := make(map[string][]models.GameCategory)
	rootKey := ""
	for _, cat := range categories {
		key := rootKey
		if cat.ParentID != nil {
			key = *cat.ParentID
		}
		parentMap[key] = append(parentMap[key], cat)
	}

	// Track which IDs have children
	hasChildren := make(map[string]bool)
	for _, cat := range categories {
		if cat.ParentID != nil {
			hasChildren[*cat.ParentID] = true
		}
	}

	// Walk the tree depth-first to produce a flat list
	var result []CategoryData
	var walk func(parentID string, depth int)
	walk = func(parentID string, depth int) {
		children := parentMap[parentID]
		for _, cat := range children {
			isLeaf := !hasChildren[cat.ID]
			result = append(result, CategoryData{
				ID:     cat.ID,
				Name:   cat.Name,
				Depth:  depth,
				IsLeaf: isLeaf,
			})
			walk(cat.ID, depth+1)
		}
	}
	walk(rootKey, 0)

	return result, nil
}

// ListPresses returns all game publishers ordered by their display order.
func ListPresses() ([]PressData, error) {
	var presses []models.GamePress
	if err := facades.Orm().Query().
		Order("\"order\" ASC").
		Get(&presses); err != nil {
		return nil, fmt.Errorf("failed to list presses: %w", err)
	}

	result := make([]PressData, 0, len(presses))
	for _, p := range presses {
		result = append(result, PressData{
			ID:   p.ID,
			Name: p.Name,
		})
	}

	return result, nil
}
