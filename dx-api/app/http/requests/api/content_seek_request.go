package api

import "github.com/goravel/framework/contracts/http"

type SubmitContentSeekRequest struct {
	CourseName  string `form:"course_name" json:"course_name"`
	Description string `form:"description" json:"description"`
	DiskUrl     string `form:"disk_url" json:"disk_url"`
}

func (r *SubmitContentSeekRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitContentSeekRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name": "required|min_len:2|max_len:30",
		"description": "required|min_len:2|max_len:200",
		"disk_url":    "required|full_url|max_len:500",
	}
}
func (r *SubmitContentSeekRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name": "trim",
		"description": "trim",
		"disk_url":    "trim",
	}
}
func (r *SubmitContentSeekRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"course_name.required": "请输入课程名称",
		"course_name.min_len":  "课程名称至少需要2个字符",
		"course_name.max_len":  "课程名称不能超过30个字符",
		"description.required": "请输入描述",
		"description.min_len":  "描述至少需要2个字符",
		"description.max_len":  "描述不能超过200个字符",
		"disk_url.required":    "请输入网盘链接",
		"disk_url.full_url":    "请输入有效的网盘链接",
		"disk_url.max_len":     "网盘链接不能超过500个字符",
	}
}
