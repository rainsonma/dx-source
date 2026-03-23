export const SOURCE_TYPES = {
  SENTENCE: "sentence",
  VOCAB: "vocab",
} as const;

export type SourceType = (typeof SOURCE_TYPES)[keyof typeof SOURCE_TYPES];

export const SOURCE_TYPE_LABELS: Record<SourceType, string> = {
  sentence: "语句",
  vocab: "词汇",
};
