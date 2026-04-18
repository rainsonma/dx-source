package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

// ---------- StartSessionRequest ----------

type StartSessionRequest struct {
	GameID      string  `form:"game_id" json:"game_id"`
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
}

func (r *StartSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"game_level_id": "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
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
		"game_id.required":       "请选择游戏",
		"game_id.uuid":           "无效的游戏ID",
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
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
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
}

func (r *CheckActiveSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *CheckActiveSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required|uuid",
		"degree":        helpers.InEnum("degree"),
		"pattern":       helpers.InEnum("pattern"),
	}
}
func (r *CheckActiveSessionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"degree.in":              "无效的难度级别",
		"pattern.in":             "无效的练习模式",
	}
}

// ---------- CompleteLevelRequest ----------

type CompleteLevelRequest struct {
	Score      int `form:"score" json:"score"`
	MaxCombo   int `form:"max_combo" json:"max_combo"`
	TotalItems int `form:"total_items" json:"total_items"`
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

// ---------- RecordAnswerRequest ----------

type RecordAnswerRequest struct {
	GameSessionId     string  `form:"game_session_id" json:"game_session_id"`
	GameLevelID       string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID     string  `form:"content_item_id" json:"content_item_id"`
	IsCorrect         bool    `form:"is_correct" json:"is_correct"`
	UserAnswer        string  `form:"user_answer" json:"user_answer"`
	SourceAnswer      string  `form:"source_answer" json:"source_answer"`
	BaseScore         int     `form:"base_score" json:"base_score"`
	ComboScore        int     `form:"combo_score" json:"combo_score"`
	Score             int     `form:"score" json:"score"`
	MaxCombo          int     `form:"max_combo" json:"max_combo"`
	PlayTime          int     `form:"play_time" json:"play_time"`
	NextContentItemID *string `form:"next_content_item_id" json:"next_content_item_id"`
	Duration          int     `form:"duration" json:"duration"`
}

func (r *RecordAnswerRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordAnswerRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_id":      "required|uuid",
		"game_level_id":        "required|uuid",
		"content_item_id":      "required|uuid",
		"base_score":           "min:0",
		"combo_score":          "min:0",
		"score":                "min:0",
		"max_combo":            "min:0",
		"play_time":            "min:0",
		"duration":             "min:0",
		"next_content_item_id": "uuid",
	}
}
func (r *RecordAnswerRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_id.required":  "请指定游戏会话",
		"game_session_id.uuid":      "无效的游戏会话ID",
		"game_level_id.required":    "请指定关卡",
		"game_level_id.uuid":        "无效的关卡ID",
		"content_item_id.required":  "请指定内容项",
		"content_item_id.uuid":      "无效的内容项ID",
		"base_score.min":            "基础分数不能为负数",
		"combo_score.min":           "连击分数不能为负数",
		"score.min":                 "分数不能为负数",
		"max_combo.min":             "最大连击不能为负数",
		"play_time.min":             "游玩时长不能为负数",
		"duration.min":              "持续时间不能为负数",
		"next_content_item_id.uuid": "无效的内容项ID",
	}
}

// ---------- RecordSkipRequest ----------

type RecordSkipRequest struct {
	GameSessionId     string  `form:"game_session_id" json:"game_session_id"`
	GameLevelID       string  `form:"game_level_id" json:"game_level_id"`
	PlayTime          int     `form:"play_time" json:"play_time"`
	NextContentItemID *string `form:"next_content_item_id" json:"next_content_item_id"`
}

func (r *RecordSkipRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordSkipRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_id":      "required|uuid",
		"game_level_id":        "required|uuid",
		"play_time":            "min:0",
		"next_content_item_id": "uuid",
	}
}
func (r *RecordSkipRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_session_id.required":  "请指定游戏会话",
		"game_session_id.uuid":      "无效的游戏会话ID",
		"game_level_id.required":    "请指定关卡",
		"game_level_id.uuid":        "无效的关卡ID",
		"play_time.min":             "游玩时长不能为负数",
		"next_content_item_id.uuid": "无效的内容项ID",
	}
}

// ---------- SyncPlayTimeRequest ----------

type SyncPlayTimeRequest struct {
	PlayTime int `form:"play_time" json:"play_time"`
}

func (r *SyncPlayTimeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SyncPlayTimeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"play_time": "required|min:0",
	}
}
func (r *SyncPlayTimeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"play_time.required": "请提供游玩时长",
		"play_time.min":      "游玩时长不能为负数",
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
	Score        int `form:"score" json:"score"`
	Exp          int `form:"exp" json:"exp"`
	MaxCombo     int `form:"max_combo" json:"max_combo"`
	CorrectCount int `form:"correct_count" json:"correct_count"`
	WrongCount   int `form:"wrong_count" json:"wrong_count"`
	SkipCount    int `form:"skip_count" json:"skip_count"`
}

func (r *EndSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *EndSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
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
		"score.min":         "分数不能为负数",
		"exp.min":           "经验值不能为负数",
		"max_combo.min":     "最大连击不能为负数",
		"correct_count.min": "正确数不能为负数",
		"wrong_count.min":   "错误数不能为负数",
		"skip_count.min":    "跳过数不能为负数",
	}
}
