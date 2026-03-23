package api

import "github.com/goravel/framework/contracts/http"

// MarkReviewRequest validates mark review data.
type MarkReviewRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}

func (r *MarkReviewRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkReviewRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "required",
		"game_id":         "required",
		"game_level_id":   "required",
	}
}
