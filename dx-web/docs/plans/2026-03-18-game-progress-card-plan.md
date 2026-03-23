# Game Progress Card Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire real `GameSessionTotal` data into the hall dashboard's "我的游戏进度" card with pagination.

**Architecture:** SSR page fetches all user sessions via service → passes to client component → client-side pagination with `DataTablePagination`.

**Tech Stack:** Next.js 16 App Router, Prisma, React client component, DataTablePagination

---

### Task 1: Add `getUserSessionTotals` query

**Files:**
- Modify: `src/models/game-session-total/game-session-total.query.ts:122` (append)

**Step 1: Add the query and exported type**

Append to the end of `game-session-total.query.ts`:

```typescript
export type SessionProgressItem = {
  id: string;
  gameId: string;
  gameName: string;
  gameMode: string;
  degree: string;
  pattern: string | null;
  playedLevelsCount: number;
  totalLevelsCount: number;
  score: number;
  exp: number;
  lastPlayedAt: Date;
  endedAt: Date | null;
};

/** Get all session totals for a user with game info, newest first */
export async function getUserSessionTotals(
  userId: string
): Promise<SessionProgressItem[]> {
  const sessions = await db.gameSessionTotal.findMany({
    where: { userId },
    select: {
      id: true,
      gameId: true,
      degree: true,
      pattern: true,
      playedLevelsCount: true,
      totalLevelsCount: true,
      score: true,
      exp: true,
      lastPlayedAt: true,
      endedAt: true,
      game: {
        select: {
          name: true,
          mode: true,
        },
      },
    },
    orderBy: { lastPlayedAt: "desc" },
  });

  return sessions.map((s) => ({
    id: s.id,
    gameId: s.gameId,
    gameName: s.game.name,
    gameMode: s.game.mode,
    degree: s.degree,
    pattern: s.pattern,
    playedLevelsCount: s.playedLevelsCount,
    totalLevelsCount: s.totalLevelsCount,
    score: s.score,
    exp: s.exp,
    lastPlayedAt: s.lastPlayedAt,
    endedAt: s.endedAt,
  }));
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No new errors

---

### Task 2: Wire sessions into `fetchDashboardStats`

**Files:**
- Modify: `src/features/web/hall/services/hall.service.ts:1-18`

**Step 1: Update the service**

Replace the entire file with:

```typescript
import "server-only";

import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { getUserMasterStats } from "@/models/user-master/user-master.query";
import { getUserReviewStats } from "@/models/user-review/user-review.query";
import { getUserSessionTotals } from "@/models/game-session-total/game-session-total.query";

