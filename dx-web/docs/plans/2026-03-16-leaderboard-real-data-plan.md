# Leaderboard Real Data Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire real data to the leaderboard page with rankings by EXP and play time, supporting all-time and time-windowed boards (daily/weekly/monthly), showing top 100 + current user's rank.

**Architecture:** Raw SQL via Prisma `$queryRaw` (tagged template literals) for efficient `RANK()` window function queries. Server component fetches default view (经验 + 总榜). Client-side tab switching via server actions.

**Tech Stack:** Next.js 16 App Router, Prisma v7 with `@prisma/adapter-pg`, PostgreSQL `RANK()`, Zod, React hooks

---

### Task 1: Create types and Zod schema

**Files:**
- Create: `src/features/web/leaderboard/types/leaderboard.types.ts`
- Create: `src/features/web/leaderboard/schemas/leaderboard.schema.ts`

**Step 1: Create types file**

```typescript
// src/features/web/leaderboard/types/leaderboard.types.ts

export type LeaderboardType = "exp" | "playTime";
export type LeaderboardPeriod = "all" | "day" | "week" | "month";

export type LeaderboardEntry = {
  id: string;
  username: string;
  nickname: string | null;
  avatarId: string | null;
  value: number;
  rank: number;
};

export type LeaderboardResult = {
  entries: LeaderboardEntry[];
  myRank: LeaderboardEntry | null;
};
```

**Step 2: Create schema file**

```typescript
// src/features/web/leaderboard/schemas/leaderboard.schema.ts

import { z } from "zod";

export const leaderboardParamsSchema = z.object({
  type: z.enum(["exp", "playTime"]),
  period: z.enum(["all", "day", "week", "month"]),
});

export type LeaderboardParams = z.infer<typeof leaderboardParamsSchema>;
```

**Step 3: Verify**

Run: `npx tsc --noEmit`
Expected: No errors from these files.

---

### Task 2: Add formatPlayTime helper and leaderboard format helper

**Files:**
- Modify: `src/lib/format.ts`
- Create: `src/features/web/leaderboard/helpers/format-value.ts`

**Step 1: Add formatPlayTime to format.ts**

Add after the existing `formatDate` function:

```typescript
/** Format seconds into human-readable duration (e.g., "3h 25m", "48m", "< 1m") */
export function formatPlayTime(seconds: number): string {
  if (seconds < 60) return "< 1m";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`;
  if (hours > 0) return `${hours}h`;
  return `${minutes}m`;
}
```

**Step 2: Create leaderboard format helper**

```typescript
// src/features/web/leaderboard/helpers/format-value.ts

import { formatPlayTime } from "@/lib/format";
import type { LeaderboardType } from "../types/leaderboard.types";

/** Format a leaderboard value based on type (EXP as number, play time as duration) */
export function formatLeaderboardValue(value: number, type: LeaderboardType): string {
  if (type === "exp") return value.toLocaleString("zh-CN");
  return formatPlayTime(value);
}
```

**Step 3: Verify**

Run: `npx tsc --noEmit`
Expected: No errors.

---

### Task 3: Create leaderboard service

This is the core task — raw SQL queries for all leaderboard variants.

**Files:**
- Create: `src/features/web/leaderboard/services/leaderboard.service.ts`

**Reference:**
- DB client: `import { db } from "@/lib/db"` (exports `db` and `Prisma`)
- Schema tables: `users`, `game_stats_totals`, `game_session_totals`
- Columns: `users.exp`, `game_stats_totals.total_play_time`, `game_session_totals.exp`, `game_session_totals.play_time`, `game_session_totals.last_played_at`

**Step 1: Create the service file**

```typescript
// src/features/web/leaderboard/services/leaderboard.service.ts

import "server-only";

import { db } from "@/lib/db";
import type {
  LeaderboardEntry,
  LeaderboardResult,
  LeaderboardType,
  LeaderboardPeriod,
} from "../types/leaderboard.types";

