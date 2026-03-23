# Invite Stats — Wire Up Real Data

## Goal

Replace hardcoded mock stats in the invite page with values computed from the existing `referrals` prop.

## Approach

Compute all four stats client-side from the `Referral[]` array already passed to `InviteContent`. No new DB queries or service changes needed.

## Stats Definitions

| Card | Value | Sublabel |
|------|-------|----------|
| 累计获得推广佣金 | Sum of `rewardAmount` where status = "rewarded", formatted as `¥ X.XX` | — |
| 好友总数 | `referrals.length` | "本月新增 {newThisMonth} 位好友" |
| 好友已注册待验证 | Count where status = "pending" | — |
| 邀请成功转化比例 | `(paid + rewarded) / total * 100`, formatted as `X%` (0 referrals → `0%`) | — |

## Files

1. **New:** `src/features/web/invite/helpers/invite-stats.helper.ts` — pure function `computeInviteStats(referrals)` returning `{ totalReward, totalFriends, newThisMonth, pendingCount, conversionRate }`
2. **Edit:** `src/features/web/invite/components/invite-content.tsx` — remove hardcoded `stats` array, call helper, build stats dynamically
