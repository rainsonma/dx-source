package api

import (
	"crypto/rand"
	"fmt"
	"time"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/oklog/ulid/v2"
)

// Content limits for course games.
const (
	MaxSentences    = 20
	MaxVocab        = 200
	MaxItemsPerMeta = 50
)

// Source types for content metadata.
const (
	SourceTypeSentence = "sentence"
	SourceTypeVocab    = "vocab"
)

// CourseGameCardData represents a user's own game in list views.
type CourseGameCardData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Mode        string  `json:"mode"`
	Status      string  `json:"status"`
	CoverURL    *string `json:"coverUrl"`
	LevelCount  int     `json:"levelCount"`
	CreatedAt   any     `json:"createdAt"`
}

// CourseGameDetailData represents a course game with levels for editing.
type CourseGameDetailData struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Description    *string               `json:"description"`
	Mode           string                `json:"mode"`
	Status         string                `json:"status"`
	GameCategoryID *string               `json:"gameCategoryId"`
	GamePressID    *string               `json:"gamePressId"`
	CoverID        *string               `json:"coverId"`
	CoverURL       *string               `json:"coverUrl"`
	Levels         []CourseGameLevelData `json:"levels"`
	CreatedAt      any                   `json:"createdAt"`
}

// CourseGameLevelData represents a level in a course game.
type CourseGameLevelData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Order       float64 `json:"order"`
}

