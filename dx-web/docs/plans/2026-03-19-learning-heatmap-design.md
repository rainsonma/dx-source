# Learning Heatmap Design

## Overview

Add a 学习热力图 (Learning Heatmap) block to the hall home page. GitHub-style contribution grid showing daily learning activity by answer count, with year selector and summary stats.

## Data Source

**Table:** `game_records`

**Query:** Parameterized raw SQL via `$queryRaw` (tagged template literal — safe from injection):

```sql
SELECT DATE(created_at AT TIME ZONE 'UTC') as date, COUNT(*) as count
FROM game_records
WHERE user_id = $1
  AND created_at >= $2
  AND created_at < $3
GROUP BY DATE(created_at AT TIME ZONE 'UTC')
```

Returns `{ date, count }[]` — at most 365/366 rows per year.

## Intensity Tiers

| Tier | Range | Color |
|------|-------|-------|
| 0 | No activity | `bg-slate-100` |
| 1 | 1–10 answers | lightest teal |
| 2 | 11–30 answers | light teal |
| 3 | 31–60 answers | medium teal |
| 4 | 60+ answers | darkest teal |

## Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ 📅 学习热力图                                                    │
│ 2026 年共学习 46 天 · 累计 2,340 题                               │
├───────────────────────────────────┬──────────────┬──────────────┤
│                                  │  年度活跃     │    年份       │
│   Heatmap Grid                   │  活跃天数 46  │   ○ 2026     │
│   (53 cols × 7 rows)             │  日均 50 题   │   ○ 2025     │
│                                  ├──────────────┤   ○ 2024     │
│   Month labels: 1月 2月 ...      │  学习强度     │              │
│   Day labels: 一 三 五           │  1~10 题 · N天│              │
│                                  │  11~30 · N天  │              │
│   Legend: 少 ■■■■■ 多            │  31~60 · N天  │              │
│                                  │  60+ 题 · N天 │              │
└───────────────────────────────────┴──────────────┴──────────────┘
```

## Summary Metrics

All derived from `{ date, count }[]` in a helper:

- **活跃天数:** `days.length`
- **累计答题:** `sum(count)`
- **日均答题:** `sum(count) / days.length` (rounded)
- **Per-tier day counts:** count of days falling in each tier range

## Files

### Data Layer

| File | Purpose |
|------|---------|
| `src/models/game-record/game-record.query.ts` | Add `getUserDailyAnswerCounts(userId, year)` |
| `src/features/web/hall/services/hall.service.ts` | Add `fetchHeatmapData(year)` — auth, validate year, call query, map bigint→number |
| `src/features/web/hall/actions/hall.action.ts` | Add `fetchHeatmapDataAction(year)` — server action for client-side year switching |

### UI Layer

| File | Type | Purpose |
|------|------|---------|
| `src/features/web/hall/types/heatmap.ts` | Types | `HeatmapDay`, `HeatmapData`, tier constants |
| `src/features/web/hall/helpers/heatmap.ts` | Helper | Grid calculation, tier mapping, date math |
| `src/features/web/hall/components/learning-heatmap.tsx` | Client component | Container, year state, fetches on year change |
| `src/features/web/hall/components/heatmap-grid.tsx` | Pure component | CSS grid rendering, month/day labels, legend |
| `src/features/web/hall/components/heatmap-summary.tsx` | Pure component | Right panel: 年度活跃 + 学习强度 |

### Integration

| File | Change |
|------|--------|
| `src/features/web/hall/services/hall.service.ts` | `fetchDashboardStats()` adds heatmap data for current year |
| `src/app/(web)/hall/(main)/(home)/page.tsx` | Add `<LearningHeatmap />` below the main content row |

## Data Flow

**SSR (initial load):**
```
fetchDashboardStats()
  → fetchHeatmapData(currentYear)
    → getUserDailyAnswerCounts(userId, year)
      → $queryRaw (parameterized)
  → { days: { date, count }[], year }
```

**Client (year switch):**
```
LearningHeatmap onClick(newYear)
  → fetchHeatmapDataAction(newYear)
  → re-render grid + summary
```

## Year Range

Derived from user's `createdAt` (account creation year) through current year. Only show buttons for valid years.

## Security

- Raw SQL uses Prisma's `$queryRaw` tagged template — parameters are automatically escaped
- `userId` comes from authenticated session, never from user input
- Year is validated (clamped to account creation year through current year)
- No schema changes, no migrations

## Grid Rendering

- Pure CSS grid — no external library
- 53 columns (weeks) × 7 rows (Mon–Sun)
- Small rounded squares with teal color tiers
- Matches project's teal theme from `globals.css`
- Month labels at top, day labels (一 三 五) on left
- Color legend (少 → 多) at bottom-left
