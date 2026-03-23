export const SOURCE_FROMS = {
  MANUAL: "manual",
  AI: "ai",
} as const;

export type SourceFrom = (typeof SOURCE_FROMS)[keyof typeof SOURCE_FROMS];

export const SOURCE_FROM_LABELS: Record<SourceFrom, string> = {
  manual: "手动添加",
  ai: "AI 生成",
};
