package api

import (
	"fmt"

	"github.com/goravel/framework/facades"
	"dx-api/app/models"
)

// FavoriteGameData represents a favorited game in list views.
type FavoriteGameData struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Mode         string  `json:"mode"`
	CoverURL     *string `json:"coverUrl"`
	CategoryName *string `json:"categoryName"`
	Author       *string `json:"author"`
	FavoritedAt  any     `json:"favoritedAt"`
}

// ToggleFavoriteResult indicates whether the game is now favorited.
type ToggleFavoriteResult struct {
	Favorited bool `json:"favorited"`
}

// ToggleFavorite adds or removes a game from the user's favorites.
func ToggleFavorite(userID, gameID string) (*ToggleFavoriteResult, error) {
	query := facades.Orm().Query()

	var existing models.UserFavorite
	if err := query.Where("user_id", userID).Where("game_id", gameID).
		First(&existing); err == nil && existing.ID != "" {
		// Remove favorite
		if _, err := query.Exec(
			"DELETE FROM user_favorites WHERE id = ?", existing.ID,
		); err != nil {
			return nil, fmt.Errorf("failed to remove favorite: %w", err)
		}
		return &ToggleFavoriteResult{Favorited: false}, nil
	}

	// Add favorite
	fav := models.UserFavorite{
		ID:     newID(),
		UserID: userID,
		GameID: gameID,
	}
	if err := query.Create(&fav); err != nil {
		return nil, fmt.Errorf("failed to add favorite: %w", err)
	}
	return &ToggleFavoriteResult{Favorited: true}, nil
}

// ListFavorites returns the user's favorited games with details.
func ListFavorites(userID string) ([]FavoriteGameData, error) {
	query := facades.Orm().Query()

	var favorites []models.UserFavorite
	if err := query.Where("user_id", userID).Order("created_at desc").Get(&favorites); err != nil {
		return nil, fmt.Errorf("failed to list favorites: %w", err)
	}

	if len(favorites) == 0 {
		return []FavoriteGameData{}, nil
	}

	// Batch load games
	gameIDs := make([]string, 0, len(favorites))
	for _, f := range favorites {
		gameIDs = append(gameIDs, f.GameID)
	}

	var games []models.Game
	facades.Orm().Query().Where("id IN ?", gameIDs).Get(&games)

	gameMap := make(map[string]models.Game, len(games))
	for _, g := range games {
		gameMap[g.ID] = g
	}

	// Batch load covers
	coverIDs := make([]string, 0)
	for _, g := range games {
		if g.CoverID != nil && *g.CoverID != "" {
			coverIDs = append(coverIDs, *g.CoverID)
		}
	}
	coverMap := make(map[string]string)
	if len(coverIDs) > 0 {
		var images []models.Image
		facades.Orm().Query().Where("id IN ?", coverIDs).Get(&images)
		for _, img := range images {
			coverMap[img.ID] = img.Url
		}
	}

	// Batch load categories
	catIDs := make([]string, 0)
	for _, g := range games {
		if g.GameCategoryID != nil && *g.GameCategoryID != "" {
			catIDs = append(catIDs, *g.GameCategoryID)
		}
	}
	catMap := make(map[string]string)
	if len(catIDs) > 0 {
		var cats []models.GameCategory
		facades.Orm().Query().Where("id IN ?", catIDs).Get(&cats)
		for _, c := range cats {
			catMap[c.ID] = c.Name
		}
	}

	// Batch load authors
	authorIDs := make([]string, 0)
	for _, g := range games {
		if g.UserID != nil && *g.UserID != "" {
			authorIDs = append(authorIDs, *g.UserID)
		}
	}
	authorMap := make(map[string]string)
	if len(authorIDs) > 0 {
		var users []models.User
		facades.Orm().Query().Where("id IN ?", authorIDs).Get(&users)
		for _, u := range users {
			authorMap[u.ID] = u.Username
		}
	}

	// Build results
	results := make([]FavoriteGameData, 0, len(favorites))
	for _, f := range favorites {
		g, ok := gameMap[f.GameID]
		if !ok {
			continue
		}
		item := FavoriteGameData{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Mode:        g.Mode,
			FavoritedAt: f.CreatedAt,
		}
		if g.CoverID != nil {
			if url, ok := coverMap[*g.CoverID]; ok {
				item.CoverURL = &url
			}
		}
		if g.GameCategoryID != nil {
			if name, ok := catMap[*g.GameCategoryID]; ok {
				item.CategoryName = &name
			}
		}
		if g.UserID != nil {
			if name, ok := authorMap[*g.UserID]; ok {
				item.Author = &name
			}
		}
		results = append(results, item)
	}

	return results, nil
}
