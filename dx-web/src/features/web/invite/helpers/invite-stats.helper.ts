import { REFERRAL_STATUSES } from "@/consts/referral-status";

type ReferralForStats = {
  status: string;
  rewardAmount: number | string | null;
  createdAt: Date | string;
};

export type InviteStats = {
  totalReward: string;
  totalFriends: number;
  newThisMonth: number;
  pendingCount: number;
  conversionRate: string;
};

/** Compute invite page stats from the referrals array */
export function computeInviteStats(referrals: ReferralForStats[]): InviteStats {
  const now = new Date();
  const currentYear = now.getFullYear();
  const currentMonth = now.getMonth();

  let rewardSum = 0;
  let pendingCount = 0;
  let convertedCount = 0;
  let newThisMonth = 0;

  for (const r of referrals) {
    if (r.status === REFERRAL_STATUSES.REWARDED) {
      rewardSum += Number(r.rewardAmount) || 0;
    }

    if (r.status === REFERRAL_STATUSES.PENDING) {
      pendingCount++;
    }

    if (
      r.status === REFERRAL_STATUSES.PAID ||
      r.status === REFERRAL_STATUSES.REWARDED
    ) {
      convertedCount++;
    }

    const created = new Date(r.createdAt);
    if (
      created.getFullYear() === currentYear &&
      created.getMonth() === currentMonth
    ) {
      newThisMonth++;
    }
  }

  const total = referrals.length;
  const rate = total > 0 ? Math.round((convertedCount / total) * 100) : 0;

  return {
    totalReward: `¥ ${rewardSum.toFixed(2)}`,
    totalFriends: total,
    newThisMonth,
    pendingCount,
    conversionRate: `${rate}%`,
  };
}