/** Raw row shape returned from PostgreSQL queries */
type RawRow = {
  id: string;
  username: string;
  nickname: string | null;
  avatar_id: string | null;
  value: number;
  rank: number;
};

/** Map snake_case raw row to camelCase LeaderboardEntry */
function mapRow(row: RawRow): LeaderboardEntry {
  return {
    id: row.id,
    username: row.username,
    nickname: row.nickname,
    avatarId: row.avatar_id,
    value: Number(row.value),
    rank: Number(row.rank),
  };
}

/** Compute the start-of-period date for a time window (Monday-based weeks) */
function getWindowStart(period: "day" | "week" | "month"): Date {
  const now = new Date();
  switch (period) {
    case "day": {
      const start = new Date(now);
      start.setHours(0, 0, 0, 0);
      return start;
    }
    case "week": {
      const start = new Date(now);
      const day = start.getDay();
      const diff = day === 0 ? 6 : day - 1;
      start.setDate(start.getDate() - diff);
      start.setHours(0, 0, 0, 0);
      return start;
    }
    case "month":
      return new Date(now.getFullYear(), now.getMonth(), 1);
  }
}

/** Split raw rows into top-100 entries and current user's rank */
function splitResults(rows: RawRow[], userId: string | null): LeaderboardResult {
  const entries = rows.filter((r) => Number(r.rank) <= 100).map(mapRow);
  const myRow = userId ? rows.find((r) => r.id === userId) : null;
  return { entries, myRank: myRow ? mapRow(myRow) : null };
}

/** All-time EXP leaderboard — ranks by User.exp */
async function getAllTimeExp(userId: string | null): Promise<LeaderboardResult> {
  const uid = userId ?? "";
  const rows = await db.$queryRaw<RawRow[]>`
    WITH ranked AS (
      SELECT id, username, nickname, avatar_id, exp AS value,
             RANK() OVER (ORDER BY exp DESC)::int AS rank
      FROM users
      WHERE is_active = true AND exp > 0
    )
    SELECT * FROM ranked
    WHERE rank <= 100 OR id = ${uid}
    ORDER BY rank
  `;
  return splitResults(rows, userId);
}

/** All-time play time leaderboard — ranks by SUM(GameStatsTotal.totalPlayTime) */
async function getAllTimePlayTime(userId: string | null): Promise<LeaderboardResult> {
  const uid = userId ?? "";
  const rows = await db.$queryRaw<RawRow[]>`
    WITH ranked AS (
      SELECT u.id, u.username, u.nickname, u.avatar_id,
             COALESCE(SUM(g.total_play_time), 0)::int AS value,
             RANK() OVER (ORDER BY COALESCE(SUM(g.total_play_time), 0) DESC)::int AS rank
      FROM users u
      INNER JOIN game_stats_totals g ON g.user_id = u.id
      WHERE u.is_active = true
      GROUP BY u.id, u.username, u.nickname, u.avatar_id
      HAVING COALESCE(SUM(g.total_play_time), 0) > 0
    )
    SELECT * FROM ranked
    WHERE rank <= 100 OR id = ${uid}
    ORDER BY rank
  `;
  return splitResults(rows, userId);
}

/** Windowed EXP leaderboard — ranks by SUM(GameSessionTotal.exp) within period */
async function getWindowedExp(
  period: "day" | "week" | "month",
  userId: string | null
): Promise<LeaderboardResult> {
  const uid = userId ?? "";
  const windowStart = getWindowStart(period);
  const windowEnd = new Date();
  const rows = await db.$queryRaw<RawRow[]>`
    WITH ranked AS (
      SELECT u.id, u.username, u.nickname, u.avatar_id,
             COALESCE(SUM(s.exp), 0)::int AS value,
             RANK() OVER (ORDER BY COALESCE(SUM(s.exp), 0) DESC)::int AS rank
      FROM users u
      INNER JOIN game_session_totals s ON s.user_id = u.id
        AND s.last_played_at >= ${windowStart}
        AND s.last_played_at < ${windowEnd}
      WHERE u.is_active = true
      GROUP BY u.id, u.username, u.nickname, u.avatar_id
      HAVING COALESCE(SUM(s.exp), 0) > 0
    )
    SELECT * FROM ranked
    WHERE rank <= 100 OR id = ${uid}
    ORDER BY rank
  `;
  return splitResults(rows, userId);
}

