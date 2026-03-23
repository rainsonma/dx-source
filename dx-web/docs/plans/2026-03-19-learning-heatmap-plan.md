# Learning Heatmap Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a 学习热力图 block to the hall home page that shows daily learning activity as a GitHub-style contribution grid, with year selector and summary stats, using answer counts from `game_records`.

**Architecture:** Raw SQL query aggregates daily answer counts per user per year from `game_records`. Server fetches current year data at SSR; year switching is handled by a server action called from a client component. Pure CSS grid renders the heatmap — no external libraries.

**Tech Stack:** Next.js 16 App Router, Prisma `$queryRaw`, React Server Components + Client Components, TailwindCSS teal theme.

**Design doc:** `docs/plans/2026-03-19-learning-heatmap-design.md`

---

### Task 1: Types

**Files:**
- Create: `src/features/web/hall/types/heatmap.ts`

**Step 1: Create the types file**

```ts
/** Single day in the heatmap */
export type HeatmapDay = {
  date: string; // "YYYY-MM-DD"
  count: number;
};

/** Full heatmap dataset for a year */
export type HeatmapData = {
  year: number;
  days: HeatmapDay[];
};

/** Intensity tier boundaries (answer counts) */
export const HEATMAP_TIERS = [
  { min: 1, max: 10, label: "1~10 题" },
  { min: 11, max: 30, label: "11~30 题" },
  { min: 31, max: 60, label: "31~60 题" },
  { min: 61, max: Infinity, label: "60+ 题" },
] as const;

/** Teal color palette for heatmap tiers (0=empty, 1-4=tiers) */
export const HEATMAP_COLORS = [
  "bg-slate-100",   // 0: no activity
  "bg-teal-200",    // 1: 1-10
  "bg-teal-400",    // 2: 11-30
  "bg-teal-600",    // 3: 31-60
  "bg-teal-800",    // 4: 60+
] as const;
```

**Step 2: Commit**

```
feat: add heatmap types and tier constants
```

---

### Task 2: Helper functions

**Files:**
- Create: `src/features/web/hall/helpers/heatmap.ts`

**Step 1: Create the helper file**

Reference: The heatmap grid is a 53-column (weeks) × 7-row (Mon–Sun) grid for a given year. Each cell represents one day. The grid starts from the first Monday on or before Jan 1 and ends at the last Sunday on or after Dec 31.

```ts
import { HEATMAP_TIERS, type HeatmapDay } from "@/features/web/hall/types/heatmap";

/** Get the intensity tier (0-4) for an answer count */
export function getTier(count: number): number {
  if (count === 0) return 0;
  const idx = HEATMAP_TIERS.findIndex((t) => count >= t.min && count <= t.max);
  return idx === -1 ? 0 : idx + 1;
}

/** Build a Map of "YYYY-MM-DD" → count from HeatmapDay[] */
export function buildDayMap(days: HeatmapDay[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const d of days) {
    map.set(d.date, d.count);
  }
  return map;
}

/** Format a Date as "YYYY-MM-DD" */
export function formatDate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

/** Get the Monday on or before a given date */
function toMonday(d: Date): Date {
  const copy = new Date(d);
  const day = copy.getDay(); // 0=Sun, 1=Mon, ...
  const diff = day === 0 ? 6 : day - 1;
  copy.setDate(copy.getDate() - diff);
  return copy;
}

/** Generate all grid cells for a year (weeks × 7 days) */
export function generateGridCells(year: number): { date: string; row: number; col: number }[] {
  const jan1 = new Date(year, 0, 1);
  const dec31 = new Date(year, 11, 31);
  const gridStart = toMonday(jan1);

  const cells: { date: string; row: number; col: number }[] = [];
  const cursor = new Date(gridStart);
  let col = 0;

  while (cursor <= dec31 || cursor.getDay() !== 1) {
    for (let row = 0; row < 7; row++) {
      cells.push({ date: formatDate(cursor), row, col });
      cursor.setDate(cursor.getDate() + 1);
    }
    col++;
    // Safety: stop after 54 weeks max
    if (col > 53) break;
  }

  return cells;
}

/** Get month label positions for the grid header */
export function getMonthLabels(year: number): { label: string; col: number }[] {
  const jan1 = new Date(year, 0, 1);
  const gridStart = toMonday(jan1);
  const labels: { label: string; col: number }[] = [];

  for (let month = 0; month < 12; month++) {
    const firstOfMonth = new Date(year, month, 1);
    const daysDiff = Math.floor(
      (firstOfMonth.getTime() - gridStart.getTime()) / (1000 * 60 * 60 * 24)
    );
    const col = Math.floor(daysDiff / 7);
    labels.push({ label: `${month + 1}月`, col });
  }

  return labels;
}

/** Compute summary stats from heatmap days */
export function computeHeatmapStats(days: HeatmapDay[]) {
  const activeDays = days.length;
  const totalAnswers = days.reduce((sum, d) => sum + d.count, 0);
  const avgPerDay = activeDays > 0 ? Math.round(totalAnswers / activeDays) : 0;

  const tierCounts = [0, 0, 0, 0]; // tier 1-4
  for (const d of days) {
    const tier = getTier(d.count);
    if (tier > 0) tierCounts[tier - 1]++;
  }

  return { activeDays, totalAnswers, avgPerDay, tierCounts };
}
```

