package constants

// Bean slug values identify specific bean transaction types.
const (
	BeanSlugMembershipGrant         = "membership-grant"
	BeanSlugMonthlyResetDebit       = "monthly-reset-debit"
	BeanSlugMonthlyResetCredit      = "monthly-reset-credit"
	BeanSlugAIGenerateConsume       = "ai-generate-consume"
	BeanSlugAIGenerateRefund        = "ai-generate-refund"
	BeanSlugAIFormatSentenceConsume = "ai-format-sentence-consume"
	BeanSlugAIFormatSentenceRefund  = "ai-format-sentence-refund"
	BeanSlugAIFormatVocabConsume    = "ai-format-vocab-consume"
	BeanSlugAIFormatVocabRefund     = "ai-format-vocab-refund"
	BeanSlugAIBreakConsume          = "ai-break-consume"
	BeanSlugAIBreakRefund           = "ai-break-refund"
	BeanSlugAIGenItemsConsume       = "ai-gen-items-consume"
	BeanSlugAIGenItemsRefund        = "ai-gen-items-refund"
)

// BeanSlugLabels maps each bean slug to its Chinese label.
var BeanSlugLabels = map[string]string{
	BeanSlugMembershipGrant:         "会员赠送",
	BeanSlugMonthlyResetDebit:       "月度清零",
	BeanSlugMonthlyResetCredit:      "月度续发",
	BeanSlugAIGenerateConsume:       "AI 生成消耗",
	BeanSlugAIGenerateRefund:        "AI 生成失败退还",
	BeanSlugAIFormatSentenceConsume: "语句格式化消耗",
	BeanSlugAIFormatSentenceRefund:  "语句格式化失败退还",
	BeanSlugAIFormatVocabConsume:    "词汇格式化消耗",
	BeanSlugAIFormatVocabRefund:     "词汇格式化失败退还",
	BeanSlugAIBreakConsume:          "分解消耗",
	BeanSlugAIBreakRefund:           "分解失败退还",
	BeanSlugAIGenItemsConsume:       "生成消耗",
	BeanSlugAIGenItemsRefund:        "生成失败退还",
}