/** Fetch all stats needed for the hall dashboard home page */
export async function fetchDashboardStats() {
  const profile = await fetchUserProfile();
  if (!profile) return null;

  const [masterStats, reviewStats, sessions] = await Promise.all([
    getUserMasterStats(profile.id),
    getUserReviewStats(profile.id),
    getUserSessionTotals(profile.id),
  ]);

  return { profile, masterStats, reviewStats, sessions };
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No new errors

---

### Task 3: Rewrite `GameProgressCard` with real data + pagination

**Files:**
- Rewrite: `src/features/web/hall/components/game-progress-card.tsx`

**Step 1: Create the helper for progress bar color cycling**

Create `src/features/web/hall/helpers/progress-color.helper.ts`:

```typescript
const PROGRESS_COLORS = [
  "bg-teal-500",
  "bg-blue-500",
  "bg-amber-500",
  "bg-pink-500",
  "bg-violet-500",
  "bg-cyan-500",
];

/** Cycle through progress bar colors by index */
export function getProgressColor(index: number): string {
  return PROGRESS_COLORS[index % PROGRESS_COLORS.length];
}
```

**Step 2: Rewrite the component**

Replace `game-progress-card.tsx` entirely with:

```tsx
"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowRight, Gamepad2 } from "lucide-react";
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { getProgressColor } from "@/features/web/hall/helpers/progress-color.helper";
import type { SessionProgressItem } from "@/models/game-session-total/game-session-total.query";

const PAGE_SIZE = 5;

/** Compute progress percentage from played/total levels */
function calcProgress(played: number, total: number): number {
  if (total === 0) return 0;
  return Math.round((played / total) * 100);
}

type GameProgressCardProps = {
  sessions: SessionProgressItem[];
};

/** Dashboard card listing user's game session progress with pagination */
export function GameProgressCard({ sessions }: GameProgressCardProps) {
  const [currentPage, setCurrentPage] = useState(1);
  const totalPages = Math.max(1, Math.ceil(sessions.length / PAGE_SIZE));

  const start = (currentPage - 1) * PAGE_SIZE;
  const pageItems = sessions.slice(start, start + PAGE_SIZE);

  if (sessions.length === 0) {
    return (
      <div className="flex w-full flex-col items-center justify-center gap-3 rounded-[14px] border border-slate-200 bg-white p-6 py-12 text-slate-400">
        <Gamepad2 className="h-10 w-10 stroke-1" />
        <p className="text-sm">还没有游戏进度，去发现课程游戏吧</p>
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-5 rounded-[14px] border border-slate-200 bg-white p-6">
      {/* Header */}
      <div className="flex w-full items-center justify-between">
        <h3 className="text-base font-bold text-slate-900">我的游戏进度</h3>
        <Link
          href="/hall/games/mine"
          className="flex items-center gap-1 text-[13px] font-semibold text-teal-600 hover:text-teal-700"
        >
          查看全部
          <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>

      {/* Progress list */}
      <div className="flex flex-col gap-3">
        {pageItems.map((session, i) => {
          const modeLabel =
            GAME_MODE_LABELS[session.gameMode as GameMode] ?? session.gameMode;
          const progress = calcProgress(
            session.playedLevelsCount,
            session.totalLevelsCount
          );
          const color = getProgressColor(start + i);

          return (
            <Link
              key={session.id}
              href={`/hall/games/${session.gameId}`}
              className="flex flex-col gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-slate-50"
            >
              <div className="flex items-center justify-between">
                <span className="text-[13px] font-medium text-slate-700">
                  {session.gameName} · {modeLabel}
                </span>
                <span className="text-[13px] font-semibold text-slate-500">
                  {progress}%
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100">
                <div
                  className={`h-full rounded-full ${color}`}
                  style={{ width: `${progress}%` }}
                />
              </div>
            </Link>
          );
        })}
      </div>

      {/* Pagination */}
      <DataTablePagination
        currentPage={currentPage}
        totalPages={totalPages}
        onPageChange={setCurrentPage}
      />
    </div>
  );
}
```

**Step 3: Verify build**

Run: `npx tsc --noEmit`
Expected: No new errors

---

### Task 4: Pass sessions to `GameProgressCard` in page

**Files:**
- Modify: `src/app/(web)/hall/(main)/(home)/page.tsx:32`

**Step 1: Update the page**

Change line 32 from:

```tsx
          <GameProgressCard />
```

to:

```tsx
          <GameProgressCard sessions={data?.sessions ?? []} />
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No new errors

**Step 3: Verify dev server**

Run: `npm run dev`
Visit: `http://localhost:3000/hall`
Expected: "我的游戏进度" shows real session data with progress bars and pagination. "查看全部" navigates to `/hall/games/mine`.

---

### Task 5: Commit

**Step 1: Stage and commit**

```bash
git add src/models/game-session-total/game-session-total.query.ts \
  src/features/web/hall/services/hall.service.ts \
  src/features/web/hall/helpers/progress-color.helper.ts \
  src/features/web/hall/components/game-progress-card.tsx \
  src/app/\(web\)/hall/\(main\)/\(home\)/page.tsx \
  docs/plans/2026-03-18-game-progress-card-design.md \
  docs/plans/2026-03-18-game-progress-card-plan.md
git commit -m "feat: wire real session data to game progress card with pagination"
```
