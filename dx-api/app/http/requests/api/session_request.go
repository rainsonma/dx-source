package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

// ---------- StartSessionRequest ----------

type StartSessionRequest struct {
	GameID  string  `form:"game_id" json:"game_id"`
	Degree  string  `form:"degree" json:"degree"`
	Pattern *string `form:"pattern" json:"pattern"`
	LevelID *string `form:"level_id" json:"level_id"`
}

func (r *StartSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":  "required|uuid",
		"degree":   helpers.InEnum("degree"),
		"pattern":  helpers.InEnum("pattern"),
		"level_id": "uuid",
	}
}
func (r *StartSessionRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"degree":  "trim",
		"pattern": "trim",
	}
}
func (r *StartSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required": "请选择游戏",
		"game_id.uuid":     "无效的游戏ID",
		"degree.in":        "无效的难度级别",
		"pattern.in":       "无效的练习模式",
		"level_id.uuid":    "无效的关卡ID",
	}
}
func (r *StartSessionRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

// ---------- CheckActiveSessionRequest ----------

type CheckActiveSessionRequest struct {
	GameID  string  `form:"game_id" json:"game_id"`
	Degree  string  `form:"degree" json:"degree"`
	Pattern *string `form:"pattern" json:"pattern"`
}

func (r *CheckActiveSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *CheckActiveSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id": "required|uuid",
		"degree":  helpers.InEnum("degree"),
		"pattern": helpers.InEnum("pattern"),
	}
}
func (r *CheckActiveSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required": "请选择游戏",
		"game_id.uuid":     "无效的游戏ID",
		"degree.in":        "无效的难度级别",
		"pattern.in":       "无效的练习模式",
	}
}

// ---------- CheckActiveLevelSessionRequest ----------

