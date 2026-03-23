import type { UserGrade } from "@/consts/user-grade";

export type UserProfile = {
  id: string;
  username: string;
  nickname: string | null;
  email: string | null;
  grade: UserGrade;
  exp: number;
  beans: number;
  currentPlayStreak: number;
  lastReadNoticeAt: string | null;
  avatarUrl: string | null;
};
