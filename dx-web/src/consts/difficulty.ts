export const DIFFICULTY_OPTIONS = [
  { value: "a1-a2", label: "初级 (A1-A2)" },
  { value: "b1-b2", label: "中级 (B1-B2)" },
  { value: "c1-c2", label: "高级 (C1-C2)" },
] as const;

export type Difficulty = (typeof DIFFICULTY_OPTIONS)[number]["value"];
