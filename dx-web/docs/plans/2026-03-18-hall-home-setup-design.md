# Hall Home Page Setup Design

Date: 2026-03-18

## Overview

Wire real data to the hall dashboard home page stats row, remove the static mini leaderboard, clean up the daily challenge card, and fix the ad cards layout with proper links.

## Changes

### 1. StatsRow — Wire Real Data

**Data sources (all existing):**
- `user.exp` — from `getUserProfile` (add `currentPlayStreak` to select)
- `user.currentPlayStreak` — from `getUserProfile`
- `getUserMasterStats(userId).total` — from `user-master.query.ts`
- `getUserReviewStats(userId).pending` — from `user-review.query.ts`

**New service:** `src/features/web/hall/services/hall.service.ts` — fetches all dashboard stats in parallel.

**Stats cards:**

| Card | Icon | Value | Subtitle |
|------|------|-------|----------|
| 总经验值 | Zap (teal) | `exp` formatted | — |
| 连续学习 | Flame (orange) | `currentPlayStreak` 天 | — |
| 已掌握词汇 | BookOpen (violet) | `masterStats.total` | 本周新增 N 个 |
| 待复习词汇 | Target (teal) | `reviewStats.pending` | 今日待复习 |

`StatsRow` becomes a props-driven component; page fetches and passes data.

### 2. Remove MiniLeaderboard

Delete `mini-leaderboard.tsx`. Remove import and usage from page. Right column keeps only `DailyChallengeCard`.

### 3. DailyChallengeCard

- Remove `Clock` icon import and the "剩余 14 小时 35 分" countdown block
- Replace `<button>` with `<Link href="/hall/games">`

### 4. AdCardsRow

- Remove both `130x130` placeholder square divs
- Move button to right side of each card (vertically centered)
- "升级 Pro 会员" button → `<Link href="/hall/redeem">`
- "邀请好友一起学" button → `<Link href="/hall/invite">`

## Files Touched

- `src/models/user/user.query.ts` — add `currentPlayStreak` to select
- `src/features/web/hall/services/hall.service.ts` — new, fetches dashboard stats
- `src/features/web/hall/components/stats-row.tsx` — accept props, render real data
- `src/features/web/hall/components/daily-challenge-card.tsx` — Link + remove countdown
- `src/features/web/hall/components/ad-cards-row.tsx` — layout change + links
- `src/features/web/hall/components/mini-leaderboard.tsx` — delete
- `src/app/(web)/hall/(main)/(home)/page.tsx` — fetch stats, pass props, remove leaderboard
