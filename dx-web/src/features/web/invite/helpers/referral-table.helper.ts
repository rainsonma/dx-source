import { REFERRAL_STATUSES } from "@/consts/referral-status";

/** Color palette for avatar initials, cycled by index */
export const AVATAR_COLORS = [
  { bg: "bg-blue-100", text: "text-blue-600" },
  { bg: "bg-purple-100", text: "text-purple-600" },
  { bg: "bg-teal-100", text: "text-teal-600" },
  { bg: "bg-amber-100", text: "text-amber-600" },
  { bg: "bg-red-100", text: "text-red-600" },
];

/** Get the display name for a referral's invitee */
export function getDisplayName(
  invitee: { nickname: string | null; username: string } | null
): string {
  if (!invitee) return "-";
  return invitee.nickname || invitee.username;
}

/** Mask an email address for privacy display */
export function maskEmail(email: string | null): string {
  if (!email) return "-";
  const [local, domain] = email.split("@");
  if (!domain) return email;
  const visible = local.slice(0, 3);
  return `${visible}***@${domain}`;
}

/** Format a date to YYYY-MM-DD string */
export function formatDate(date: Date): string {
  return new Date(date).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}

/** Format reward amount for display */
export function formatReward(amount: unknown, status: string): string {
  if (status === REFERRAL_STATUSES.PENDING) return "-";
  const num = Number(amount);
  if (!num) return "-";
  return `¥ ${num.toFixed(2)}`;
}

/** Get CSS classes for referral status badge */
export function getStatusClasses(status: string): string {
  if (status === REFERRAL_STATUSES.PENDING) {
    return "bg-amber-100 text-amber-700";
  }
  return "bg-teal-600/10 text-teal-600";
}
