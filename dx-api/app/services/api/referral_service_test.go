package api

import (
	"testing"
)

func TestDerefStr(t *testing.T) {
	tests := []struct {
		name  string
		input *string
		want  string
	}{
		{"nil returns empty", nil, ""},
		{"non-nil returns value", strPtr("hello"), "hello"},
		{"empty string returns empty", strPtr(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := derefStr(tt.input)
			if got != tt.want {
				t.Errorf("derefStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
