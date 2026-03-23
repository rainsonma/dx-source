package api

import "github.com/goravel/framework/contracts/http"

// MarkMasteredRequest validates mark mastered data.
type MarkMasteredRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}

func (r *MarkMasteredRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkMasteredRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "required",
		"game_id":         "required",
		"game_level_id":   "required",
	}
}

// BulkDeleteRequest validates bulk delete data.
type BulkDeleteRequest struct {
	IDs []string `form:"ids" json:"ids"`
}

func (r *BulkDeleteRequest) Authorize(ctx http.Context) error { return nil }
func (r *BulkDeleteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{"ids": "required|min_len:1"}
}