**Step 2: Commit**

```
feat: add heatmap helper functions
```

---

### Task 3: Data query

**Files:**
- Create: `src/models/game-record/game-record.query.ts`

**Step 1: Create the query file**

Note: Currently only `game-record.mutation.ts` exists. We need to create the query file. Follow the pattern from `src/lib/db.ts` — import `db` from `@/lib/db`.

```ts
import "server-only";

import { db } from "@/lib/db";

/** Get daily answer counts for a user in a given year */
export async function getUserDailyAnswerCounts(
  userId: string,
  year: number
): Promise<{ date: Date; count: bigint }[]> {
  const startDate = new Date(`${year}-01-01T00:00:00Z`);
  const endDate = new Date(`${year + 1}-01-01T00:00:00Z`);

  return db.$queryRaw<{ date: Date; count: bigint }[]>`
    SELECT DATE(created_at AT TIME ZONE 'UTC') as date, COUNT(*) as count
    FROM game_records
    WHERE user_id = ${userId}
      AND created_at >= ${startDate}
      AND created_at < ${endDate}
    GROUP BY DATE(created_at AT TIME ZONE 'UTC')
  `;
}
```

**Step 2: Commit**

```
feat: add getUserDailyAnswerCounts query for heatmap
```

---

### Task 4: Service and server action

**Files:**
- Modify: `src/features/web/hall/services/hall.service.ts`
- Create: `src/features/web/hall/actions/heatmap.action.ts`

**Step 1: Add `fetchHeatmapData` to hall.service.ts**

Add to the existing imports at the top:

```ts
import { getUserDailyAnswerCounts } from "@/models/game-record/game-record.query";
```

Add this function after `fetchDashboardStats`:

```ts
/** Fetch heatmap data for a user in a given year */
export async function fetchHeatmapData(year: number) {
  const profile = await fetchUserProfile();
  if (!profile) return null;

  // Clamp year to valid range
  const accountYear = new Date(profile.createdAt).getFullYear();
  const currentYear = new Date().getFullYear();
  const clampedYear = Math.max(accountYear, Math.min(currentYear, year));

  const rows = await getUserDailyAnswerCounts(profile.id, clampedYear);

  const days = rows.map((r) => ({
    date: r.date instanceof Date
      ? r.date.toISOString().split("T")[0]
      : String(r.date),
    count: Number(r.count),
  }));

  return {
    year: clampedYear,
    days,
    accountYear,
  };
}
```

Also update `fetchDashboardStats` to include heatmap data:

