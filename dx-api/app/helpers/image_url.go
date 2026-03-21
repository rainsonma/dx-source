package helpers

import "fmt"

// ImageServeURL returns the API-relative URL for serving an image by its ID.
func ImageServeURL(imageID string) string {
	return fmt.Sprintf("/api/uploads/images/%s", imageID)
}
