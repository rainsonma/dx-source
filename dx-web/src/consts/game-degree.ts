export const GAME_DEGREES = {
  PRACTICE: "practice",
  BEGINNER: "beginner",
  INTERMEDIATE: "intermediate",
  ADVANCED: "advanced",
} as const;

export type GameDegree = (typeof GAME_DEGREES)[keyof typeof GAME_DEGREES];

export const GAME_DEGREE_LABELS: Record<GameDegree, string> = {
  practice: "练习",
  beginner: "初级",
  intermediate: "中级",
  advanced: "高级",
};

export const DEGREE_CONTENT_TYPES: Record<GameDegree, string[] | null> = {
  practice: null,
  beginner: null,
  intermediate: ["block", "phrase", "sentence"],
  advanced: ["sentence"],
};
