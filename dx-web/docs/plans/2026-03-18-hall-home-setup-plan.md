# Hall Home Page Setup — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire real data to the hall dashboard stats, remove mini leaderboard, fix daily challenge card, and update ad cards layout with links.

**Architecture:** The page already fetches `fetchUserProfile()`. We add `currentPlayStreak` to the profile select, create a `hall.service.ts` to fetch dashboard stats (master/review counts) in parallel, and pass everything as props. Components become data-driven instead of hardcoded.

**Tech Stack:** Next.js 16 App Router, Prisma, React server components, Lucide icons, TailwindCSS v4

**Design doc:** `docs/plans/2026-03-18-hall-home-setup-design.md`

---

### Task 1: Add `currentPlayStreak` to user profile query

**Files:**
- Modify: `src/models/user/user.query.ts:49-80`

**Step 1: Add field to select and return**

In `getUserProfile`, add `currentPlayStreak: true` to the select object (after line 58), and add `currentPlayStreak: user.currentPlayStreak` to the return object (after line 75).

```typescript
// In select (line 58 area):
exp: true,
currentPlayStreak: true,  // <-- add
inviteCode: true,

// In return (line 75 area):
exp: user.exp,
currentPlayStreak: user.currentPlayStreak,  // <-- add
inviteCode: user.inviteCode,
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`

**Step 3: Commit**

```
feat: add currentPlayStreak to user profile query
```

---

### Task 2: Create hall dashboard service

**Files:**
- Create: `src/features/web/hall/services/hall.service.ts`

**Step 1: Create the service**

```typescript
import "server-only";

import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { getUserMasterStats } from "@/models/user-master/user-master.query";
import { getUserReviewStats } from "@/models/user-review/user-review.query";

/** Fetch all stats needed for the hall dashboard home page */
export async function fetchDashboardStats() {
  const profile = await fetchUserProfile();
  if (!profile) return null;

  const [masterStats, reviewStats] = await Promise.all([
    getUserMasterStats(profile.id),
    getUserReviewStats(profile.id),
  ]);

  return { profile, masterStats, reviewStats };
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`

**Step 3: Commit**

```
feat: add hall dashboard service for stats fetching
```

---

### Task 3: Update StatsRow to accept real data via props

**Files:**
- Modify: `src/features/web/hall/components/stats-row.tsx`

**Step 1: Rewrite component to accept props**

Replace entire file. Remove hardcoded `stats` array. Accept props for the 4 values. Format numbers with `toLocaleString()`.

```typescript
import { Zap, Flame, BookOpen, Target } from "lucide-react";

interface StatsRowProps {
  exp: number;
  currentPlayStreak: number;
  masteredTotal: number;
  masteredThisWeek: number;
  reviewPending: number;
}

/** Dashboard stats row showing EXP, streak, mastered, and review counts */
export function StatsRow({
  exp,
  currentPlayStreak,
  masteredTotal,
  masteredThisWeek,
  reviewPending,
}: StatsRowProps) {
  const stats = [
    {
      icon: Zap,
      iconColor: "text-teal-600",
      label: "总经验值",
      value: exp.toLocaleString(),
      sub: `累计获得经验值`,
    },
    {
      icon: Flame,
      iconColor: "text-orange-500",
      label: "连续学习",
      value: `${currentPlayStreak} 天`,
      sub: "保持连续学习记录！",
    },
    {
      icon: BookOpen,
      iconColor: "text-violet-500",
      label: "已掌握词汇",
      value: masteredTotal.toLocaleString(),
      sub: `本周新增 ${masteredThisWeek} 个`,
    },
    {
      icon: Target,
      iconColor: "text-teal-600",
      label: "待复习词汇",
      value: reviewPending.toLocaleString(),
      sub: "今日待复习",
    },
  ];

  return (
    <div className="flex w-full gap-4">
      {stats.map((stat) => (
        <div
          key={stat.label}
          className="flex w-full flex-col gap-2 rounded-[14px] border border-slate-200 bg-white p-5"
        >
          <div className="flex items-center gap-2">
            <stat.icon className={`h-[18px] w-[18px] ${stat.iconColor}`} />
            <span className="text-[13px] font-medium text-slate-500">
              {stat.label}
            </span>
          </div>
          <span className="text-[28px] font-extrabold text-slate-900">
            {stat.value}
          </span>
          <span className="text-xs text-slate-400">{stat.sub}</span>
        </div>
      ))}
    </div>
  );
}
```

