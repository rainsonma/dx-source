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
