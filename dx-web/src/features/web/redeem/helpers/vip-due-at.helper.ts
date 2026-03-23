import type { UserGrade } from "@/consts/user-grade";
import { USER_GRADE_MONTHS } from "@/consts/user-grade";

/**
 * Add N calendar months to a date, then subtract 1 day.
 * If the target month has fewer days, clamp to the last day.
 * Example: March 17 + 1 month = April 16
 * Example: Jan 31 + 1 month = Feb 28 (or 29)
 */
function addMonthsMinusOneDay(base: Date, months: number): Date {
  const year = base.getFullYear();
  const month = base.getMonth() + months;
  const day = base.getDate();

  // Create a date in the target month with the same day
  const target = new Date(year, month, day);

  // If the day overflowed (e.g., Jan 31 → Mar 3), clamp to last day of target month
  if (target.getMonth() !== ((base.getMonth() + months) % 12 + 12) % 12) {
    // Set to last day of the intended target month
    return new Date(year, month, 0);
  }

  // Subtract 1 day
  target.setDate(target.getDate() - 1);
  return target;
}

/**
 * Calculate the new vipDueAt date after redeeming a code.
 *
 * Rules:
 * - lifetime → return null (never expires)
 * - free/expired → base = today, add grade months
 * - not expired → base = current vipDueAt, add grade months
 */
export function calcVipDueAt(
  grade: UserGrade,
  currentVipDueAt: Date | null,
): Date | null {
  const months = USER_GRADE_MONTHS[grade];

  // Lifetime never expires
  if (months === null) return null;

  const now = new Date();
  const isExpired = !currentVipDueAt || currentVipDueAt < now;
  const base = isExpired ? now : currentVipDueAt;

  return addMonthsMinusOneDay(base, months);
}
