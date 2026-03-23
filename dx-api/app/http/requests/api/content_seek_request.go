package api

import "github.com/goravel/framework/contracts/http"

// SubmitContentSeekRequest validates a content seek submission.
type SubmitContentSeekRequest struct {
	CourseName  string `form:"course_name" json:"course_name"`
	Description string `form:"description" json:"description"`
	DiskUrl     string `form:"disk_url" json:"disk_url"`
}

func (r *SubmitContentSeekRequest) Authorize(ctx http.Context) error { return nil }

func (r *SubmitContentSeekRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name": "required|max_len:30",
		"description": "required|max_len:30",
		"disk_url":    "required|max_len:30",
	}
}

func (r *SubmitContentSeekRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name.max_len": "course name must be at most 30 characters",
		"description.max_len": "description must be at most 30 characters",
		"disk_url.max_len":    "disk url must be at most 30 characters",
	}
}
