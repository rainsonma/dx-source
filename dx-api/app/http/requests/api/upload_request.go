package api

import "github.com/goravel/framework/contracts/http"

// UploadImageRequest holds the metadata fields from the multipart upload form.
type UploadImageRequest struct {
	Role string `form:"role" json:"role"`
}

func (r *UploadImageRequest) Authorize(ctx http.Context) error { return nil }
func (r *UploadImageRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{"role": "required"}
}
