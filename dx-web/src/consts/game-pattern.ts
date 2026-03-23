export const GAME_PATTERNS = {
  LISTEN: "listen",
  SPEAK: "speak",
  READ: "read",
  WRITE: "write",
} as const;

export type GamePattern = (typeof GAME_PATTERNS)[keyof typeof GAME_PATTERNS];

export const GAME_PATTERN_LABELS: Record<GamePattern, string> = {
  listen: "听",
  speak: "说",
  read: "读",
  write: "写",
};

export const DEFAULT_GAME_PATTERN: GamePattern = GAME_PATTERNS.WRITE;
