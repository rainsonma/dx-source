# Hall Page Three-Column Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the hall dashboard's main content row into a 3-column equal-width grid with a new 今日明星榜 leaderboard card, refactored 今日打卡 card, and preserved 我的学习进度 card.

**Architecture:** Frontend-only changes. Reuse existing `GET /api/leaderboard?type=exp&period=day` endpoint, slice to 50 on client. New hook `useTodayStars` manages type switching and data fetching. Three equal columns via CSS grid on desktop, stacked on mobile with 今日明星榜 first.

**Tech Stack:** Next.js 16, React 19, TailwindCSS v4, Lucide React, existing leaderboard types/helpers

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `dx-web/src/features/web/hall/hooks/use-today-stars.ts` | Create | Hook: fetch leaderboard with `period=day`, type switching, slice to 50 |
| `dx-web/src/features/web/hall/components/today-stars-card.tsx` | Create | 今日明星榜 card: mini podium, scrollable list, my-rank bar, type tabs |
| `dx-web/src/features/web/hall/components/daily-challenge-card.tsx` | Modify | Refactor into 今日打卡 grouped card with two task items |
| `dx-web/src/features/web/hall/components/game-progress-card.tsx` | Modify | Remove `flex-1` width control, let grid handle sizing |
| `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` | Modify | Change main content row to 3-column grid, add TodayStarsCard |

---

### Task 1: Create `useTodayStars` Hook

**Files:**
- Create: `dx-web/src/features/web/hall/hooks/use-today-stars.ts`

- [ ] **Step 1: Create the hook file**

```ts
"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { fetchLeaderboardAction } from "@/features/web/leaderboard/actions/leaderboard.action";
import type {
  LeaderboardType,
  LeaderboardResult,
} from "@/features/web/leaderboard/types/leaderboard.types";

const MAX_ENTRIES = 50;

/** Fetch today's leaderboard (day period only), type-switchable, sliced to 50 */
export function useTodayStars() {
  const [type, setType] = useState<LeaderboardType>("exp");
  const [data, setData] = useState<LeaderboardResult>({ entries: [], myRank: null });
  const [isLoading, setIsLoading] = useState(true);

  const fetchData = useCallback(async (t: LeaderboardType) => {
    setIsLoading(true);
    try {
      const result = await fetchLeaderboardAction(t, "day");
      if ("error" in result) {
        toast.error(result.error);
        return;
      }
      setData({
        entries: result.entries.slice(0, MAX_ENTRIES),
        myRank: result.myRank,
      });
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData("exp");
  }, [fetchData]);

  const handleTypeChange = useCallback(
    (newType: LeaderboardType) => {
      if (newType === type) return;
      setType(newType);
      fetchData(newType);
    },
    [type, fetchData]
  );

  return { type, data, isLoading, handleTypeChange };
}
```

- [ ] **Step 2: Verify no lint errors**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx eslint src/features/web/hall/hooks/use-today-stars.ts`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/hall/hooks/use-today-stars.ts
git commit -m "feat: add useTodayStars hook for hall leaderboard card"
```

---

### Task 2: Create `TodayStarsCard` Component

**Files:**
- Create: `dx-web/src/features/web/hall/components/today-stars-card.tsx`

- [ ] **Step 1: Create the component file**