// ListUserGames returns the user's own games with cursor pagination and optional status filter.
func ListUserGames(userID string, status string, cursor string, limit int) ([]CourseGameCardData, string, bool, error) {
	if limit <= 0 {
		limit = 12
	}

	query := facades.Orm().Query().Where("user_id", userID)
	if status != "" {
		query = query.Where("status", status)
	}

	if cursor != "" {
		var cursorGame models.Game
		if err := facades.Orm().Query().Where("id", cursor).First(&cursorGame); err == nil && cursorGame.ID != "" {
			query = query.Where("created_at <= ?", cursorGame.CreatedAt).Where("id != ?", cursor)
		}
	}

	var games []models.Game
	if err := query.Order("created_at DESC").Order("id DESC").Limit(limit + 1).Get(&games); err != nil {
		return nil, "", false, fmt.Errorf("failed to list user games: %w", err)
	}

	hasMore := len(games) > limit
	if hasMore {
		games = games[:limit]
	}

	nextCursor := ""
	if hasMore && len(games) > 0 {
		nextCursor = games[len(games)-1].ID
	}

	// Batch load covers
	coverIDs := make([]string, 0, len(games))
	gameIDs := make([]string, 0, len(games))
	for _, g := range games {
		gameIDs = append(gameIDs, g.ID)
		if g.CoverID != nil && *g.CoverID != "" {
			coverIDs = append(coverIDs, *g.CoverID)
		}
	}

	coverMap := make(map[string]string)
	if len(coverIDs) > 0 {
		var images []models.Image
		if err := facades.Orm().Query().Where("id IN ?", coverIDs).Get(&images); err == nil {
			for _, img := range images {
				coverMap[img.ID] = img.Url
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

	result := make([]CourseGameCardData, 0, len(games))
	for _, g := range games {
		card := CourseGameCardData{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Mode:        g.Mode,
			Status:      g.Status,
			CreatedAt:   g.CreatedAt,
			LevelCount:  levelCountMap[g.ID],
		}
		if g.CoverID != nil {
			if url, ok := coverMap[*g.CoverID]; ok {
				card.CoverURL = &url
			}
		}
		result = append(result, card)
	}

	return result, nextCursor, hasMore, nil
}

// getCourseGameOwned fetches a game and verifies the user owns it.
func getCourseGameOwned(userID, gameID string) (*models.Game, error) {
	var game models.Game
	if err := facades.Orm().Query().Where("id", gameID).First(&game); err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}
	if game.ID == "" {
		return nil, ErrGameNotFound
	}
	if game.UserID == nil || *game.UserID != userID {
		return nil, ErrForbidden
	}
	return &game, nil
}

// CreateGame creates a new course game in draft status.
func CreateGame(userID, name string, description *string, mode string, categoryID, pressID, coverID *string) (string, error) {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()

	game := models.Game{
		ID:             id,
		Name:           name,
		Description:    description,
		UserID:         &userID,
		Mode:           mode,
		GameCategoryID: categoryID,
		GamePressID:    pressID,
		CoverID:        coverID,
		Order:          1000,
		IsActive:       true,
		Status:         constants.GameStatusDraft,
	}

	if err := facades.Orm().Query().Create(&game); err != nil {
		return "", fmt.Errorf("failed to create game: %w", err)
	}

	return id, nil
}

// UpdateGame updates a course game's properties. Rejects edits to published games.
func UpdateGame(userID, gameID, name string, description *string, mode string, categoryID, pressID, coverID *string) error {
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == constants.GameStatusPublished {
		return ErrGamePublished
	}

	if _, err := facades.Orm().Query().Model(&models.Game{}).Where("id", gameID).Update(map[string]any{
		"name":             name,
		"description":      description,
		"mode":             mode,
		"game_category_id": categoryID,
		"game_press_id":    pressID,
		"cover_id":         coverID,
	}); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

// DeleteGame deletes a course game and cascades to levels and content. Rejects published games.
func DeleteGame(userID, gameID string) error {
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == constants.GameStatusPublished {
		return ErrGamePublished
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Get level IDs for cascade
		var levels []models.GameLevel
		if err := tx.Where("game_id", gameID).Get(&levels); err != nil {
			return fmt.Errorf("failed to load levels: %w", err)
		}

		levelIDs := make([]string, 0, len(levels))
		for _, l := range levels {
			levelIDs = append(levelIDs, l.ID)
		}

		// Cascade delete content items and metas for all levels
		if len(levelIDs) > 0 {
			if _, err := tx.Where("game_level_id IN ?", levelIDs).Delete(&models.ContentItem{}); err != nil {
				return fmt.Errorf("failed to delete content items: %w", err)
			}
			if _, err := tx.Where("game_level_id IN ?", levelIDs).Delete(&models.ContentMeta{}); err != nil {
				return fmt.Errorf("failed to delete content metas: %w", err)
			}
		}

		// Delete levels
		if _, err := tx.Where("game_id", gameID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete levels: %w", err)
		}

		// Delete game
		if _, err := tx.Where("id", gameID).Delete(&models.Game{}); err != nil {
			return fmt.Errorf("failed to delete game: %w", err)
		}

		return nil
	})
}

// PublishGame validates readiness and sets status to published.
func PublishGame(userID, gameID string) error {
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == constants.GameStatusPublished {
		return ErrGameAlreadyPublished
	}

	// Check game has active levels
	levelCount, err2 := facades.Orm().Query().Model(&models.GameLevel{}).Where("game_id", gameID).Where("is_active", true).Count()
	if err2 != nil {
		return fmt.Errorf("failed to count levels: %w", err2)
	}
	if levelCount == 0 {
		return ErrNoGameLevels
	}

	// Check each level has content items
	var levels []models.GameLevel
	if err := facades.Orm().Query().Where("game_id", gameID).Where("is_active", true).Get(&levels); err != nil {
		return fmt.Errorf("failed to load levels: %w", err)
	}

	for _, l := range levels {
		itemCount, err3 := facades.Orm().Query().Model(&models.ContentItem{}).Where("game_level_id", l.ID).Where("is_active", true).Count()
		if err3 != nil {
			return fmt.Errorf("failed to count items: %w", err3)
		}
		if itemCount == 0 {
			return fmt.Errorf("关卡「%s」没有练习内容", l.Name)
		}
	}

	if _, err := facades.Orm().Query().Model(&models.Game{}).Where("id", gameID).Update("status", constants.GameStatusPublished); err != nil {
		return fmt.Errorf("failed to publish game: %w", err)
	}

	return nil
}

// WithdrawGame sets a published game back to withdraw status.
func WithdrawGame(userID, gameID string) error {
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status != constants.GameStatusPublished {
		return ErrGameNotPublished
	}

	if _, err := facades.Orm().Query().Model(&models.Game{}).Where("id", gameID).Update("status", constants.GameStatusWithdraw); err != nil {
		return fmt.Errorf("failed to withdraw game: %w", err)
	}

	return nil
}

// CreateLevel adds a new level to a course game with auto-incremented order.
func CreateLevel(userID, gameID, name string, description *string) (string, error) {
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return "", err
	}

	if game.Status == constants.GameStatusPublished {
		return "", ErrGamePublished
	}

	// Get max order for auto-increment
	var maxLevel models.GameLevel
	if err := facades.Orm().Query().Where("game_id", gameID).Order("\"order\" DESC").First(&maxLevel); err != nil || maxLevel.ID == "" {
		maxLevel.Order = 0
	}

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
	level := models.GameLevel{
		ID:           id,
		GameID:       gameID,
		Name:         name,
		Description:  description,
		Order:        maxLevel.Order + 1000,
		PassingScore: 60,
		IsActive:     true,
	}

	if err := facades.Orm().Query().Create(&level); err != nil {
		return "", fmt.Errorf("failed to create level: %w", err)
	}

	return id, nil
}

// DeleteLevel removes a level and its content from a course game.
func DeleteLevel(userID, gameID, levelID string) error {
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == constants.GameStatusPublished {
		return ErrGamePublished
	}

	// Verify level belongs to game
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", levelID).Where("game_id", gameID).First(&level); err != nil {
		return fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return ErrLevelNotFound
	}

	// Cascade delete content + level in transaction
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Where("game_level_id", levelID).Delete(&models.ContentItem{}); err != nil {
			return fmt.Errorf("failed to delete content items: %w", err)
		}
		if _, err := tx.Where("game_level_id", levelID).Delete(&models.ContentMeta{}); err != nil {
			return fmt.Errorf("failed to delete content metas: %w", err)
		}
		if _, err := tx.Where("id", levelID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete level: %w", err)
		}
		return nil
	})
}

