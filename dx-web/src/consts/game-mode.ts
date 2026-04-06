export const GAME_MODES = {
  WORD_SENTENCE: "word-sentence",
  VOCAB_BATTLE: "vocab-battle",
  VOCAB_MATCH: "vocab-match",
  VOCAB_ELIMINATION: "vocab-elimination",
} as const;

export type GameMode = (typeof GAME_MODES)[keyof typeof GAME_MODES];

export const GAME_MODE_LABELS: Record<GameMode, string> = {
  "word-sentence": "连词成句",
  "vocab-battle": "词汇对轰",
  "vocab-match": "词汇配对",
  "vocab-elimination": "词汇消消乐",
};

export const GAME_MODE_OPTIONS: { value: GameMode; label: string }[] =
  Object.entries(GAME_MODE_LABELS).map(([value, label]) => ({
    value: value as GameMode,
    label,
  }));
