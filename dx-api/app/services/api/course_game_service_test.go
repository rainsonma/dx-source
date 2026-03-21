package api

import (
	"testing"
)

func TestContentLimitConstants(t *testing.T) {
	if MaxSentences != 20 {
		t.Errorf("MaxSentences = %d, want 20", MaxSentences)
	}
	if MaxVocab != 200 {
		t.Errorf("MaxVocab = %d, want 200", MaxVocab)
	}
	if MaxItemsPerMeta != 50 {
		t.Errorf("MaxItemsPerMeta = %d, want 50", MaxItemsPerMeta)
	}
}

func TestSourceTypeConstants(t *testing.T) {
	if SourceTypeSentence != "sentence" {
		t.Errorf("SourceTypeSentence = %q, want %q", SourceTypeSentence, "sentence")
	}
	if SourceTypeVocab != "vocab" {
		t.Errorf("SourceTypeVocab = %q, want %q", SourceTypeVocab, "vocab")
	}
}

func TestCapacityFormula(t *testing.T) {
	// The capacity formula: sentences/MaxSentences + vocabs/MaxVocab <= 1
	tests := []struct {
		name      string
		sentences int
		vocabs    int
		exceeds   bool
	}{
		{"empty", 0, 0, false},
		{"max sentences only", MaxSentences, 0, false},
		{"max vocabs only", 0, MaxVocab, false},
		{"half and half", MaxSentences / 2, MaxVocab / 2, false},
		{"one over sentence limit", MaxSentences + 1, 0, true},
		{"one over vocab limit", 0, MaxVocab + 1, true},
		{"combined exceeds", MaxSentences/2 + 1, MaxVocab/2 + 1, true},
		{"10 sentences + 100 vocabs = exactly 1.0", 10, 100, false},
		{"11 sentences + 100 vocabs > 1.0", 11, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exceeds := float64(tt.sentences)/float64(MaxSentences)+float64(tt.vocabs)/float64(MaxVocab) > 1
			if exceeds != tt.exceeds {
				t.Errorf("capacity(%d sentences, %d vocabs) exceeds = %v, want %v",
					tt.sentences, tt.vocabs, exceeds, tt.exceeds)
			}
		})
	}
}

func TestErrorSentinels(t *testing.T) {
	// Verify key error sentinels are distinct and non-nil
	errors := map[string]error{
		"ErrGameNotFound":         ErrGameNotFound,
		"ErrGamePublished":        ErrGamePublished,
		"ErrGameAlreadyPublished": ErrGameAlreadyPublished,
		"ErrGameNotPublished":     ErrGameNotPublished,
		"ErrLevelNotFound":        ErrLevelNotFound,
		"ErrNoGameLevels":         ErrNoGameLevels,
		"ErrForbidden":            ErrForbidden,
		"ErrCapacityExceeded":     ErrCapacityExceeded,
		"ErrItemLimitExceeded":    ErrItemLimitExceeded,
		"ErrMetaNotFound":         ErrMetaNotFound,
		"ErrContentItemNotFound":  ErrContentItemNotFound,
	}

	for name, err := range errors {
		if err == nil {
			t.Errorf("%s should not be nil", name)
		}
	}

	// Verify distinct error messages
	seen := make(map[string]string)
	for name, err := range errors {
		msg := err.Error()
		if prev, ok := seen[msg]; ok {
			t.Errorf("%s and %s have the same error message: %q", name, prev, msg)
		}
		seen[msg] = name
	}
}
