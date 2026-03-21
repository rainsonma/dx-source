package api

// SubmitFeedbackRequest validates feedback submission data.
type SubmitFeedbackRequest struct {
	Type        string `form:"type" json:"type"`
	Description string `form:"description" json:"description"`
}

// SubmitReportRequest validates game report submission data.
type SubmitReportRequest struct {
	GameID        string  `form:"game_id" json:"game_id"`
	GameLevelID   string  `form:"game_level_id" json:"game_level_id"`
	ContentItemID string  `form:"content_item_id" json:"content_item_id"`
	Reason        string  `form:"reason" json:"reason"`
	Note          *string `form:"note" json:"note"`
}

// RedeemCodeRequest validates a redeem code submission.
type RedeemCodeRequest struct {
	Code string `form:"code" json:"code"`
}

// SubmitContentSeekRequest validates a content seek submission.
type SubmitContentSeekRequest struct {
	CourseName  string `form:"course_name" json:"course_name"`
	Description string `form:"description" json:"description"`
	DiskUrl     string `form:"disk_url" json:"disk_url"`
}
