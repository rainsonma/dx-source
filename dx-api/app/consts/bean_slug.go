package consts

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
	BeanSlugAIVocabGenerateConsume  = "ai-vocab-generate-consume"
	BeanSlugAIVocabGenerateRefund   = "ai-vocab-generate-refund"
	BeanSlugAIVocabFormatConsume    = "ai-vocab-format-consume"
	BeanSlugAIVocabFormatRefund     = "ai-vocab-format-refund"
	BeanSlugAIVocabBreakConsume     = "ai-vocab-break-consume"
	BeanSlugAIVocabBreakRefund      = "ai-vocab-break-refund"
	BeanSlugAIVocabGenItemsConsume  = "ai-vocab-gen-items-consume"
	BeanSlugAIVocabGenItemsRefund   = "ai-vocab-gen-items-refund"
	BeanSlugSeederGrant             = "seeder-grant"
	BeanSlugPurchaseGrant           = "purchase-grant"
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
	BeanSlugAIVocabGenerateConsume:  "词汇 AI 生成消耗",
	BeanSlugAIVocabGenerateRefund:   "词汇 AI 生成失败退还",
	BeanSlugAIVocabFormatConsume:    "词汇格式化消耗",
	BeanSlugAIVocabFormatRefund:     "词汇格式化失败退还",
	BeanSlugAIVocabBreakConsume:     "词汇分解消耗",
	BeanSlugAIVocabBreakRefund:      "词汇分解失败退还",
	BeanSlugAIVocabGenItemsConsume:  "词汇生成消耗",
	BeanSlugAIVocabGenItemsRefund:   "词汇生成失败退还",
	BeanSlugSeederGrant:             "种子数据赠送",
	BeanSlugPurchaseGrant:           "能量豆充值",
}
