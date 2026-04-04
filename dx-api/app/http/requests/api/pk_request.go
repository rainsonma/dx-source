package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

// ---------- PkStartRequest ----------

type PkStartRequest struct {
	GameID     string  `form:"game_id" json:"game_id"`
	Degree     string  `form:"degree" json:"degree"`
	Pattern    *string `form:"pattern" json:"pattern"`
	LevelID    *string `form:"level_id" json:"level_id"`
	Difficulty string  `form:"difficulty" json:"difficulty"`
}

func (r *PkStartRequest) Authorize(ctx http.Context) error { return nil }
func (r *PkStartRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":    "required|uuid",
		"degree":     helpers.InEnum("degree"),
		"pattern":    helpers.InEnum("pattern"),
		"level_id":   "uuid",
		"difficulty": "required|" + helpers.InEnum("pk_difficulty"),
	}
}
func (r *PkStartRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"degree":     "trim",
		"pattern":    "trim",
		"difficulty": "trim",
	}
}
func (r *PkStartRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":    "请选择游戏",
		"game_id.uuid":        "无效的游戏ID",
		"degree.in":           "无效的难度级别",
		"pattern.in":          "无效的练习模式",
		"level_id.uuid":       "无效的关卡ID",
		"difficulty.required": "请选择PK难度",
		"difficulty.in":       "无效的PK难度",
	}
}
func (r *PkStartRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

// ---------- PkNextLevelRequest ----------

type PkNextLevelRequest struct {
	CurrentLevelID string `form:"current_level_id" json:"current_level_id"`
}

func (r *PkNextLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *PkNextLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"current_level_id": "required|uuid",
	}
}
func (r *PkNextLevelRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{}
}
func (r *PkNextLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"current_level_id.required": "请指定当前关卡",
		"current_level_id.uuid":    "无效的关卡ID",
	}
}
func (r *PkNextLevelRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
