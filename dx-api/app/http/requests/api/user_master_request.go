package api

// MarkMasteredRequest validates mark mastered data.
type MarkMasteredRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}

// BulkDeleteRequest validates bulk delete data.
type BulkDeleteRequest struct {
	IDs []string `form:"ids" json:"ids"`
}
