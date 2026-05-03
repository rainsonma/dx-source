package api

import "github.com/goravel/framework/contracts/http"

type MarkUnknownRequest struct {
	ContentItemID  *string `form:"content_item_id" json:"content_item_id"`
	ContentVocabID *string `form:"content_vocab_id" json:"content_vocab_id"`
	GameID         string  `form:"game_id" json:"game_id"`
	GameLevelID    string  `form:"game_level_id" json:"game_level_id"`
}

func (r *MarkUnknownRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkUnknownRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":       "required|uuid",
		"game_level_id": "required|uuid",
	}
}
func (r *MarkUnknownRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":       "请选择游戏",
		"game_id.uuid":           "无效的游戏ID",
		"game_level_id.required": "请指定关卡",
		"game_level_id.uuid":     "无效的关卡ID",
	}
}
