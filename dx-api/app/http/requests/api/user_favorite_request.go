package api

// ToggleFavoriteRequest validates favorite toggle data.
type ToggleFavoriteRequest struct {
	GameID string `form:"game_id" json:"game_id"`
}