```ts
export async function fetchDashboardStats() {
  const profile = await fetchUserProfile();
  if (!profile) return null;

  const currentYear = new Date().getFullYear();
  const accountYear = new Date(profile.createdAt).getFullYear();

  const [masterStats, reviewStats, sessions, heatmapRows] = await Promise.all([
    getUserMasterStats(profile.id),
    getUserReviewStats(profile.id),
    getUserSessionTotals(profile.id),
    getUserDailyAnswerCounts(profile.id, currentYear),
  ]);

  const heatmapDays = heatmapRows.map((r) => ({
    date: r.date instanceof Date
      ? r.date.toISOString().split("T")[0]
      : String(r.date),
    count: Number(r.count),
  }));

  return {
    profile,
    masterStats,
    reviewStats,
    sessions,
    heatmap: { year: currentYear, days: heatmapDays, accountYear },
  };
}
```

Note: `profile.createdAt` must be selected by `getUserProfile`. Check that it is — if not, add `createdAt: true` to the select in `src/models/user/user.query.ts` → `getUserProfile`.

**Step 2: Create server action for year switching**

Create `src/features/web/hall/actions/heatmap.action.ts`:

```ts
"use server";

import { fetchHeatmapData } from "@/features/web/hall/services/hall.service";
import type { HeatmapData } from "@/features/web/hall/types/heatmap";

export type HeatmapActionResult = {
  data: (HeatmapData & { accountYear: number }) | null;
  error?: string;
};

/** Fetch heatmap data for a specific year (client-side year switching) */
export async function fetchHeatmapDataAction(
  year: number
): Promise<HeatmapActionResult> {
  try {
    const result = await fetchHeatmapData(year);
    if (!result) return { data: null, error: "未登录" };
    return { data: result };
  } catch {
    return { data: null, error: "加载失败，请重试" };
  }
}
```

**Step 3: Commit**

```
feat: add heatmap data service and server action
```

---

### Task 5: Heatmap summary component

**Files:**
- Create: `src/features/web/hall/components/heatmap-summary.tsx`

**Step 1: Create the component**

```tsx
import { HEATMAP_TIERS, HEATMAP_COLORS } from "@/features/web/hall/types/heatmap";

type HeatmapSummaryProps = {
  activeDays: number;
  avgPerDay: number;
  tierCounts: number[];
};

/** Right-side panel showing yearly activity and intensity breakdown */
export function HeatmapSummary({
  activeDays,
  avgPerDay,
  tierCounts,
}: HeatmapSummaryProps) {
  return (
    <div className="flex flex-col gap-4">
      {/* Yearly activity */}
      <div className="flex flex-col gap-2 rounded-xl border border-slate-200 bg-white p-4">
        <h4 className="text-sm font-bold text-slate-900">年度活跃</h4>
        <p className="text-[13px] text-slate-600">
          活跃天数 {activeDays} 天
        </p>
        <p className="text-[13px] text-slate-600">
          日均 {avgPerDay} 题/天
        </p>
      </div>

      {/* Intensity breakdown */}
      <div className="flex flex-col gap-2 rounded-xl border border-slate-200 bg-white p-4">
        <h4 className="text-sm font-bold text-slate-900">学习强度</h4>
        <div className="flex flex-col gap-1.5">
          {HEATMAP_TIERS.map((tier, i) => (
            <div key={tier.label} className="flex items-center gap-2 text-[13px] text-slate-600">
              <span className={`inline-block h-3 w-3 rounded-sm ${HEATMAP_COLORS[i + 1]}`} />
              {tier.label} · {tierCounts[i]} 天
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
```

**Step 2: Commit**

```
feat: add heatmap summary component
```

---

### Task 6: Heatmap grid component

**Files:**
- Create: `src/features/web/hall/components/heatmap-grid.tsx`

**Step 1: Create the component**

The grid is a CSS grid with 7 rows (Mon–Sun) and dynamic columns (weeks). Each cell is a small rounded square colored by tier.

