export const GAME_MODES = {
  LSRW: "lsrw",
  VOCAB_BATTLE: "vocab-battle",
  VOCAB_MATCH: "vocab-match",
  VOCAB_ELIMINATION: "vocab-elimination",
  LISTENING_CHALLENGE: "listening-challenge",
} as const;

export type GameMode = (typeof GAME_MODES)[keyof typeof GAME_MODES];

export const GAME_MODE_LABELS: Record<GameMode, string> = {
  "lsrw": "听说读写",
  "vocab-battle": "词汇对轰",
  "vocab-match": "词汇配对",
  "vocab-elimination": "消消乐",
  "listening-challenge": "听力闯关",
};

export const GAME_MODE_OPTIONS: { value: GameMode; label: string }[] =
  Object.entries(GAME_MODE_LABELS).map(([value, label]) => ({
    value: value as GameMode,
    label,
  }));
