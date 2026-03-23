package api

// MarkUnknownRequest validates mark unknown data.
type MarkUnknownRequest struct {
	ContentItemID string `form:"content_item_id" json:"content_item_id"`
	GameID        string `form:"game_id" json:"game_id"`
	GameLevelID   string `form:"game_level_id" json:"game_level_id"`
}
