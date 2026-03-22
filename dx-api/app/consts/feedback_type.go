package consts

// Feedback type values.
const (
	FeedbackTypeFeature = "feature"
	FeedbackTypeContent = "content"
	FeedbackTypeUX      = "ux"
	FeedbackTypeBug     = "bug"
	FeedbackTypeOther   = "other"
)

// FeedbackTypeLabels maps each feedback type to its Chinese label.
var FeedbackTypeLabels = map[string]string{
	FeedbackTypeFeature: "功能建议",
	FeedbackTypeContent: "内容纠错",
	FeedbackTypeUX:      "界面体验",
	FeedbackTypeBug:     "Bug 报告",
	FeedbackTypeOther:   "其它",
}