```tsx
import { HEATMAP_COLORS } from "@/features/web/hall/types/heatmap";
import type { HeatmapDay } from "@/features/web/hall/types/heatmap";
import {
  generateGridCells,
  getMonthLabels,
  buildDayMap,
  getTier,
} from "@/features/web/hall/helpers/heatmap";

type HeatmapGridProps = {
  year: number;
  days: HeatmapDay[];
};

const DAY_LABELS = ["一", "", "三", "", "五", "", ""];

/** GitHub-style heatmap grid with month and day labels */
export function HeatmapGrid({ year, days }: HeatmapGridProps) {
  const cells = generateGridCells(year);
  const dayMap = buildDayMap(days);
  const monthLabels = getMonthLabels(year);
  const totalCols = cells.length > 0 ? cells[cells.length - 1].col + 1 : 53;

  return (
    <div className="flex flex-col gap-2">
      {/* Month labels */}
      <div
        className="grid gap-[3px] text-xs text-slate-400"
        style={{
          gridTemplateColumns: `24px repeat(${totalCols}, 1fr)`,
        }}
      >
        <span /> {/* spacer for day labels column */}
        {Array.from({ length: totalCols }, (_, col) => {
          const label = monthLabels.find((m) => m.col === col);
          return (
            <span key={col} className="text-center text-[11px]">
              {label?.label ?? ""}
            </span>
          );
        })}
      </div>

      {/* Grid with day labels */}
      <div className="flex gap-[3px]">
        {/* Day labels */}
        <div className="flex w-6 flex-col gap-[3px]">
          {DAY_LABELS.map((label, i) => (
            <div
              key={i}
              className="flex h-[14px] items-center text-[11px] text-slate-400"
            >
              {label}
            </div>
          ))}
        </div>

        {/* Grid cells */}
        <div
          className="grid flex-1 gap-[3px]"
          style={{
            gridTemplateRows: "repeat(7, 14px)",
            gridTemplateColumns: `repeat(${totalCols}, 1fr)`,
            gridAutoFlow: "column",
          }}
        >
          {cells.map((cell) => {
            const count = dayMap.get(cell.date) ?? 0;
            const tier = getTier(count);
            const isInYear =
              cell.date.startsWith(String(year));

            return (
              <div
                key={cell.date}
                className={`h-[14px] rounded-[3px] ${
                  isInYear ? HEATMAP_COLORS[tier] : "bg-transparent"
                }`}
                title={
                  isInYear
                    ? `${cell.date}: ${count > 0 ? `${count} 题` : "无记录"}`
                    : undefined
                }
              />
            );
          })}
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-1.5 text-[11px] text-slate-400">
        <span>少</span>
        {HEATMAP_COLORS.map((color, i) => (
          <div
            key={i}
            className={`h-[12px] w-[12px] rounded-[2px] ${color}`}
          />
        ))}
        <span>多</span>
      </div>
    </div>
  );
}
```

**Step 2: Commit**

```
feat: add heatmap grid component
```

---

### Task 7: Learning heatmap container component

**Files:**
- Create: `src/features/web/hall/components/learning-heatmap.tsx`

**Step 1: Create the container component**

This is a client component that manages year state and fetches data on year change.