/** Windowed play time leaderboard — ranks by SUM(GameSessionTotal.playTime) within period */
async function getWindowedPlayTime(
  period: "day" | "week" | "month",
  userId: string | null
): Promise<LeaderboardResult> {
  const uid = userId ?? "";
  const windowStart = getWindowStart(period);
  const windowEnd = new Date();
  const rows = await db.$queryRaw<RawRow[]>`
    WITH ranked AS (
      SELECT u.id, u.username, u.nickname, u.avatar_id,
             COALESCE(SUM(s.play_time), 0)::int AS value,
             RANK() OVER (ORDER BY COALESCE(SUM(s.play_time), 0) DESC)::int AS rank
      FROM users u
      INNER JOIN game_session_totals s ON s.user_id = u.id
        AND s.last_played_at >= ${windowStart}
        AND s.last_played_at < ${windowEnd}
      WHERE u.is_active = true
      GROUP BY u.id, u.username, u.nickname, u.avatar_id
      HAVING COALESCE(SUM(s.play_time), 0) > 0
    )
    SELECT * FROM ranked
    WHERE rank <= 100 OR id = ${uid}
    ORDER BY rank
  `;
  return splitResults(rows, userId);
}

/** Get leaderboard data for the given type and period */
export async function getLeaderboard(
  type: LeaderboardType,
  period: LeaderboardPeriod,
  userId: string | null
): Promise<LeaderboardResult> {
  if (period === "all") {
    return type === "exp" ? getAllTimeExp(userId) : getAllTimePlayTime(userId);
  }
  return type === "exp"
    ? getWindowedExp(period, userId)
    : getWindowedPlayTime(period, userId);
}
```

**Step 2: Verify**

Run: `npx tsc --noEmit`
Expected: No errors. Key safety checks:
- All queries use `$queryRaw` tagged templates (auto-parameterized)
- `type`/`period` select query branches — never interpolated into SQL
- `userId` passed as parameter, never concatenated
- Date boundaries computed server-side

---

### Task 4: Create server action

**Files:**
- Create: `src/features/web/leaderboard/actions/leaderboard.action.ts`

**Step 1: Create the action file**

```typescript
// src/features/web/leaderboard/actions/leaderboard.action.ts

"use server";

import { auth } from "@/lib/auth";
import { leaderboardParamsSchema } from "../schemas/leaderboard.schema";
import { getLeaderboard } from "../services/leaderboard.service";
import type { LeaderboardResult } from "../types/leaderboard.types";

type FetchLeaderboardResult = LeaderboardResult | { error: string };

/** Fetch leaderboard data for the given type and period */
export async function fetchLeaderboardAction(
  type: string,
  period: string
): Promise<FetchLeaderboardResult> {
  const parsed = leaderboardParamsSchema.safeParse({ type, period });
  if (!parsed.success) return { error: "参数无效" };

  const session = await auth();
  const userId = session?.user?.id ?? null;

  try {
    return await getLeaderboard(parsed.data.type, parsed.data.period, userId);
  } catch {
    return { error: "获取排行榜失败，请重试" };
  }
}
```

**Step 2: Verify**

Run: `npx tsc --noEmit`
Expected: No errors.

---

### Task 5: Create useLeaderboard hook

**Files:**
- Create: `src/features/web/leaderboard/hooks/use-leaderboard.ts`

**Step 1: Create the hook file**

```typescript
// src/features/web/leaderboard/hooks/use-leaderboard.ts

"use client";

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { fetchLeaderboardAction } from "../actions/leaderboard.action";
import type {
  LeaderboardType,
  LeaderboardPeriod,
  LeaderboardResult,
} from "../types/leaderboard.types";