// CourseGameCounts represents the count of user's games by status.
type CourseGameCounts struct {
	All       int64 `json:"all"`
	Draft     int64 `json:"draft"`
	Published int64 `json:"published"`
	Withdraw  int64 `json:"withdraw"`
}

// GetUserGameCounts returns the count of a user's games grouped by status.
func GetUserGameCounts(userID string) (*CourseGameCounts, error) {
	q := facades.Orm().Query().Model(&models.Game{}).Where("user_id", userID)

	all, err := q.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count games: %w", err)
	}

	draft, _ := facades.Orm().Query().Model(&models.Game{}).Where("user_id", userID).Where("status", constants.GameStatusDraft).Count()
	published, _ := facades.Orm().Query().Model(&models.Game{}).Where("user_id", userID).Where("status", constants.GameStatusPublished).Count()
	withdraw, _ := facades.Orm().Query().Model(&models.Game{}).Where("user_id", userID).Where("status", constants.GameStatusWithdraw).Count()

	return &CourseGameCounts{
		All:       all,
		Draft:     draft,
		Published: published,
		Withdraw:  withdraw,
	}, nil
}

// GetCourseGameDetail returns a user's game detail with levels for editing.
func GetCourseGameDetail(userID, gameID string) (*CourseGameDetailData, error) {
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return nil, err
	}

	// Load levels
	var levels []models.GameLevel
	if err := facades.Orm().Query().
		Where("game_id", gameID).
		Where("is_active", true).
		Order("\"order\" ASC").
		Get(&levels); err != nil {
		return nil, fmt.Errorf("failed to load levels: %w", err)
	}

	// Load cover URL
	var coverURL *string
	if game.CoverID != nil && *game.CoverID != "" {
		var image models.Image
		if err := facades.Orm().Query().Where("id", *game.CoverID).First(&image); err == nil && image.ID != "" {
			coverURL = &image.Url
		}
	}

	levelData := make([]CourseGameLevelData, 0, len(levels))
	for _, l := range levels {
		levelData = append(levelData, CourseGameLevelData{
			ID:          l.ID,
			Name:        l.Name,
			Description: l.Description,
			Order:       l.Order,
		})
	}

	return &CourseGameDetailData{
		ID:             game.ID,
		Name:           game.Name,
		Description:    game.Description,
		Mode:           game.Mode,
		Status:         game.Status,
		GameCategoryID: game.GameCategoryID,
		GamePressID:    game.GamePressID,
		CoverID:        game.CoverID,
		CoverURL:       coverURL,
		Levels:         levelData,
		CreatedAt:      game.CreatedAt,
	}, nil
}
