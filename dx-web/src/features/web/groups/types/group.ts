export type Group = {
  id: string;
  name: string;
  description: string | null;
  owner_id: string;
  owner_name: string;
  member_count: number;
  invite_code: string;
  is_member: boolean;
  has_applied: boolean;
  is_owner: boolean;
  created_at: string;
};

export type GroupDetail = Group & {
  is_active: boolean;
  current_game_id: string | null;
  game_mode: string | null;
  current_game_name: string | null;
  invite_qrcode_url: string | null;
  level_time_limit: number;
  is_playing: boolean;
  start_game_level_id: string | null;
  start_game_level_name: string | null;
};

export type RoomMemberEvent = {
  user_id: string;
};

export type RoomMember = {
  user_id: string;
  user_name: string;
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

export type GroupGameStartEvent = {
  game_group_id: string;
  game_id: string;
  game_name: string;
  game_mode: "group_solo" | "group_team";
  degree: string;
  pattern: string | null;
  level_time_limit: number;
  level_id: string | null;
  level_name: string;
  participants: Participants;
};

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
};

export type GroupForceEndEvent = {
  results: GroupLevelCompleteEvent[];
};

export type GroupGameSearchItem = {
  id: string;
  name: string;
  mode: string;
  category_name: string | null;
};

export type GroupMember = {
  id: string;
  user_id: string;
  user_name: string;
  is_owner: boolean;
  created_at: string;
};

export type Subgroup = {
  id: string;
  name: string;
  description: string | null;
  member_count: number;
  order: number;
};

export type SubgroupMember = {
  id: string;
  user_id: string;
  user_name: string;
};

export type GroupApplication = {
  id: string;
  user_id: string;
  user_name: string;
  status: string;
  created_at: string;
};
