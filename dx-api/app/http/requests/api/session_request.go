package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
)

// StartSessionRequest validates session start data.
type StartSessionRequest struct {
	GameID  string  `form:"game_id" json:"game_id"`
	Degree  string  `form:"degree" json:"degree"`
	Pattern *string `form:"pattern" json:"pattern"`
	LevelID *string `form:"level_id" json:"level_id"`
}

func (r *StartSessionRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartSessionRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id": "required",
	}
}
func (r *StartSessionRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

// CheckActiveSessionRequest validates active session check data.
type CheckActiveSessionRequest struct {
	GameID  string  `form:"game_id" json:"game_id"`
	Degree  string  `form:"degree" json:"degree"`
	Pattern *string `form:"pattern" json:"pattern"`
}

// CheckActiveLevelSessionRequest validates active level session check data.
type CheckActiveLevelSessionRequest struct {
	GameID      string  `form:"game_id" json:"game_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
}

// StartLevelRequest validates level start data.
type StartLevelRequest struct {
	GameLevelID string  `form:"game_level_id" json:"game_level_id"`
	Degree      string  `form:"degree" json:"degree"`
	Pattern     *string `form:"pattern" json:"pattern"`
}

func (r *StartLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *StartLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required",
	}
}
func (r *StartLevelRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeIntermediate)
	}
	return nil
}

// CompleteLevelRequest validates level completion data.
type CompleteLevelRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
	Score       int    `form:"score" json:"score"`
	MaxCombo    int    `form:"max_combo" json:"max_combo"`
	TotalItems  int    `form:"total_items" json:"total_items"`
}

// AdvanceLevelRequest validates level advance data.
type AdvanceLevelRequest struct {
	NextLevelID string `form:"next_level_id" json:"next_level_id"`
}

// RestartLevelRequest validates level restart data.
type RestartLevelRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
}

// RecordAnswerRequest validates answer recording data.
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
		"game_session_level_id": "required",
		"game_level_id":         "required",
		"content_item_id":       "required",
	}
}

// RecordSkipRequest validates skip recording data.
type RecordSkipRequest struct {
	GameLevelID       string  `form:"game_level_id" json:"game_level_id"`
	PlayTime          int     `form:"play_time" json:"play_time"`
	NextContentItemID *string `form:"next_content_item_id" json:"next_content_item_id"`
}

func (r *RecordSkipRequest) Authorize(ctx http.Context) error { return nil }
func (r *RecordSkipRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required",
	}
}

// SyncPlayTimeRequest validates playtime sync data.
type SyncPlayTimeRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
	PlayTime    int    `form:"play_time" json:"play_time"`
}

func (r *SyncPlayTimeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SyncPlayTimeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_level_id": "required",
	}
}

// UpdateContentItemRequest validates content item update data.
type UpdateContentItemRequest struct {
	ContentItemID *string `form:"content_item_id" json:"content_item_id"`
}

// EndSessionRequest validates session end data.
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
		"game_id": "required",
	}
}

// RestoreSessionRequest validates restore data query params.
type RestoreSessionRequest struct {
	GameLevelID string `form:"game_level_id" json:"game_level_id"`
}
