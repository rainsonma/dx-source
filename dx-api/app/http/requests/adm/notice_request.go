package adm

import "github.com/goravel/framework/contracts/http"

type CreateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}

func (r *CreateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateNoticeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "required|min_len:2|max_len:200",
		"content": "required|max_len:5000",
		"icon":    "max_len:50",
	}
}
func (r *CreateNoticeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "trim",
		"content": "trim",
		"icon":    "trim",
	}
}
func (r *CreateNoticeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.required":   "Title is required",
		"title.min_len":    "Title must be at least 2 characters",
		"title.max_len":    "Title must not exceed 200 characters",
		"content.required": "Content is required",
		"content.max_len":  "Content must not exceed 5000 characters",
		"icon.max_len":     "Icon must not exceed 50 characters",
	}
}

type UpdateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}

func (r *UpdateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateNoticeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "required|min_len:2|max_len:200",
		"content": "required|max_len:5000",
		"icon":    "max_len:50",
	}
}
func (r *UpdateNoticeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"title":   "trim",
		"content": "trim",
		"icon":    "trim",
	}
}
func (r *UpdateNoticeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.required":   "Title is required",
		"title.min_len":    "Title must be at least 2 characters",
		"title.max_len":    "Title must not exceed 200 characters",
		"content.required": "Content is required",
		"content.max_len":  "Content must not exceed 5000 characters",
		"icon.max_len":     "Icon must not exceed 50 characters",
	}
}
