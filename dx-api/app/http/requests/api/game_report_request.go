package api

import "github.com/goravel/framework/contracts/http"

type SubmitReportRequest struct {
	GameID         string  `form:"game_id" json:"game_id"`
	GameLevelID    string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID  *string `form:"content_item_id" json:"content_item_id"`
	ContentVocabID *string `form:"content_vocab_id" json:"content_vocab_id"`
	Reason         string  `form:"reason" json:"reason"`
	Note           *string `form:"note" json:"note"`
}

func (r *SubmitReportRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitReportRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"game_level_id": "required|uuid",
		"reason":        "required|max_len:200",
		"note":          "max_len:500",
	}
}
func (r *SubmitReportRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"reason": "trim",
		"note":   "trim",
	}
}
func (r *SubmitReportRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":       "请指定游戏",
		"game_id.uuid":           "无效的游戏ID",
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
		"reason.required":        "请选择举报原因",
		"reason.max_len":         "举报原因不能超过200个字符",
		"note.max_len":           "备注不能超过500个字符",
	}
}
