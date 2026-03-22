package consts

// Content type values.
const (
	ContentTypeWord     = "word"
	ContentTypeBlock    = "block"
	ContentTypePhrase   = "phrase"
	ContentTypeSentence = "sentence"
)

// ContentTypeLabels maps each content type to its Chinese label.
var ContentTypeLabels = map[string]string{
	ContentTypeWord:     "单词",
	ContentTypeBlock:    "组合",
	ContentTypePhrase:   "短语",
	ContentTypeSentence: "语句",
}
