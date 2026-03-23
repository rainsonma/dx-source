package api

import "github.com/goravel/framework/contracts/http"

// SubmitReportRequest validates game report submission data.
type SubmitReportRequest struct {
	GameID        string  `form:"game_id" json:"game_id"`
	GameLevelID   string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID string  `form:"content_item_id" json:"content_item_id"`
	Reason        string  `form:"reason" json:"reason"`
	Note          *string `form:"note" json:"note"`
}

func (r *SubmitReportRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitReportRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":         "required",
		"game_level_id":   "required",
		"content_item_id": "required",
		"reason":          "required",
	}
}
