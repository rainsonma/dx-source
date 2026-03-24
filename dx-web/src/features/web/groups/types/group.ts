export type Group = {
  id: string;
  name: string;
  description: string | null;
  owner_id: string;
  owner_name: string;
  member_count: number;
  invite_code: string;
  is_member: boolean;
  is_owner: boolean;
  created_at: string;
};

export type GroupDetail = Group & {
  is_active: boolean;
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