```tsx
"use client";

import { useState, useTransition } from "react";
import { CalendarDays } from "lucide-react";
import { HeatmapGrid } from "@/features/web/hall/components/heatmap-grid";
import { HeatmapSummary } from "@/features/web/hall/components/heatmap-summary";
import { computeHeatmapStats } from "@/features/web/hall/helpers/heatmap";
import { fetchHeatmapDataAction } from "@/features/web/hall/actions/heatmap.action";
import type { HeatmapDay } from "@/features/web/hall/types/heatmap";

type LearningHeatmapProps = {
  initialYear: number;
  initialDays: HeatmapDay[];
  accountYear: number;
};

/** Learning heatmap block with year selector */
export function LearningHeatmap({
  initialYear,
  initialDays,
  accountYear,
}: LearningHeatmapProps) {
  const currentYear = new Date().getFullYear();
  const years = Array.from(
    { length: currentYear - accountYear + 1 },
    (_, i) => currentYear - i
  );

  const [selectedYear, setSelectedYear] = useState(initialYear);
  const [days, setDays] = useState(initialDays);
  const [isPending, startTransition] = useTransition();

  const stats = computeHeatmapStats(days);

  /** Switch to a different year */
  function handleYearChange(year: number) {
    if (year === selectedYear) return;
    setSelectedYear(year);
    startTransition(async () => {
      const result = await fetchHeatmapDataAction(year);
      if (result.data) {
        setDays(result.data.days);
      }
    });
  }

  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-slate-200 bg-white p-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-teal-50">
          <CalendarDays className="h-5 w-5 text-teal-600" />
        </div>
        <div>
          <h3 className="text-base font-bold text-slate-900">学习热力图</h3>
          <p className="text-[13px] text-slate-500">
            {selectedYear} 年共学习 {stats.activeDays} 天 · 累计{" "}
            {stats.totalAnswers.toLocaleString()} 题
          </p>
        </div>
      </div>

      {/* Body: grid + summary + year selector */}
      <div className="flex gap-5">
        {/* Left: heatmap grid */}
        <div
          className={`min-w-0 flex-1 rounded-xl border border-slate-200 p-4 ${
            isPending ? "opacity-50" : ""
          }`}
        >
          <HeatmapGrid year={selectedYear} days={days} />
        </div>

        {/* Right: summary + year selector */}
        <div className="flex w-44 shrink-0 flex-col gap-4">
          {/* Year selector */}
          <div className="flex flex-col items-end gap-2">
            <span className="text-xs text-slate-400">年份</span>
            {years.map((year) => (
              <button
                key={year}
                onClick={() => handleYearChange(year)}
                disabled={isPending}
                className={`rounded-full border px-3.5 py-1 text-sm font-medium transition-colors ${
                  year === selectedYear
                    ? "border-teal-500 bg-teal-50 text-teal-700"
                    : "border-slate-200 text-slate-500 hover:border-slate-300"
                }`}
              >
                {year}
              </button>
            ))}
          </div>

          {/* Stats */}
          <HeatmapSummary
            activeDays={stats.activeDays}
            avgPerDay={stats.avgPerDay}
            tierCounts={stats.tierCounts}
          />
        </div>
      </div>
    </div>
  );
}
```

**Step 2: Commit**

```
feat: add learning heatmap container component
```

---

### Task 8: Integrate into hall home page

**Files:**
- Modify: `src/app/(web)/hall/(main)/(home)/page.tsx`

**Step 1: Check that `profile.createdAt` is available**

Read `src/models/user/user.query.ts` → `getUserProfile`. If `createdAt` is not in the `select`, add it.

**Step 2: Update the page**

Add import:

```ts
import { LearningHeatmap } from "@/features/web/hall/components/learning-heatmap";
```

Add the heatmap block below the main content `</div>`:

```tsx
{/* Learning heatmap */}
{data?.heatmap && (
  <LearningHeatmap
    initialYear={data.heatmap.year}
    initialDays={data.heatmap.days}
    accountYear={data.heatmap.accountYear}
  />
)}
```

**Step 3: Run dev server, verify the heatmap renders**

```bash
npm run dev
```

Open `http://localhost:3000/hall` and confirm:
- Heatmap grid renders with teal colors
- Year selector works
- Summary stats display correctly
- Empty state (no records) shows empty grid

**Step 4: Commit**

```
feat: integrate learning heatmap into hall home page
```

---

### Task 9: Visual polish and verify

**Step 1: Compare against the reference image**

Check:
- Teal color tiers match the theme
- Grid cell sizing and spacing look right
- Month labels align above correct columns
- Day labels (一 三 五) align to correct rows
- Legend (少 → 多) displays at bottom-left
- Year selector pills are styled correctly (selected = teal border + fill)
- Summary cards match the layout

**Step 2: Fix any visual issues**

Adjust spacing, colors, font sizes as needed.

**Step 3: Commit**

```
fix: polish heatmap visual styling
```

---

### Task 10: Build verification

**Step 1: Run lint**

```bash
npm run lint
```

Fix any issues.

**Step 2: Run build**

```bash
npm run build
```

Fix any type or build errors.

**Step 3: Commit if needed**

```
fix: resolve lint/build issues in heatmap
```
