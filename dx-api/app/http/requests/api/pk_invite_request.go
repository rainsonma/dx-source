package api

import (
	"dx-api/app/helpers"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type PkInviteRequest struct {
	GameID      string  `form:"game_id" json:"game_id"`
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
	OpponentID  string  `form:"opponent_id" json:"opponent_id"`
}

func (r *PkInviteRequest) Authorize(ctx http.Context) error { return nil }

func (r *PkInviteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"game_level_id": "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
		"opponent_id":   "required|uuid",
	}
}

func (r *PkInviteRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"degree":  "trim",
		"pattern": "trim",
	}
}

func (r *PkInviteRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":       "请选择游戏",
		"game_id.uuid":           "无效的游戏ID",
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
		"opponent_id.required":   "请指定对手",
		"opponent_id.uuid":       "无效的对手ID",
	}
}

func (r *PkInviteRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", "intermediate")
	}
	return nil
}
