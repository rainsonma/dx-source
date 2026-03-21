package constants

// Source type values.
const (
	SourceTypeSentence = "sentence"
	SourceTypeVocab    = "vocab"
)

// SourceTypeLabels maps each source type to its Chinese label.
var SourceTypeLabels = map[string]string{
	SourceTypeSentence: "语句",
	SourceTypeVocab:    "词汇",
}
