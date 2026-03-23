export const BEAN_SLUGS = {
  MEMBERSHIP_GRANT: "membership-grant",
  MONTHLY_RESET_DEBIT: "monthly-reset-debit",
  MONTHLY_RESET_CREDIT: "monthly-reset-credit",
  AI_GENERATE_CONSUME: "ai-generate-consume",
  AI_GENERATE_REFUND: "ai-generate-refund",
  AI_FORMAT_SENTENCE_CONSUME: "ai-format-sentence-consume",
  AI_FORMAT_SENTENCE_REFUND: "ai-format-sentence-refund",
  AI_FORMAT_VOCAB_CONSUME: "ai-format-vocab-consume",
  AI_FORMAT_VOCAB_REFUND: "ai-format-vocab-refund",
  AI_BREAK_CONSUME: "ai-break-consume",
  AI_BREAK_REFUND: "ai-break-refund",
  AI_GEN_ITEMS_CONSUME: "ai-gen-items-consume",
  AI_GEN_ITEMS_REFUND: "ai-gen-items-refund",
} as const;

export type BeanSlug = (typeof BEAN_SLUGS)[keyof typeof BEAN_SLUGS];

export const BEAN_SLUG_LABELS: Record<BeanSlug, string> = {
  "membership-grant": "会员赠送",
  "monthly-reset-debit": "月度清零",
  "monthly-reset-credit": "月度续发",
  "ai-generate-consume": "AI 生成消耗",
  "ai-generate-refund": "AI 生成失败退还",
  "ai-format-sentence-consume": "语句格式化消耗",
  "ai-format-sentence-refund": "语句格式化失败退还",
  "ai-format-vocab-consume": "词汇格式化消耗",
  "ai-format-vocab-refund": "词汇格式化失败退还",
  "ai-break-consume": "分解消耗",
  "ai-break-refund": "分解失败退还",
  "ai-gen-items-consume": "生成消耗",
  "ai-gen-items-refund": "生成失败退还",
};
