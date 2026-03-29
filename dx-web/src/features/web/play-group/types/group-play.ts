export type SoloWinner = {
  user_id: string;
  user_name: string;
  score: number;
};

export type TeamWinner = {
  subgroup_id: string;
  subgroup_name: string;
  total_score: number;
  members: { user_id: string; user_name: string; score: number }[];
};

export type GroupLevelCompleteEvent = {
  game_level_id: string;
  mode: "solo" | "team";
  winner: SoloWinner | TeamWinner;
  participants: SoloWinner[];
  teams?: TeamWinner[];
};

export type GroupForceEndEvent = {
  results: GroupLevelCompleteEvent[];
};
