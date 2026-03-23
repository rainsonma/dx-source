# Invite Stats Real Data Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace hardcoded mock stats in the invite page with values computed from the existing referrals data.

**Architecture:** A pure helper function computes all 4 stats from the `Referral[]` array already passed to the component. The component builds its stats config dynamically using the helper output. No DB, service, or API changes needed.

**Tech Stack:** TypeScript, React (client component)

---

### Task 1: Create invite stats helper

**Files:**
- Create: `src/features/web/invite/helpers/invite-stats.helper.ts`

**Step 1: Write the helper**

```typescript
import { REFERRAL_STATUSES } from "@/consts/referral-status";

type ReferralForStats = {
  status: string;
  rewardAmount: unknown;
  createdAt: Date;
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
```

**Step 2: Verify no build errors**

Run: `npx tsc --noEmit --pretty`
Expected: No errors related to the new file

---

### Task 2: Wire stats into invite-content component

**Files:**
- Modify: `src/features/web/invite/components/invite-content.tsx`

**Step 1: Update the component**

Changes:
1. Add import for `computeInviteStats`
2. Remove the hardcoded `stats` constant (lines 28-33)
3. Inside the component function, call `computeInviteStats(referrals)` and build the stats array dynamically

Replace the hardcoded `stats` array with this inside the component:

```typescript
const inviteStats = computeInviteStats(referrals);

const stats = [
  { icon: DollarSign, iconBg: "bg-teal-100", iconColor: "text-teal-600", value: inviteStats.totalReward, label: "累计获得推广佣金" },
  { icon: Users, iconBg: "bg-blue-100", iconColor: "text-blue-500", value: String(inviteStats.totalFriends), label: `本月新增 ${inviteStats.newThisMonth} 位好友` },
  { icon: UserCheck, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(inviteStats.pendingCount), label: "好友已注册待验证" },
  { icon: TrendingUp, iconBg: "bg-purple-100", iconColor: "text-purple-600", value: inviteStats.conversionRate, label: "邀请成功转化比例" },
];
```

**Step 2: Verify build passes**

Run: `npx tsc --noEmit --pretty`
Expected: No errors

**Step 3: Verify dev server renders correctly**

Run: `npm run dev` and navigate to `/hall/invite`
Expected: Stats cards show computed values (likely all zeros/0% if no referrals exist yet)

**Step 4: Commit**

```
feat: wire invite page stats to real referral data
```
