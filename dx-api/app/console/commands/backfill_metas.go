package commands

import (
	"dx-api/app/consts"
)

// deriveSourceType maps a content_items.content_type to the corresponding
// content_metas.source_type per the backfill rule:
//   sentence → sentence (complete sentence)
//   word, phrase, block → vocab (all non-complete units)
func deriveSourceType(contentType string) string {
	if contentType == consts.ContentTypeSentence {
		return consts.SourceTypeSentence
	}
	return consts.SourceTypeVocab
}
