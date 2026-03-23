package api

import "github.com/goravel/framework/contracts/http"

type MarkMasteredRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}

func (r *MarkMasteredRequest) Authorize(ctx http.Context) error { return nil }
func (r *MarkMasteredRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id": "required|uuid",
		"game_id":         "required|uuid",
		"game_level_id":   "required|uuid",
	}
}
func (r *MarkMasteredRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content_item_id.required": "请指定内容",
		"content_item_id.uuid":     "无效的内容ID",
		"game_id.required":         "请选择游戏",
		"game_id.uuid":             "无效的游戏ID",
		"game_level_id.required":   "请指定关卡",
		"game_level_id.uuid":       "无效的关卡ID",
	}
}

type BulkDeleteRequest struct {
	IDs []string `form:"ids" json:"ids"`
}

func (r *BulkDeleteRequest) Authorize(ctx http.Context) error { return nil }
func (r *BulkDeleteRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"ids":   "required|min_len:1|max_len:100",
		"ids.*": "uuid",
	}
}
func (r *BulkDeleteRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"ids.required": "请选择要删除的项目",
		"ids.min_len":  "请至少选择一项",
		"ids.max_len":  "单次最多删除100条",
		"ids.*.uuid":   "包含无效的ID",
	}
}
