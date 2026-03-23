package api

// SubmitReportRequest validates game report submission data.
type SubmitReportRequest struct {
	GameID        string  `form:"game_id" json:"game_id"`
	GameLevelID   string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID string  `form:"content_item_id" json:"content_item_id"`
	Reason        string  `form:"reason" json:"reason"`
	Note          *string `form:"note" json:"note"`
}
