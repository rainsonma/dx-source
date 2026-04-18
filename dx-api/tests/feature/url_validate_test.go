package feature

import (
	"testing"

	"dx-api/app/helpers"
)

func TestIsUploadedImageURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid jpg", "/api/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", true},
		{"valid png", "/api/uploads/images/2026/4/8/0194a2f0-7b7b-7000-8000-0194a2f07b7b.png", true},
		{"empty", "", false},
		{"absolute http", "https://evil.com/img.jpg", false},
		{"wrong prefix", "/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"audio prefix", "/api/uploads/audios/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"bad extension", "/api/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.gif", false},
		{"no extension", "/api/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b", false},
		{"non-uuid filename", "/api/uploads/images/2026/04/18/hello.jpg", false},
		{"traversal", "/api/uploads/images/2026/04/18/../../../etc/passwd", false},
		{"5-digit year", "/api/uploads/images/20260/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"javascript scheme", "javascript:alert(1)", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := helpers.IsUploadedImageURL(c.in)
			if got != c.want {
				t.Fatalf("IsUploadedImageURL(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
