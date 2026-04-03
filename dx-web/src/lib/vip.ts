import type { UserGrade } from "@/consts/user-grade";

/**
 * Check whether the user has an active VIP membership.
 * VIP = lifetime grade, or paid grade with vipDueAt in the future.
 */
export function isVipActive(grade: UserGrade, vipDueAt: string | null): boolean {
  if (grade === "lifetime") return true;
  if (grade === "free") return false;
  if (!vipDueAt) return false;
  return new Date(vipDueAt) > new Date();
}
