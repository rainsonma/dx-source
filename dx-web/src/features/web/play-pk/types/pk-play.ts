export type PkWinner = {
  user_id: string;
  user_name: string;
  score: number;
};

export type PkLevelCompleteEvent = {
  game_level_id: string;
  winner: PkWinner;
  participants: PkWinner[];
};

export type PkForceEndEvent = {
  pk_id: string;
};

export type PkNextLevelEvent = {
  pk_id: string;
  game_id: string;
  level_id: string;
  level_name: string;
  degree: string;
  pattern: string | null;
};

export type PkPlayerCompleteEvent = {
  user_id: string;
  user_name: string;
  game_level_id: string;
};

export type PkPlayerActionEvent = {
  user_id: string;
  user_name: string;
  action: "score" | "skip" | "combo";
  combo_streak?: number;
};

export type PkTimeoutWarningEvent = {
  countdown: number;
};

export type PkTimeoutEvent = {
  game_level_id: string;
};
