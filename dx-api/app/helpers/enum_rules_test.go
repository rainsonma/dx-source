package helpers

import (
	"strings"
	"testing"
)

func TestInEnum_KnownEnum(t *testing.T) {
	result := InEnum("degree")
	if !strings.HasPrefix(result, "in:") {
		t.Fatalf("expected 'in:' prefix, got %q", result)
	}
	if !strings.Contains(result, "intermediate") {
		t.Fatalf("expected 'intermediate' in result, got %q", result)
	}
}

func TestInEnum_UnknownEnum_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unknown enum")
		}
	}()
	InEnum("nonexistent")
}