```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, Clock, Crown, Loader2, Trophy, Zap } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { TabPill } from "@/components/in/tab-pill";
import { getAvatarColor } from "@/lib/avatar";
import { formatLeaderboardValue } from "@/features/web/leaderboard/helpers/format-value";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { useTodayStars } from "@/features/web/hall/hooks/use-today-stars";
import type { LeaderboardEntry, LeaderboardType } from "@/features/web/leaderboard/types/leaderboard.types";

const TYPE_TABS: { label: string; value: LeaderboardType }[] = [
  { label: "经验", value: "exp" },
  { label: "时长", value: "playTime" },
];

const PODIUM_SLOTS = [
  { index: 1, color: "bg-slate-300", height: "h-16" },
  { index: 0, color: "bg-amber-400", height: "h-20" },
  { index: 2, color: "bg-amber-600", height: "h-12" },
];

/** Mini podium for top 3 users */
function MiniPodium({ entries, type }: { entries: LeaderboardEntry[]; type: LeaderboardType }) {
  const Icon = type === "exp" ? Zap : Clock;

  return (
    <div className="flex items-end justify-center gap-3 rounded-lg bg-gradient-to-b from-teal-50 to-white px-4 pb-0 pt-4">
      {PODIUM_SLOTS.map(({ index, color, height }) => {
        const entry = entries[index];
        if (!entry) return null;
        const displayName = entry.nickname ?? entry.username;
        const fallbackChar = displayName.charAt(0).toUpperCase();
        const avatarBg = getAvatarColor(entry.id);

        return (
          <div key={entry.id} className="flex flex-col items-center gap-1">
            {entry.rank === 1 && (
              <Crown className="h-4 w-4 text-amber-400" />
            )}
            <Avatar size="sm">
              {entry.avatarUrl && (
                <AvatarImage src={entry.avatarUrl} alt={displayName} />
              )}
              <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>
                {fallbackChar}
              </AvatarFallback>
            </Avatar>
            <span className="max-w-[72px] truncate text-xs font-semibold text-foreground">
              {displayName}
            </span>
            <div className="flex items-center gap-0.5">
              <Icon className="h-2.5 w-2.5 text-amber-500" />
              <span className="text-[10px] font-semibold text-muted-foreground">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
            <div
              className={`w-16 ${height} rounded-t-md ${color} flex items-center justify-center text-base font-extrabold text-white`}
            >
              {entry.rank}
            </div>
          </div>
        );
      })}
    </div>
  );
}

/** Scrollable list for rank 4–50 */
function StarsList({ entries, type }: { entries: LeaderboardEntry[]; type: LeaderboardType }) {
  const Icon = type === "exp" ? Zap : Clock;

  if (entries.length === 0) return null;

  return (
    <div className="max-h-[280px] overflow-y-auto">
      {entries.map((entry) => {
        const displayName = entry.nickname ?? entry.username;
        const fallbackChar = displayName.charAt(0).toUpperCase();
        const avatarBg = getAvatarColor(entry.id);

        return (
          <div
            key={entry.id}
            className="flex items-center gap-3 border-b border-border px-4 py-2.5 last:border-b-0"
          >
            <span className="w-5 text-xs font-bold text-muted-foreground">
              {entry.rank}
            </span>
            <Avatar size="sm">
              {entry.avatarUrl && (
                <AvatarImage src={entry.avatarUrl} alt={displayName} />
              )}
              <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>
                {fallbackChar}
              </AvatarFallback>
            </Avatar>
            <span className="flex-1 truncate text-xs font-semibold text-foreground">
              {displayName}
            </span>
            <div className="flex items-center gap-0.5">
              <Icon className="h-3 w-3 text-amber-500" />
              <span className="text-xs font-semibold text-foreground">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

/** My rank bar */
function MyRankBar({
  entry,
  type,
  user,
}: {
  entry: LeaderboardEntry | null;
  type: LeaderboardType;
  user: { id: string; username: string; nickname: string | null; avatarUrl: string | null };
}) {
  const displayName = user.nickname ?? user.username;
  const rank = entry?.rank ?? null;
  const value = entry ? formatLeaderboardValue(entry.value, type) : "0";
  const Icon = type === "exp" ? Zap : Clock;
  const fallbackChar = displayName.charAt(0).toUpperCase();
  const avatarBg = getAvatarColor(user.id);

  return (
    <div className="flex items-center gap-3 rounded-full border-[1.5px] border-teal-600 bg-teal-50 px-3 py-2">
      {rank !== null && (
        <span className="text-sm font-bold text-teal-600">{rank}</span>
      )}
      <Avatar size="sm">
        {user.avatarUrl && (
          <AvatarImage src={user.avatarUrl} alt={displayName} />
        )}
        <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>
          {fallbackChar}
        </AvatarFallback>
      </Avatar>
      <span className="flex-1 truncate text-xs font-semibold text-foreground">
        {displayName}
      </span>
      <div className="flex items-center gap-0.5">
        <Icon className="h-3 w-3 text-amber-500" />
        <span className="text-xs font-semibold text-foreground">{value}</span>
      </div>
    </div>
  );
}

/** 今日明星榜 — Today's star leaderboard for the hall dashboard */
export function TodayStarsCard() {
  const { type, data, isLoading, handleTypeChange } = useTodayStars();

  const [user, setUser] = useState<{
    id: string; username: string; nickname: string | null; avatarUrl: string | null;
  } | null>(null);

  useEffect(() => {
    fetchUserProfile().then((profile) => {
      if (profile) setUser(profile);
    });
  }, []);

  const podiumEntries = data.entries.slice(0, 3);
  const listEntries = data.entries.slice(3);

  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-amber-50">
            <Trophy className="h-5 w-5 text-amber-500" />
          </div>
          <h3 className="text-base font-bold text-foreground">今日明星榜</h3>
        </div>
        <Link
          href="/hall/leaderboard"
          className="flex items-center gap-1 text-[13px] font-semibold text-teal-600 hover:text-teal-700"
        >
          查看全部
          <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>

      {/* Type tabs */}
      <div className="flex items-center gap-2">
        {TYPE_TABS.map((tab) => (
          <TabPill
            key={tab.value}
            label={tab.label}
            active={type === tab.value}
            onClick={() => handleTypeChange(tab.value)}
            size="sm"
          />
        ))}
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-5 w-5 animate-spin text-teal-600" />
        </div>
      ) : data.entries.length === 0 ? (
        <div className="flex items-center justify-center rounded-lg border border-border py-12 text-sm text-muted-foreground">
          暂无排名数据
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-border">
          {podiumEntries.length > 0 && (
            <>
              <MiniPodium entries={podiumEntries} type={type} />
              <div className="h-px w-full bg-border" />
            </>
          )}
          <StarsList entries={listEntries} type={type} />
        </div>
      )}

      {/* My rank */}
      {user && <MyRankBar entry={data.myRank} type={type} user={user} />}
    </div>
  );
}
```

