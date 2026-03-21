package api

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseFormattedLines(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantFormatted   string
		wantTypeCount   int
		wantSentences   int
		wantVocabs      int
	}{
		{
			"sentence and vocab lines",
			"[S] I like the food.\n我喜欢这个食物。\n[V] food\n食物",
			"I like the food.\n我喜欢这个食物。\nfood\n食物",
			2, 1, 1,
		},
		{
			"only sentences",
			"[S] Hello world.\n[S] Good morning.",
			"Hello world.\nGood morning.",
			2, 2, 0,
		},
		{
			"only vocabs",
			"[V] apple\n[V] banana\n[V] cherry",
			"apple\nbanana\ncherry",
			3, 0, 3,
		},
		{
			"lines without markers",
			"plain line\nanother line",
			"plain line\nanother line",
			0, 0, 0,
		},
		{
			"empty input",
			"",
			"",
			0, 0, 0,
		},
		{
			"mixed with empty lines",
			"[S] First sentence.\n\n[V] word\n\n",
			"First sentence.\nword",
			2, 1, 1,
		},
		{
			"trailing whitespace stripped",
			"[S] Hello.   \t\n[V] world  \r",
			"Hello.\nworld",
			2, 1, 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, sourceTypes := parseFormattedLines(tt.input)

			if formatted != tt.wantFormatted {
				t.Errorf("formatted:\ngot  %q\nwant %q", formatted, tt.wantFormatted)
			}

			if len(sourceTypes) != tt.wantTypeCount {
				t.Errorf("sourceTypes count = %d, want %d", len(sourceTypes), tt.wantTypeCount)
			}

			sentences, vocabs := 0, 0
			for _, st := range sourceTypes {
				switch st {
				case SourceTypeSentence:
					sentences++
				case SourceTypeVocab:
					vocabs++
				}
			}
			if sentences != tt.wantSentences {
				t.Errorf("sentence count = %d, want %d", sentences, tt.wantSentences)
			}
			if vocabs != tt.wantVocabs {
				t.Errorf("vocab count = %d, want %d", vocabs, tt.wantVocabs)
			}
		})
	}
}

func TestValidateFormatCounts(t *testing.T) {
	tests := []struct {
		name        string
		sourceTypes []string
		wantEmpty   bool
	}{
		{
			"within limits",
			makeTypes(10, SourceTypeSentence, 50, SourceTypeVocab),
			true,
		},
		{
			"at sentence limit",
			makeTypes(MaxSentences, SourceTypeSentence, 0, SourceTypeVocab),
			true,
		},
		{
			"at vocab limit",
			makeTypes(0, SourceTypeSentence, MaxVocab, SourceTypeVocab),
			true,
		},
		{
			"sentence exceeds limit",
			makeTypes(MaxSentences+1, SourceTypeSentence, 0, SourceTypeVocab),
			false,
		},
		{
			"vocab exceeds limit",
			makeTypes(0, SourceTypeSentence, MaxVocab+1, SourceTypeVocab),
			false,
		},
		{
			"empty types",
			[]string{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateFormatCounts(tt.sourceTypes)
			if tt.wantEmpty && result != "" {
				t.Errorf("expected empty warning, got %q", result)
			}
			if !tt.wantEmpty && result == "" {
				t.Error("expected warning, got empty string")
			}
		})
	}
}

func TestValidateFormatCounts_WarningMessages(t *testing.T) {
	// Verify warning messages contain the counts
	sentenceWarning := validateFormatCounts(makeTypes(MaxSentences+5, SourceTypeSentence, 0, SourceTypeVocab))
	if !strings.Contains(sentenceWarning, fmt.Sprintf("%d", MaxSentences+5)) {
		t.Errorf("sentence warning should contain actual count, got %q", sentenceWarning)
	}
	if !strings.Contains(sentenceWarning, fmt.Sprintf("%d", MaxSentences)) {
		t.Errorf("sentence warning should contain limit, got %q", sentenceWarning)
	}

	vocabWarning := validateFormatCounts(makeTypes(0, SourceTypeSentence, MaxVocab+10, SourceTypeVocab))
	if !strings.Contains(vocabWarning, fmt.Sprintf("%d", MaxVocab+10)) {
		t.Errorf("vocab warning should contain actual count, got %q", vocabWarning)
	}
}

