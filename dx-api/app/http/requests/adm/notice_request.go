package adm

import "github.com/goravel/framework/contracts/http"

// CreateNoticeRequest validates notice creation data.
type CreateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}

func (r *CreateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateNoticeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"title": "required|max_len:200",
	}
}
func (r *CreateNoticeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.max_len": "title must be at most 200 characters",
	}
}

// UpdateNoticeRequest validates notice update data.
type UpdateNoticeRequest struct {
	Title   string  `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Icon    *string `form:"icon" json:"icon"`
}

func (r *UpdateNoticeRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateNoticeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"title": "required|max_len:200",
	}
}
func (r *UpdateNoticeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.max_len": "title must be at most 200 characters",
	}
}
