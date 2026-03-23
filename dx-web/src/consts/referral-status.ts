export const REFERRAL_STATUSES = {
  PENDING: "pending",
  PAID: "paid",
  REWARDED: "rewarded",
} as const;

export type ReferralStatus =
  (typeof REFERRAL_STATUSES)[keyof typeof REFERRAL_STATUSES];

export const REFERRAL_STATUS_LABELS: Record<ReferralStatus, string> = {
  pending: "待验证",
  paid: "已付费",
  rewarded: "已发放",
};
