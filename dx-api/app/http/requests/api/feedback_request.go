package api

// SubmitFeedbackRequest validates feedback submission data.
type SubmitFeedbackRequest struct {
	Type        string `form:"type" json:"type"`
	Description string `form:"description" json:"description"`
}
