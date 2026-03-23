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
		"course_name.max_len": "课程名称不能超过30个字符",
		"description.max_len": "描述不能超过30个字符",
		"disk_url.max_len":    "网盘链接不能超过30个字符",
	}
}
