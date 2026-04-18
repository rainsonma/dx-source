package feature

import (
	"testing"

	services "dx-api/app/services/api"
)

func TestResolveImagePath(t *testing.T) {
	// Set STORAGE_PATH so the resolver doesn't call facades.Config() during valid-path tests.
	t.Setenv("STORAGE_PATH", t.TempDir())

	cases := []struct {
		name    string
		y, m, d string
		f       string
		wantErr bool
	}{
		{"valid", "2026", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"valid png", "2026", "4", "8", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.png", false},
		{"traversal via filename", "2026", "04", "18", "../../../etc/passwd", true},
		{"non-numeric year", "abcd", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", true},
		{"5-digit year", "20260", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", true},
		{"bad extension", "2026", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.gif", true},
		{"short uuid", "2026", "04", "18", "abc.jpg", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := services.ResolveImagePath(c.y, c.m, c.d, c.f)
			if (err != nil) != c.wantErr {
				t.Fatalf("ResolveImagePath(%q,%q,%q,%q) err=%v wantErr=%v", c.y, c.m, c.d, c.f, err, c.wantErr)
			}
		})
	}
}
