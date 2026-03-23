package api

import (
	"github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
)

type UploadImageRequest struct {
	Role string `form:"role" json:"role"`
}

func (r *UploadImageRequest) Authorize(ctx http.Context) error { return nil }
func (r *UploadImageRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"role": "required|" + helpers.InEnum("image_role"),
	}
}
func (r *UploadImageRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"role.required": "请指定图片用途",
		"role.in":       "无效的图片用途",
	}
}