- [ ] **Step 2: Verify no lint errors**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx eslint src/features/web/hall/components/today-stars-card.tsx`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/hall/components/today-stars-card.tsx
git commit -m "feat: add TodayStarsCard component for hall dashboard"
```

---

### Task 3: Refactor `DailyChallengeCard` into 今日打卡

**Files:**
- Modify: `dx-web/src/features/web/hall/components/daily-challenge-card.tsx`

- [ ] **Step 1: Rewrite the component**

Replace the entire file with:

```tsx
import Link from "next/link";
import { Flame, MessageCircle, Play } from "lucide-react";

/** 今日打卡 — grouped card with two daily task items */
export function DailyChallengeCard() {
  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-5">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-orange-50">
          <Flame className="h-5 w-5 text-orange-500" />
        </div>
        <h3 className="text-base font-bold text-foreground">今日打卡</h3>
      </div>

      {/* Task 1 — Game challenge */}
      <div className="flex flex-col gap-3 rounded-xl bg-teal-600 p-5">
        <p className="text-sm font-bold leading-relaxed text-white">
          完成今天的连词成句
          <br />
          赢取双倍经验值！
        </p>
        <Link
          href="/hall/games"
          className="flex h-10 w-full items-center justify-center gap-2 rounded-[10px] bg-white text-[13px] font-semibold text-teal-700 hover:bg-white/90"
        >
          <Play className="h-4 w-4" />
          开始挑战
        </Link>
      </div>

      {/* Task 2 — Community post */}
      <div className="flex flex-col gap-3 rounded-xl bg-gradient-to-br from-teal-500 to-teal-700 p-5">
        <p className="text-sm font-bold leading-relaxed text-white">
          前往「斗学社」发表一条英文动态贴，进步需要坚持不懈!
        </p>
        <Link
          href="/hall/community"
          className="flex h-10 w-full items-center justify-center gap-2 rounded-[10px] bg-white text-[13px] font-semibold text-teal-700 hover:bg-white/90"
        >
          <MessageCircle className="h-4 w-4" />
          去发帖
        </Link>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify no lint errors**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx eslint src/features/web/hall/components/daily-challenge-card.tsx`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/hall/components/daily-challenge-card.tsx