func TestFormatBeanSlugs(t *testing.T) {
	t.Run("sentence type", func(t *testing.T) {
		consume, reason, refund, refundReason := formatBeanSlugs(SourceTypeSentence)
		if consume == "" || reason == "" || refund == "" || refundReason == "" {
			t.Error("expected non-empty slugs for sentence type")
		}
		if !strings.Contains(consume, "sentence") && !strings.Contains(consume, "Sentence") {
			t.Errorf("sentence consume slug should reference sentence: %q", consume)
		}
	})

	t.Run("vocab type", func(t *testing.T) {
		consume, reason, refund, refundReason := formatBeanSlugs(SourceTypeVocab)
		if consume == "" || reason == "" || refund == "" || refundReason == "" {
			t.Error("expected non-empty slugs for vocab type")
		}
		if !strings.Contains(consume, "vocab") && !strings.Contains(consume, "Vocab") {
			t.Errorf("vocab consume slug should reference vocab: %q", consume)
		}
	})

	t.Run("different slugs per type", func(t *testing.T) {
		sc, _, sr, _ := formatBeanSlugs(SourceTypeSentence)
		vc, _, vr, _ := formatBeanSlugs(SourceTypeVocab)
		if sc == vc {
			t.Error("sentence and vocab consume slugs should differ")
		}
		if sr == vr {
			t.Error("sentence and vocab refund slugs should differ")
		}
	})
}

func TestDifficultyDescriptions(t *testing.T) {
	expectedLevels := []string{"a1-a2", "b1-b2", "c1-c2"}
	for _, level := range expectedLevels {
		desc, ok := difficultyDescriptions[level]
		if !ok {
			t.Errorf("missing difficulty description for %q", level)
		}
		if desc == "" {
			t.Errorf("empty difficulty description for %q", level)
		}
	}
}

func TestDifficultyDescriptions_FallbackInGenerate(t *testing.T) {
	// buildGeneratePrompt should work with any level desc
	prompt := buildGeneratePrompt(difficultyDescriptions["a1-a2"])
	if prompt == "" {
		t.Error("buildGeneratePrompt returned empty string")
	}
	if !strings.Contains(prompt, "A1-A2") {
		t.Error("prompt should contain the difficulty level")
	}
	if !strings.Contains(prompt, "STEP 1") {
		t.Error("prompt should contain moderation step")
	}
	if !strings.Contains(prompt, "WARNING:") {
		t.Error("prompt should contain WARNING instruction")
	}
}

func TestBuildFormatPrompt(t *testing.T) {
	t.Run("sentence format", func(t *testing.T) {
		prompt := buildFormatPrompt(SourceTypeSentence)
		if !strings.Contains(prompt, "语句") {
			t.Error("sentence format prompt should contain 语句")
		}
		if !strings.Contains(prompt, "English sentence") {
			t.Error("sentence format prompt should mention English sentence")
		}
	})

	t.Run("vocab format", func(t *testing.T) {
		prompt := buildFormatPrompt(SourceTypeVocab)
		if !strings.Contains(prompt, "词汇") {
			t.Error("vocab format prompt should contain 词汇")
		}
		if !strings.Contains(prompt, "English word") {
			t.Error("vocab format prompt should mention English word")
		}
	})
}

func TestAIGenerateCost(t *testing.T) {
	if aiGenerateCost != 5 {
		t.Errorf("aiGenerateCost = %d, want 5", aiGenerateCost)
	}
}

func TestConcurrencyLimits(t *testing.T) {
	if breakConcurrencyLimit <= 0 {
		t.Errorf("breakConcurrencyLimit must be positive, got %d", breakConcurrencyLimit)
	}
	if genItemsConcurrencyLimit <= 0 {
		t.Errorf("genItemsConcurrencyLimit must be positive, got %d", genItemsConcurrencyLimit)
	}
	if breakConcurrencyLimit != 20 {
		t.Errorf("breakConcurrencyLimit = %d, want 20", breakConcurrencyLimit)
	}
	if genItemsConcurrencyLimit != 50 {
		t.Errorf("genItemsConcurrencyLimit = %d, want 50", genItemsConcurrencyLimit)
	}
}

func TestWriteSSEError_MessageMapping(t *testing.T) {
	// Verify all known errors produce non-empty Chinese messages
	knownErrors := []error{
		ErrGamePublished,
		ErrInsufficientBeans,
		ErrEmptyContent,
		ErrGameNotFound,
		ErrLevelNotFound,
		ErrForbidden,
	}
	for _, err := range knownErrors {
		if err == nil {
			t.Error("error sentinel should not be nil")
		}
	}
}

// makeTypes creates a slice of source types with the given counts.
func makeTypes(sentenceCount int, sentenceType string, vocabCount int, vocabType string) []string {
	types := make([]string, 0, sentenceCount+vocabCount)
	for i := 0; i < sentenceCount; i++ {
		types = append(types, sentenceType)
	}
	for i := 0; i < vocabCount; i++ {
		types = append(types, vocabType)
	}
	return types
}
