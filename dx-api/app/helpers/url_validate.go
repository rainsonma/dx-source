package helpers

import "regexp"

var uploadImageURLRegex = regexp.MustCompile(`^/api/uploads/images/\d{4}/\d{1,2}/\d{1,2}/[0-9a-f-]{36}\.(jpg|png)$`)

// IsUploadedImageURL reports whether s is a URL produced by this service's
// image upload endpoint. Used to validate client-submitted *_url fields.
func IsUploadedImageURL(s string) bool {
	return uploadImageURLRegex.MatchString(s)
}
