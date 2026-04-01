export const GAME_DEGREES = {
  BEGINNER: "beginner",
  INTERMEDIATE: "intermediate",
  ADVANCED: "advanced",
} as const;

export type GameDegree = (typeof GAME_DEGREES)[keyof typeof GAME_DEGREES];

export const GAME_DEGREE_LABELS: Record<GameDegree, string> = {
  beginner: "初级",
  intermediate: "中级",
  advanced: "高级",
};

export const DEGREE_CONTENT_TYPES: Record<GameDegree, string[] | null> = {
  beginner: null,
  intermediate: ["block", "phrase", "sentence"],
  advanced: ["sentence"],
};
