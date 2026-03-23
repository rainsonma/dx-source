export const GAME_STATUSES = {
  DRAFT: "draft",
  PUBLISHED: "published",
  WITHDRAW: "withdraw",
} as const;

export type GameStatus = (typeof GAME_STATUSES)[keyof typeof GAME_STATUSES];

export const GAME_STATUS_LABELS: Record<GameStatus, string> = {
  draft: "草稿",
  published: "已发布",
  withdraw: "已撤回",
};

export const GAME_STATUS_OPTIONS: { value: GameStatus; label: string }[] =
  Object.entries(GAME_STATUS_LABELS).map(([value, label]) => ({
    value: value as GameStatus,
    label,
  }));
