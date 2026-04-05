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
  mode: "group_solo" | "group_team";
  winner: SoloWinner | TeamWinner;
  participants: SoloWinner[];
  teams?: TeamWinner[];
};

export type GroupForceEndEvent = {
  results: GroupLevelCompleteEvent[];
};

export type ParticipantMember = {
  user_id: string;
  user_name: string;
};

export type SoloParticipants = {
  mode: "group_solo";
  members: ParticipantMember[];
};

export type TeamParticipants = {
  mode: "group_team";
  teams: {
    subgroup_id: string;
    subgroup_name: string;
    members: ParticipantMember[];
  }[];
};

export type Participants = SoloParticipants | TeamParticipants;

export type GroupPlayerCompleteEvent = {
  user_id: string;
  user_name: string;
  game_level_id: string;
  score: number;
  participants: { user_id: string; user_name: string; score: number }[];
};

export type GroupPlayerActionEvent = {
  user_id: string;
  user_name: string;
  action: "score" | "combo";
  combo_streak?: number;
};
