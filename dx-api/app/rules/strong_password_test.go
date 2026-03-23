package rules

import (
	"context"
	"testing"
)

func TestStrongPassword_Passes(t *testing.T) {
	rule := &StrongPassword{}

	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"valid", "Abc123!@", true},
		{"missing uppercase", "abc123!@", false},
		{"missing lowercase", "ABC123!@", false},
		{"missing digit", "Abcdef!@", false},
		{"missing special", "Abc12345", false},
		{"empty", "", false},
		{"all types", "P@ssw0rd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.Passes(context.Background(), nil, tt.val)
			if got != tt.want {
				t.Errorf("Passes(%q) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestStrongPassword_Signature(t *testing.T) {
	rule := &StrongPassword{}
	if rule.Signature() != "strong_password" {
		t.Errorf("Signature() = %q, want %q", rule.Signature(), "strong_password")
	}
}