git commit -m "refactor: convert DailyChallengeCard into grouped 今日打卡 card with two tasks"
```

---

### Task 4: Update `GameProgressCard` Width

**Files:**
- Modify: `dx-web/src/features/web/hall/components/game-progress-card.tsx:42`

- [ ] **Step 1: Remove flex-1 from the outer wrapper**

The outer `<div>` at line 42 currently has `flex w-full`. No change needed here — the `flex-1` was on the parent wrapper in `page.tsx`, not on this component. The component already uses `w-full` which will fill the grid cell correctly.

Verify by reading line 42:
```tsx
<div className="flex w-full flex-col gap-5 rounded-[14px] border border-border bg-card p-6">
```

No code change needed for this component — it already works correctly in a grid context.

- [ ] **Step 2: Verify no lint errors**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx eslint src/features/web/hall/components/game-progress-card.tsx`
Expected: No errors

---

### Task 5: Update Hall Dashboard Page Layout

**Files:**
- Modify: `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx`

- [ ] **Step 1: Add TodayStarsCard import**

Add after the DailyChallengeCard import (line 9):

```tsx
import { TodayStarsCard } from "@/features/web/hall/components/today-stars-card";
```

- [ ] **Step 2: Replace the main content row**

Replace lines 75–86 (the `{/* Main content row */}` section):

```tsx
      {/* Main content row */}
      <div className="flex flex-col gap-5 lg:grid lg:grid-cols-3">
        {/* Right column - today's stars (shows first on mobile) */}
        <div className="order-first lg:order-last">
          <TodayStarsCard />
        </div>

        {/* Left column - game progress */}
        <div className="order-2 lg:order-first">
          <GameProgressCard sessions={data?.sessions ?? []} />
        </div>

        {/* Center column - daily check-in */}
        <div className="order-3 lg:order-2">
          <DailyChallengeCard />
        </div>
      </div>
```

- [ ] **Step 3: Verify no lint errors**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx eslint src/app/\(web\)/hall/\(main\)/\(home\)/page.tsx`
Expected: No errors

- [ ] **Step 4: Verify the build compiles**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run build`
Expected: Build succeeds with no errors

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/app/\(web\)/hall/\(main\)/\(home\)/page.tsx
git commit -m "feat: redesign hall dashboard to 3-column grid with today's stars leaderboard"
```

---

### Task 6: Final Verification

- [ ] **Step 1: Run full lint check on all changed files**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx eslint src/features/web/hall/hooks/use-today-stars.ts src/features/web/hall/components/today-stars-card.tsx src/features/web/hall/components/daily-challenge-card.tsx src/features/web/hall/components/game-progress-card.tsx src/app/\(web\)/hall/\(main\)/\(home\)/page.tsx`
Expected: No errors or warnings

- [ ] **Step 2: Run full build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run build`
Expected: Build succeeds

- [ ] **Step 3: Visual smoke test**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run dev`

Verify in browser at `http://localhost:3000/hall`:
1. Desktop (lg+): Three equal-width columns appear
2. Left column: 我的学习进度 with pagination works
3. Center column: 今日打卡 card with two task blocks
4. Right column: 今日明星榜 with podium, list, my-rank, type switching
5. Mobile: Columns stack — stars first, then progress, then check-in
6. Existing pages (`/hall/leaderboard`, game pages, etc.) still work
