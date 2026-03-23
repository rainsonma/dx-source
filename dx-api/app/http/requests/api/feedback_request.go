package api

import (
	"github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
)

type SubmitFeedbackRequest struct {
	Type        string `form:"type" json:"type"`
	Description string `form:"description" json:"description"`
}

func (r *SubmitFeedbackRequest) Authorize(ctx http.Context) error { return nil }
func (r *SubmitFeedbackRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"type":        "required|" + helpers.InEnum("feedback_type"),
		"description": "required|min_len:2|max_len:200",
	}
}
func (r *SubmitFeedbackRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"description": "trim",
	}
}
func (r *SubmitFeedbackRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.required":        "请选择反馈类型",
		"type.in":              "无效的反馈类型",
		"description.required": "请输入反馈内容",
		"description.min_len":  "反馈内容不能少于2个字符",
		"description.max_len":  "描述不能超过200个字符",
	}
}
