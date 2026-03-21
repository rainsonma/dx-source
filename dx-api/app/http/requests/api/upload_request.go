package api

// UploadImageRequest holds the metadata fields from the multipart upload form.
type UploadImageRequest struct {
	Role string `form:"role" json:"role"`
}
