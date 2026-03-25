package api

import (
	"github.com/goravel/framework/contracts/http"
)

// ---------- SetGroupGameRequest ----------

type SetGroupGameRequest struct {
	GameID   string `form:"game_id" json:"game_id"`
	GameMode string `form:"game_mode" json:"game_mode"`
}

func (r *SetGroupGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *SetGroupGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":   "required",
		"game_mode": "required|in:solo,team",
	}
}
func (r *SetGroupGameRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{}
}
func (r *SetGroupGameRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":   "请指定游戏",
		"game_mode.required": "请指定游戏模式",
		"game_mode.in":       "游戏模式只能为solo或team",
	}
}
