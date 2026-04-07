package consts

// Bean reason values describe why a bean transaction occurred.
const (
	BeanReasonMembershipGrant         = "会员购买赠送"
	BeanReasonMonthlyResetDebit       = "月度未使用能量豆清零"
	BeanReasonMonthlyResetCredit      = "月度能量豆续发"
	BeanReasonAIGenerateConsume       = "AI 生成消耗"
	BeanReasonAIGenerateRefund        = "AI 生成失败退还"
	BeanReasonAIFormatSentenceConsume = "语句格式化消耗"
	BeanReasonAIFormatSentenceRefund  = "语句格式化失败退还"
	BeanReasonAIFormatVocabConsume    = "词汇格式化消耗"
	BeanReasonAIFormatVocabRefund     = "词汇格式化失败退还"
	BeanReasonAIBreakConsume          = "分解消耗"
	BeanReasonAIBreakRefund           = "分解失败退还"
	BeanReasonAIGenItemsConsume       = "生成消耗"
	BeanReasonAIGenItemsRefund        = "生成失败退还"
	BeanReasonAIVocabGenerateConsume  = "词汇 AI 生成消耗"
	BeanReasonAIVocabGenerateRefund   = "词汇 AI 生成失败退还"
	BeanReasonAIVocabFormatConsume    = "词汇格式化消耗"
	BeanReasonAIVocabFormatRefund     = "词汇格式化失败退还"
	BeanReasonAIVocabBreakConsume     = "词汇分解消耗"
	BeanReasonAIVocabBreakRefund      = "词汇分解失败退还"
	BeanReasonAIVocabGenItemsConsume  = "词汇生成消耗"
	BeanReasonAIVocabGenItemsRefund   = "词汇生成失败退还"
	BeanReasonSeederGrant             = "种子数据赠送能量豆"
	BeanReasonPurchaseGrant           = "能量豆充值"
)
