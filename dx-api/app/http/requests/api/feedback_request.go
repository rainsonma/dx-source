package api

import "github.com/goravel/framework/contracts/http"

// SubmitFeedbackRequest validates feedback submission data.
type SubmitFeedbackRequest struct {
	Type        string `form:"type" json:"type"`
	Description string `form:"description" json:"description"`
}

func (r *SubmitFeedbackRequest) Authorize(ctx http.Context) error { return nil }

func (r *SubmitFeedbackRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"type":        "required",
		"description": "required|max_len:200",
	}
}

func (r *SubmitFeedbackRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"description.max_len": "description must be at most 200 characters",
	}
}
