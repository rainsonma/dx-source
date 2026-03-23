package api

import "github.com/goravel/framework/contracts/http"

type ToggleFavoriteRequest struct {
	GameID string `form:"game_id" json:"game_id"`
}

func (r *ToggleFavoriteRequest) Authorize(ctx http.Context) error { return nil }
func (r *ToggleFavoriteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id": "required|uuid",
	}
}
func (r *ToggleFavoriteRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required": "请选择游戏",
		"game_id.uuid":     "无效的游戏ID",
	}
}
