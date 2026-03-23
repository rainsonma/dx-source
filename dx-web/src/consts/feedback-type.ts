export const FEEDBACK_TYPES = {
  FEATURE: "feature",
  CONTENT: "content",
  UX: "ux",
  BUG: "bug",
  OTHER: "other",
} as const;

export type FeedbackType = (typeof FEEDBACK_TYPES)[keyof typeof FEEDBACK_TYPES];

export const FEEDBACK_TYPE_LABELS: Record<FeedbackType, string> = {
  feature: "功能建议",
  content: "内容纠错",
  ux: "界面体验",
  bug: "Bug 报告",
  other: "其它",
};
