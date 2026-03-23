# Leaderboard Real Data Design

## Overview

Wire real data to the leaderboard page (`/hall/leaderboard`), replacing hardcoded placeholder data. Rankings by EXP (经验) and play time (时长), with time-windowed boards (总榜, 日榜, 周榜, 月榜). Top 100 leaders, plus current user's rank always shown.

## Data Sources

### 总榜 (All-time)

- **经验:** `users.exp` — already aggregated globally per user
- **时长:** `SUM(game_stats_totals.total_play_time)` — aggregate across all games per user

### 日/周/月榜 (Time-windowed)

- **经验:** `SUM(game_session_totals.exp)` filtered by `ended_at` within window
- **时长:** `SUM(game_session_totals.play_time)` filtered by `ended_at` within window
- Users with zero activity in the window are excluded (`HAVING > 0`)

### My Rank

Same ranking query wrapped in a CTE, filtered to current user's row.

## Query Strategy

Raw SQL via `db.$queryRaw` (tagged template literals) for parameterized, injection-safe queries. `RANK()` window function computes ranks in one pass.

Column selection (`exp` vs `play_time`) uses `Prisma.sql` fragments — never string interpolation.

### Safety

- `$queryRaw` with tagged templates (auto-parameterized)
- `type` and `period` select pre-written query branches, never interpolated into SQL
- `userId` from server-side `auth()` session only
- Date window boundaries computed server-side
- Zod validation on incoming `type`/`period` params

## File Structure

```
src/features/web/leaderboard/
├── components/
│   ├── leaderboard-content.tsx    # Rewrite — manages tab state, fetches data
│   ├── leaderboard-podium.tsx     # Top 3 podium display
│   ├── leaderboard-list.tsx       # Ranks #4-#100 scrollable list
│   └── leaderboard-my-rank.tsx    # Current user's rank bar (always visible)
├── hooks/
│   └── use-leaderboard.ts         # Client state for type/period + data fetching
├── services/
│   └── leaderboard.service.ts     # Server-side raw SQL queries
├── actions/
│   └── leaderboard.action.ts      # Server actions wrapping service
├── schemas/
│   └── leaderboard.schema.ts      # Zod schemas for type/period
└── types/
    └── leaderboard.types.ts       # LeaderboardEntry, LeaderboardResult
```

No new model files — cross-model aggregations live in the leaderboard service.

## UI Behavior

### Tab defaults

- Type: **经验** (default) | 时长
- Period: **总榜** (default) | 日榜 | 周榜 | 月榜

### Components

- **LeaderboardContent:** `"use client"`, manages tab state via `useLeaderboard` hook. Fetches 经验 + 总榜 on mount. Shows loading skeleton during fetch.
- **LeaderboardMyRank:** Always visible. Teal border highlight. Shows rank, avatar placeholder, nickname (fallback username), value. No activity = rank "—", value 0.
- **LeaderboardPodium:** Top 3 with medal heights (gold/silver/bronze). Avatar placeholders.
- **LeaderboardList:** Ranks #4-#100, scrollable within card.

### Value formatting

- **经验:** comma-separated number via `toLocaleString()` (e.g., "15,860")
- **时长:** human-readable duration from seconds (e.g., "12h 30m", "45m", "< 1m")

### Icons

- 经验: Zap icon
- 时长: Clock icon

## Data Flow

1. Page loads → server action fetches 总榜 经验 (default)
2. User switches tab → hook calls server action with new `type`/`period`
3. Server validates params (Zod), computes date window, runs raw SQL
4. Returns `{ entries: LeaderboardEntry[], myRank: LeaderboardEntry | null }`