**Step 2: Verify build** (will fail until page is updated in Task 6 — that's OK)

---

### Task 4: Update DailyChallengeCard — remove countdown, add Link

**Files:**
- Modify: `src/features/web/hall/components/daily-challenge-card.tsx`

**Step 1: Rewrite component**

- Remove `Clock` from imports
- Remove the countdown `<div>` block (lines 19-22)
- Replace `<button>` with `<Link href="/hall/games">`

```typescript
import Link from "next/link";
import { Flame, Play } from "lucide-react";

export function DailyChallengeCard() {
  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] bg-teal-600 p-6">
      <div className="flex items-center gap-2">
        <Flame className="h-4 w-4 text-amber-300" />
        <span className="text-xs font-semibold text-teal-100">今日挑战</span>
      </div>
      <p className="text-base font-bold leading-[1.5] text-white">
        完成今天的听说读写
        <br />
        赢取双倍经验值！
      </p>
      <Link
        href="/hall/games"
        className="flex h-11 w-full items-center justify-center gap-2 rounded-[10px] bg-white text-[13px] font-semibold text-teal-700 hover:bg-white/90"
      >
        <Play className="h-4 w-4" />
        开始挑战
      </Link>
    </div>
  );
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`

**Step 3: Commit**

```
feat: link daily challenge to games page, remove countdown
```

---

### Task 5: Update AdCardsRow — remove squares, move buttons right, add links

**Files:**
- Modify: `src/features/web/hall/components/ad-cards-row.tsx`

**Step 1: Rewrite component**

- Add `Link` from `next/link`
- Remove both `130x130` placeholder `<div>`s
- Change layout so button is on the right side, vertically centered
- "升级 Pro 会员" links to `/hall/redeem`
- "邀请好友一起学" links to `/hall/invite`

```typescript
import Link from "next/link";
import { Crown, ArrowRight, Gift } from "lucide-react";

export function AdCardsRow() {
  return (
    <div className="flex h-[210px] w-full gap-4">
      {/* Upgrade Pro card */}
      <div className="flex w-full items-center justify-between overflow-hidden rounded-[14px] bg-gradient-to-b from-teal-600 to-teal-700 px-7 py-6">
        <div className="flex h-full flex-col justify-center gap-2">
          <span className="w-fit rounded-full bg-white/20 px-3 py-1 text-[11px] font-semibold text-white">
            限时优惠
          </span>
          <h3 className="text-[22px] font-extrabold text-white">
            升级 Pro 会员
          </h3>
          <p className="w-60 text-[13px] leading-[1.5] text-white/80">
            解锁无限词汇量、AI 对话练习和专属学习报告
          </p>
        </div>
        <Link
          href="/hall/redeem"
          className="flex shrink-0 items-center gap-1.5 rounded-lg bg-white px-5 py-2 text-[13px] font-semibold text-teal-700 hover:bg-white/90"
        >
          <Crown className="h-3.5 w-3.5" />
          立即升级
        </Link>
      </div>

      {/* Invite friends card */}
      <div className="flex w-full items-center justify-between overflow-hidden rounded-[14px] bg-gradient-to-b from-violet-600 to-violet-700 px-7 py-6">
        <div className="flex h-full flex-col justify-center gap-2">
          <span className="w-fit rounded-full bg-white/20 px-3 py-1 text-[11px] font-semibold text-white">
            限时活动
          </span>
          <h3 className="text-[22px] font-extrabold text-white">
            邀请好友一起学
          </h3>
          <p className="w-60 text-[13px] leading-[1.5] text-white/80">
            每邀请一位好友，双方各获 200 XP 奖励
          </p>
        </div>
        <Link
          href="/hall/invite"
          className="flex shrink-0 items-center gap-1.5 rounded-lg bg-white px-5 py-2 text-[13px] font-semibold text-violet-700 hover:bg-white/90"
        >
          <Gift className="h-3.5 w-3.5" />
          邀请好友
          <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>
    </div>
  );
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`

**Step 3: Commit**

```
feat: update ad cards with links and move buttons to right side
```

---

### Task 6: Update page — wire stats, remove leaderboard

**Files:**
- Modify: `src/app/(web)/hall/(main)/(home)/page.tsx`
- Delete: `src/features/web/hall/components/mini-leaderboard.tsx`

**Step 1: Delete mini-leaderboard.tsx**

```bash
rm src/features/web/hall/components/mini-leaderboard.tsx
```

**Step 2: Rewrite the page**

- Replace `fetchUserProfile` with `fetchDashboardStats` from hall service
- Remove `MiniLeaderboard` import
- Pass stats props to `StatsRow`
- Remove `MiniLeaderboard` from right column
- Update right column comment

```typescript
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar";
import { AdCardsRow } from "@/features/web/hall/components/ad-cards-row";
import { StatsRow } from "@/features/web/hall/components/stats-row";
import { GameProgressCard } from "@/features/web/hall/components/game-progress-card";
import { DailyChallengeCard } from "@/features/web/hall/components/daily-challenge-card";
import { fetchDashboardStats } from "@/features/web/hall/services/hall.service";

export default async function HallDashboardPage() {
  const data = await fetchDashboardStats();
  const displayName =
    data?.profile.nickname ?? data?.profile.username ?? "同学";

  return (
    <div className="flex h-full flex-col gap-6 px-8 py-7">
      <GreetingTopBar
        title={`早上好，${displayName} 👋`}
        subtitle="继续你的学习之旅，今天也要加油！"
      />
      <AdCardsRow />
      <StatsRow
        exp={data?.profile.exp ?? 0}
        currentPlayStreak={data?.profile.currentPlayStreak ?? 0}
        masteredTotal={data?.masterStats.total ?? 0}
        masteredThisWeek={data?.masterStats.thisWeek ?? 0}
        reviewPending={data?.reviewStats.pending ?? 0}
      />

      {/* Main content row */}
      <div className="flex flex-1 gap-5">
        {/* Left column - game progress */}
        <div className="flex flex-1 flex-col gap-5">
          <GameProgressCard />
        </div>

        {/* Right column - daily challenge */}
        <div className="flex w-80 shrink-0 flex-col gap-5">
          <DailyChallengeCard />
        </div>
      </div>
    </div>
  );
}
```

**Step 3: Verify build**

Run: `npx tsc --noEmit`

**Step 4: Run dev server and verify page loads**

Run: `npm run dev`

Visit `http://localhost:3000/hall` — verify stats show real data, no leaderboard, daily challenge links to games, ad cards have links and no placeholder squares.

**Step 5: Commit**

```
feat: wire real data to hall dashboard, remove mini leaderboard
```
