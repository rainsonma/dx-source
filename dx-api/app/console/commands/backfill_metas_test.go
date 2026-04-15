package commands

import (
	"testing"

	"dx-api/app/consts"
)

func TestDeriveSourceType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        string
	}{
		{"word is vocab", consts.ContentTypeWord, consts.SourceTypeVocab},
		{"phrase is vocab", consts.ContentTypePhrase, consts.SourceTypeVocab},
		{"block is vocab", consts.ContentTypeBlock, consts.SourceTypeVocab},
		{"sentence is sentence", consts.ContentTypeSentence, consts.SourceTypeSentence},
		{"unknown defaults to vocab", "unknown", consts.SourceTypeVocab},
		{"empty defaults to vocab", "", consts.SourceTypeVocab},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveSourceType(tt.contentType)
			if got != tt.want {
				t.Errorf("deriveSourceType(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}
