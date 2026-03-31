package api

import "github.com/goravel/framework/contracts/http"

type CreateCommentRequest struct {
	Content  string  `form:"content" json:"content"`
	ParentID *string `form:"parent_id" json:"parent_id"`
}

func (r *CreateCommentRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateCommentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:500",
	}
}
func (r *CreateCommentRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}
func (r *CreateCommentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入评论内容",
		"content.min_len":  "评论内容不能为空",
		"content.max_len":  "评论内容不能超过500个字符",
	}
}

type UpdateCommentRequest struct {
	Content string `form:"content" json:"content"`
}

func (r *UpdateCommentRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateCommentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:500",
	}
}
func (r *UpdateCommentRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}
func (r *UpdateCommentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入评论内容",
		"content.min_len":  "评论内容不能为空",
		"content.max_len":  "评论内容不能超过500个字符",
	}
}
