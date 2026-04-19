package helpers

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGenerateCode(t *testing.T) {
	for _, length := range []int{1, 4, 6, 16} {
		got := GenerateCode(length)
		if len(got) != length {
			t.Errorf("GenerateCode(%d) len = %d, want %d", length, len(got), length)
		}
		for _, r := range got {
			if r < '0' || r > '9' {
				t.Errorf("GenerateCode(%d) = %q, contains non-digit %q", length, got, r)
				break
			}
		}
	}
}

func TestGenerateInviteCode(t *testing.T) {
	got := GenerateInviteCode(8)
	if len(got) != 8 {
		t.Errorf("GenerateInviteCode(8) len = %d, want 8", len(got))
	}
	for _, r := range got {
		ok := (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
		if !ok {
			t.Errorf("GenerateInviteCode(8) = %q, contains non-alphanumeric %q", got, r)
			break
		}
	}
}

func TestGenerateDefaultNickname(t *testing.T) {
	got := GenerateDefaultNickname()

	if !strings.HasPrefix(got, "斗友_") {
		t.Fatalf("GenerateDefaultNickname() = %q, want prefix %q", got, "斗友_")
	}

	suffix := strings.TrimPrefix(got, "斗友_")
	if utf8.RuneCountInString(suffix) != 6 {
		t.Errorf("suffix rune count = %d, want 6 (got %q)", utf8.RuneCountInString(suffix), suffix)
	}
	for _, r := range suffix {
		if r < '0' || r > '9' {
			t.Errorf("suffix %q contains non-digit %q", suffix, r)
			break
		}
	}

	// Sanity: two back-to-back calls should differ. Not a strict uniqueness
	// claim — this catches a broken PRNG that returns constant output.
	// Collision odds for two legitimate calls: 1 in 10^6.
	a, b := GenerateDefaultNickname(), GenerateDefaultNickname()
	if a == b {
		t.Errorf("two consecutive calls returned the same nickname %q", a)
	}
}
