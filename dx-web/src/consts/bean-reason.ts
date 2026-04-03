export const BEAN_REASONS = {
  MEMBERSHIP_GRANT: "会员购买赠送",
  MONTHLY_RESET_DEBIT: "月度未使用能量豆清零",
  MONTHLY_RESET_CREDIT: "月度能量豆续发",
  AI_GENERATE_CONSUME: "AI 生成消耗",
  AI_GENERATE_REFUND: "AI 生成失败退还",
  AI_FORMAT_SENTENCE_CONSUME: "语句格式化消耗",
  AI_FORMAT_SENTENCE_REFUND: "语句格式化失败退还",
  AI_FORMAT_VOCAB_CONSUME: "词汇格式化消耗",
  AI_FORMAT_VOCAB_REFUND: "词汇格式化失败退还",
  AI_BREAK_CONSUME: "分解消耗",
  AI_BREAK_REFUND: "分解失败退还",
  AI_GEN_ITEMS_CONSUME: "生成消耗",
  AI_GEN_ITEMS_REFUND: "生成失败退还",
  PURCHASE_GRANT: "能量豆充值",
} as const;

export type BeanReason = (typeof BEAN_REASONS)[keyof typeof BEAN_REASONS];
