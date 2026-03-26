package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

// ---------- GroupPlayStartSessionRequest ----------

type GroupPlayStartSessionRequest struct {
	GameID      string  `form:"game_id" json:"game_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
	LevelID     *string `form:"level_id" json:"level_id"`
	GameGroupID string  `form:"game_group_id" json:"game_group_id"`
}

func (r *GroupPlayStartSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *GroupPlayStartSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
		"level_id":      "uuid",
		"game_group_id": "required|uuid",
	}
}
func (r *GroupPlayStartSessionRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"degree":  "trim",
		"pattern": "trim",
	}
}
func (r *GroupPlayStartSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":       "请选择游戏",
		"game_id.uuid":           "无效的游戏ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
		"level_id.uuid":          "无效的关卡ID",
		"game_group_id.required": "请指定群组",
		"game_group_id.uuid":     "无效的群组ID",
	}
}
func (r *GroupPlayStartSessionRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}