interface UseLeaderboardParams {
  initialData: LeaderboardResult;
}

/** Manage leaderboard tab state and data fetching */
export function useLeaderboard({ initialData }: UseLeaderboardParams) {
  const [type, setType] = useState<LeaderboardType>("exp");
  const [period, setPeriod] = useState<LeaderboardPeriod>("all");
  const [data, setData] = useState<LeaderboardResult>(initialData);
  const [isLoading, setIsLoading] = useState(false);

  /** Fetch leaderboard data for a type+period combination */
  const fetchData = useCallback(
    async (newType: LeaderboardType, newPeriod: LeaderboardPeriod) => {
      setIsLoading(true);
      try {
        const result = await fetchLeaderboardAction(newType, newPeriod);
        if ("error" in result) {
          toast.error(result.error);
          return;
        }
        setData(result);
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  /** Switch the leaderboard type tab */
  const handleTypeChange = useCallback(
    (newType: LeaderboardType) => {
      if (newType === type) return;
      setType(newType);
      fetchData(newType, period);
    },
    [type, period, fetchData]
  );

  /** Switch the leaderboard period tab */
  const handlePeriodChange = useCallback(
    (newPeriod: LeaderboardPeriod) => {
      if (newPeriod === period) return;
      setPeriod(newPeriod);
      fetchData(type, newPeriod);
    },
    [type, period, fetchData]
  );

  return { type, period, data, isLoading, handleTypeChange, handlePeriodChange };
}
```

**Step 2: Verify**

Run: `npx tsc --noEmit`
Expected: No errors.

---

### Task 6: Create LeaderboardMyRank component

**Files:**
- Create: `src/features/web/leaderboard/components/leaderboard-my-rank.tsx`

**Step 1: Create the component**

```tsx
// src/features/web/leaderboard/components/leaderboard-my-rank.tsx

import { Zap, Clock } from "lucide-react";
import { formatLeaderboardValue } from "../helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "../types/leaderboard.types";

interface LeaderboardMyRankProps {
  entry: LeaderboardEntry | null;
  type: LeaderboardType;
}

/** Current user's rank bar — always visible at the top */
export function LeaderboardMyRank({ entry, type }: LeaderboardMyRankProps) {
  const displayName = entry ? (entry.nickname ?? entry.username) : "我";
  const rank = entry?.rank ?? "—";
  const value = entry ? formatLeaderboardValue(entry.value, type) : "0";
  const Icon = type === "exp" ? Zap : Clock;

  return (
    <div className="flex w-full items-center gap-4 rounded-full border-[1.5px] border-teal-600 bg-teal-50 px-4 py-3.5 md:px-6">
      <span className="text-base font-bold text-teal-600">{rank}</span>
      <div className="h-10 w-10 shrink-0 rounded-full bg-slate-200" />
      <div className="flex flex-1 items-center gap-1.5">
        <span className="text-sm font-semibold text-slate-900">
          {displayName} (我)
        </span>
      </div>
      <span className="hidden rounded-xl bg-teal-600 px-3 py-1 text-[11px] font-semibold text-white sm:inline">
        我的排名
      </span>
      <div className="flex items-center gap-1">
        <Icon className="h-3.5 w-3.5 text-amber-500" />
        <span className="text-sm font-semibold text-slate-900">{value}</span>
      </div>
    </div>
  );
}
```

---

### Task 7: Create LeaderboardPodium component

**Files:**
- Create: `src/features/web/leaderboard/components/leaderboard-podium.tsx`

**Step 1: Create the component**

Note: Podium visual order is [2nd, 1st, 3rd] (silver-gold-bronze left to right). Entries are sorted by rank, so we rearrange indices.

```tsx
// src/features/web/leaderboard/components/leaderboard-podium.tsx

import { Crown, Zap, Clock } from "lucide-react";
import { formatLeaderboardValue } from "../helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "../types/leaderboard.types";

const PODIUM_SLOTS = [
  { index: 1, color: "bg-slate-300", height: "h-24" },
  { index: 0, color: "bg-amber-400", height: "h-32" },
  { index: 2, color: "bg-amber-600", height: "h-20" },
];

interface LeaderboardPodiumProps {
  entries: LeaderboardEntry[];
  type: LeaderboardType;
}

/** Top-3 podium display with medal heights */
export function LeaderboardPodium({ entries, type }: LeaderboardPodiumProps) {
  const Icon = type === "exp" ? Zap : Clock;

  return (
    <div className="flex flex-col items-center gap-4 bg-gradient-to-b from-teal-50 to-white px-6 pb-0 pt-6 sm:flex-row sm:items-end sm:justify-center sm:gap-6 sm:px-10">
      {PODIUM_SLOTS.map(({ index, color, height }) => {
        const entry = entries[index];
        if (!entry) return null;
        const displayName = entry.nickname ?? entry.username;

        return (
          <div key={entry.id} className="flex flex-col items-center gap-1.5">
            {entry.rank === 1 && (
              <Crown className="h-5 w-5 text-amber-400" />
            )}
            <div className="h-12 w-12 rounded-full bg-slate-200" />
            <span className="text-[13px] font-semibold text-slate-700">
              {displayName}
            </span>
            <div className="flex items-center gap-1">
              <Icon className="h-3 w-3 text-amber-500" />
              <span className="text-xs font-semibold text-slate-500">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
            <div
              className={`hidden w-40 ${height} rounded-t-lg ${color} items-center justify-center text-2xl font-extrabold text-white sm:flex`}
            >
              {entry.rank}
            </div>
          </div>
        );
      })}
    </div>
  );
}
```

---

### Task 8: Create LeaderboardList component

**Files:**
- Create: `src/features/web/leaderboard/components/leaderboard-list.tsx`

**Step 1: Create the component**

```tsx
// src/features/web/leaderboard/components/leaderboard-list.tsx

import { Zap, Clock } from "lucide-react";
import { formatLeaderboardValue } from "../helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "../types/leaderboard.types";

interface LeaderboardListProps {
  entries: LeaderboardEntry[];
  type: LeaderboardType;
}

/** Rank list for positions #4 and beyond */
export function LeaderboardList({ entries, type }: LeaderboardListProps) {
  const Icon = type === "exp" ? Zap : Clock;

  if (entries.length === 0) return null;

  return (
    <>
      {entries.map((entry) => {
        const displayName = entry.nickname ?? entry.username;

        return (
          <div
            key={entry.id}
            className="flex items-center gap-4 border-b border-slate-100 px-4 py-3 last:border-b-0 md:px-6"
          >
            <span className="w-6 text-sm font-bold text-slate-400">
              {entry.rank}
            </span>
            <div className="h-9 w-9 shrink-0 rounded-full bg-slate-200" />
            <span className="flex-1 text-[13px] font-semibold text-slate-700">
              {displayName}
            </span>
            <div className="flex items-center gap-1">
              <Icon className="h-3.5 w-3.5 text-amber-500" />
              <span className="text-sm font-semibold text-slate-900">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
          </div>
        );
      })}
    </>
  );
}
```

---

### Task 9: Rewrite LeaderboardContent and update page.tsx

**Files:**
- Rewrite: `src/features/web/leaderboard/components/leaderboard-content.tsx`
- Modify: `src/app/(web)/hall/(main)/leaderboard/page.tsx`

**Step 1: Rewrite LeaderboardContent**

Replace entire file contents:

```tsx
// src/features/web/leaderboard/components/leaderboard-content.tsx

"use client";

import { Loader2 } from "lucide-react";
import { TabPill } from "@/components/in/tab-pill";
import { useLeaderboard } from "../hooks/use-leaderboard";
import { LeaderboardMyRank } from "./leaderboard-my-rank";
import { LeaderboardPodium } from "./leaderboard-podium";
import { LeaderboardList } from "./leaderboard-list";
import type {
  LeaderboardType,
  LeaderboardPeriod,
  LeaderboardResult,
} from "../types/leaderboard.types";

const TYPE_TABS: { label: string; value: LeaderboardType }[] = [
  { label: "经验", value: "exp" },
  { label: "时长", value: "playTime" },
];

const PERIOD_TABS: { label: string; value: LeaderboardPeriod }[] = [
  { label: "总榜", value: "all" },
  { label: "日榜", value: "day" },
  { label: "周榜", value: "week" },
  { label: "月榜", value: "month" },
];

interface LeaderboardContentProps {
  initialData: LeaderboardResult;
}

/** Leaderboard content with type/period tab switching */
export function LeaderboardContent({ initialData }: LeaderboardContentProps) {
  const { type, period, data, isLoading, handleTypeChange, handlePeriodChange } =
    useLeaderboard({ initialData });

  const podiumEntries = data.entries.slice(0, 3);
  const listEntries = data.entries.slice(3);

  return (
    <>
      {/* Tab rows */}
      <div className="flex w-full flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2">
          {TYPE_TABS.map((tab) => (
            <TabPill
              key={tab.value}
              label={tab.label}
              active={type === tab.value}
              onClick={() => handleTypeChange(tab.value)}
            />
          ))}
        </div>
        <div className="flex items-center gap-2">
          {PERIOD_TABS.map((tab) => (
            <TabPill
              key={tab.value}
              label={tab.label}
              active={period === tab.value}
              onClick={() => handlePeriodChange(tab.value)}
              size="sm"
            />
          ))}
        </div>
      </div>

      {/* My rank */}
      <LeaderboardMyRank entry={data.myRank} type={type} />

      {/* Leaderboard content */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
        </div>
      ) : data.entries.length === 0 ? (
        <div className="flex items-center justify-center rounded-xl border border-slate-200 bg-white py-20 text-sm text-slate-400">
          暂无排名数据
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          {podiumEntries.length > 0 && (
            <>
              <LeaderboardPodium entries={podiumEntries} type={type} />
              <div className="h-px w-full bg-slate-100" />
            </>
          )}
          <LeaderboardList entries={listEntries} type={type} />
        </div>
      )}
    </>
  );
}
```

**Step 2: Update page.tsx**

Replace entire file contents:

```tsx
// src/app/(web)/hall/(main)/leaderboard/page.tsx

import { auth } from "@/lib/auth";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { LeaderboardContent } from "@/features/web/leaderboard/components/leaderboard-content";
import { getLeaderboard } from "@/features/web/leaderboard/services/leaderboard.service";

export default async function LeaderboardPage() {
  const session = await auth();
  const userId = session?.user?.id ?? null;
  const initialData = await getLeaderboard("exp", "all", userId);

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <PageTopBar
        title="排行榜"
        subtitle="查看学习排名，与好友一起进步"
        searchPlaceholder="搜索用户..."
      />
      <LeaderboardContent initialData={initialData} />
    </div>
  );
}
```

---

### Task 10: Verify and commit

**Step 1: Type check**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 2: Build check**

Run: `npm run build`
Expected: Build succeeds.

**Step 3: Manual verification**

Run: `npm run dev`
Navigate to: `http://localhost:3000/hall/leaderboard`
Verify:
- Default view shows 经验 + 总榜
- Switching type tab (经验/时长) fetches new data
- Switching period tab (总榜/日榜/周榜/月榜) fetches new data
- Current user's rank shows at top
- Top 3 appear in podium, rest in list
- Loading spinner shows during fetch
- Empty state shows "暂无排名数据" when no data
- EXP values show comma-formatted numbers
- Play time values show human-readable durations

**Step 4: Commit**

```bash
git add src/features/web/leaderboard/ src/app/\(web\)/hall/\(main\)/leaderboard/page.tsx src/lib/format.ts
git commit -m "feat: wire real data to leaderboard page with EXP and play time rankings"
```
