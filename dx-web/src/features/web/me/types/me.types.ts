import type { UserGrade } from "@/consts/user-grade";

/** Full user profile data for the personal center page */
export type MeProfile = {
  id: string;
  username: string;
  nickname: string | null;
  email: string | null;
  phone: string | null;
  city: string | null;
  introduction: string | null;
  grade: UserGrade;
  vipDueAt: string | null;
  beans: number;
  exp: number;
  level: number;
  currentPlayStreak: number;
  maxPlayStreak: number;
  lastPlayedAt: string | null;
  inviteCode: string;
  createdAt: string;
  avatarUrl: string | null;
};

/** Raw profile shape returned by the Go API (snake_case) */
export type ApiProfileData = {
  id: string;
  username: string;
  nickname: string | null;
  email: string | null;
  phone: string | null;
  avatar_id: string | null;
  avatar_url: string | null;
  city: string | null;
  introduction: string | null;
  grade: string;
  is_active: boolean;
  beans: number;
  granted_beans: number;
  exp: number;
  level: number;
  invite_code: string;
  current_play_streak: number;
  max_play_streak: number;
  last_played_at: string | null;
  vip_due_at: string | null;
  created_at: string;
  updated_at: string;
};

/** Map Go API snake_case profile to frontend camelCase MeProfile */
export function toMeProfile(raw: ApiProfileData): MeProfile {
  return {
    id: raw.id,
    username: raw.username,
    nickname: raw.nickname,
    email: raw.email,
    phone: raw.phone,
    city: raw.city,
    introduction: raw.introduction,
    grade: raw.grade as UserGrade,
    vipDueAt: raw.vip_due_at,
    beans: raw.beans,
    exp: raw.exp,
    level: raw.level,
    currentPlayStreak: raw.current_play_streak,
    maxPlayStreak: raw.max_play_streak,
    lastPlayedAt: raw.last_played_at,
    inviteCode: raw.invite_code,
    createdAt: raw.created_at,
    avatarUrl: raw.avatar_url ?? null,
  };
}
