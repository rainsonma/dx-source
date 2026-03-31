package api

import "github.com/goravel/framework/contracts/http"

type CreatePostRequest struct {
	Content string   `form:"content" json:"content"`
	ImageID *string  `form:"image_id" json:"image_id"`
	Tags    []string `form:"tags" json:"tags"`
}

func (r *CreatePostRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreatePostRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:2000",
		"tags":    "max_len:5",
	}
}
func (r *CreatePostRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}
func (r *CreatePostRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入帖子内容",
		"content.min_len":  "帖子内容不能为空",
		"content.max_len":  "帖子内容不能超过2000个字符",
		"tags.max_len":     "标签最多5个",
	}
}

type UpdatePostRequest struct {
	Content string   `form:"content" json:"content"`
	ImageID *string  `form:"image_id" json:"image_id"`
	Tags    []string `form:"tags" json:"tags"`
}

func (r *UpdatePostRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdatePostRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:2000",
		"tags":    "max_len:5",
	}
}
func (r *UpdatePostRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}
func (r *UpdatePostRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入帖子内容",
		"content.min_len":  "帖子内容不能为空",
		"content.max_len":  "帖子内容不能超过2000个字符",
		"tags.max_len":     "标签最多5个",
	}
}
