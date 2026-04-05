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

export type PkPlayerCompleteEvent = {
  user_id: string;
  user_name: string;
  game_level_id: string;
  score: number;
};

export type PkPlayerActionEvent = {
  user_id: string;
  user_name: string;
  action: "score" | "combo";
  combo_streak?: number;
};

export type PkTimeoutWarningEvent = {
  countdown: number;
};

export type PkTimeoutEvent = {
  game_level_id: string;
};