type CheckActiveLevelSessionRequest struct {
	GameID      string  `form:"game_id" json:"game_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
}

func (r *CheckActiveLevelSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *CheckActiveLevelSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
		"game_level_id": "required|uuid",
	}
}
func (r *CheckActiveLevelSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":       "请选择游戏",
		"game_id.uuid":           "无效的游戏ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
	}
}

// ---------- StartLevelRequest ----------

type StartLevelRequest struct {
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
}

func (r *StartLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
	}
}
func (r *StartLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
	}
}
func (r *StartLevelRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

// ---------- CompleteLevelRequest ----------
// Controller reads game_level_id from route param — only body fields validated here.

type CompleteLevelRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
	Score       int    `form:"score" json:"score"`
	MaxCombo    int    `form:"max_combo" json:"max_combo"`
	TotalItems  int    `form:"total_items" json:"total_items"`
}

func (r *CompleteLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *CompleteLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"score":       "min:0",
		"max_combo":   "min:0",
		"total_items": "min:0",
	}
}
func (r *CompleteLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"score.min":       "分数不能为负数",
		"max_combo.min":   "连击数不能为负数",
		"total_items.min": "总数不能为负数",
	}
}

// ---------- AdvanceLevelRequest ----------
// Controller has fallback: if next_level_id is empty, uses route param levelId.

type AdvanceLevelRequest struct {
	NextLevelID string `form:"next_level_id" json:"next_level_id"`
}

func (r *AdvanceLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *AdvanceLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"next_level_id": "uuid",
	}
}
func (r *AdvanceLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"next_level_id.uuid": "无效的关卡ID",
	}
}

// ---------- RecordAnswerRequest ----------

type RecordAnswerRequest struct {
	GameSessionLevelID string  `form:"game_session_level_id" json:"game_session_level_id"`
	GameLevelID        string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID      string  `form:"content_item_id" json:"content_item_id"`
	IsCorrect          bool    `form:"is_correct" json:"is_correct"`
	UserAnswer         string  `form:"user_answer" json:"user_answer"`
	SourceAnswer       string  `form:"source_answer" json:"source_answer"`
	BaseScore          int     `form:"base_score" json:"base_score"`
	ComboScore         int     `form:"combo_score" json:"combo_score"`
	Score              int     `form:"score" json:"score"`
	MaxCombo           int     `form:"max_combo" json:"max_combo"`
	PlayTime           int     `form:"play_time" json:"play_time"`
	NextContentItemID  *string `form:"next_content_item_id" json:"next_content_item_id"`
	Duration           int     `form:"duration" json:"duration"`
}

func (r *RecordAnswerRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordAnswerRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_level_id": "required|uuid",
		"game_level_id":         "required|uuid",
		"content_item_id":       "required|uuid",
		"base_score":            "min:0",
		"combo_score":           "min:0",
		"score":                 "min:0",
		"max_combo":             "min:0",
		"play_time":             "min:0",
		"duration":              "min:0",
		"next_content_item_id":  "uuid",
	}
}
func (r *RecordAnswerRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_level_id.required": "请指定关卡会话",
		"game_session_level_id.uuid":     "无效的关卡会话ID",
		"game_level_id.required":         "请指定关卡",
		"game_level_id.uuid":             "无效的关卡ID",
		"content_item_id.required":       "请指定内容项",
		"content_item_id.uuid":           "无效的内容项ID",
		"base_score.min":                 "基础分数不能为负数",
		"combo_score.min":                "连击分数不能为负数",
		"score.min":                      "分数不能为负数",
		"max_combo.min":                  "最大连击不能为负数",
		"play_time.min":                  "游玩时长不能为负数",
		"duration.min":                   "持续时间不能为负数",
		"next_content_item_id.uuid":      "无效的内容项ID",
	}
}

// ---------- RecordSkipRequest ----------

type RecordSkipRequest struct {
	GameLevelID       string  `form:"game_level_id" json:"game_level_id"`
	PlayTime          int     `form:"play_time" json:"play_time"`
	NextContentItemID *string `form:"next_content_item_id" json:"next_content_item_id"`
}

func (r *RecordSkipRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordSkipRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id":        "required|uuid",
		"play_time":            "min:0",
		"next_content_item_id": "uuid",
	}
}
func (r *RecordSkipRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required":    "请指定关卡",
		"game_level_id.uuid":        "无效的关卡ID",
		"play_time.min":             "游玩时长不能为负数",
		"next_content_item_id.uuid": "无效的内容项ID",
	}
}

// ---------- SyncPlayTimeRequest ----------

type SyncPlayTimeRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
	PlayTime    int    `form:"play_time" json:"play_time"`
}

func (r *SyncPlayTimeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SyncPlayTimeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required|uuid",
		"play_time":     "required|min:0",
	}
}
func (r *SyncPlayTimeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"play_time.required":     "请提供游玩时长",
		"play_time.min":          "游玩时长不能为负数",
	}
}

// ---------- UpdateContentItemRequest ----------

type UpdateContentItemRequest struct {
	ContentItemID *string `form:"content_item_id" json:"content_item_id"`
}

func (r *UpdateContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "uuid",
	}
}
func (r *UpdateContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id.uuid": "无效的内容ID",
	}
}

// ---------- EndSessionRequest ----------

type EndSessionRequest struct {
	GameID             string `form:"game_id" json:"game_id"`
	Score              int    `form:"score" json:"score"`
	Exp                int    `form:"exp" json:"exp"`
	MaxCombo           int    `form:"max_combo" json:"max_combo"`
	CorrectCount       int    `form:"correct_count" json:"correct_count"`
	WrongCount         int    `form:"wrong_count" json:"wrong_count"`
	SkipCount          int    `form:"skip_count" json:"skip_count"`
	AllLevelsCompleted bool   `form:"all_levels_completed" json:"all_levels_completed"`
}

func (r *EndSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *EndSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"score":         "min:0",
		"exp":           "min:0",
		"max_combo":     "min:0",
		"correct_count": "min:0",
		"wrong_count":   "min:0",
		"skip_count":    "min:0",
	}
}
func (r *EndSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":  "请选择游戏",
		"game_id.uuid":      "无效的游戏ID",
		"score.min":         "分数不能为负数",
		"exp.min":           "经验值不能为负数",
		"max_combo.min":     "最大连击不能为负数",
		"correct_count.min": "正确数不能为负数",
		"wrong_count.min":   "错误数不能为负数",
		"skip_count.min":    "跳过数不能为负数",
	}
}

// ---------- RestoreSessionRequest ----------
// Controller currently reads from query params — converting to FormRequest.

type RestoreSessionRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
}

func (r *RestoreSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *RestoreSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required|uuid",
	}
}
func (r *RestoreSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
	}
}
