package api

import "github.com/goravel/framework/contracts/http"

// ToggleFavoriteRequest validates favorite toggle data.
type ToggleFavoriteRequest struct {
	GameID string `form:"game_id" json:"game_id"`
}

func (r *ToggleFavoriteRequest) Authorize(ctx http.Context) error { return nil }
func (r *ToggleFavoriteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{"game_id": "required"}
}
