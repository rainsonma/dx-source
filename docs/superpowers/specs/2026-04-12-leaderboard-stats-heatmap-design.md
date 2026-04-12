# Leaderboard Cleanup, Stats Scoping & Heatmap Index

Date: 2026-04-12

Three focused changes: remove 总榜 from leaderboard, scope 我的战绩 to last 30 days, and add a compound index for heatmap efficiency.

## 1. Remove 总榜 from Leaderboard

Remove the all-time leaderboard period from both frontend and backend. Default period changes to 月榜 (monthly).

### Frontend

| File | Change |
|------|--------|
| `features/web/leaderboard/types/leaderboard.types.ts` | Remove `"all"` from `LeaderboardPeriod` union |
| `features/web/leaderboard/schemas/leaderboard.schema.ts` | Remove `"all"` from Zod enum |
| `features/web/leaderboard/components/leaderboard-content.tsx` | Remove 总榜 entry from `PERIOD_TABS` |
| `features/web/leaderboard/hooks/use-leaderboard.ts` | Change initial `period` from `"all"` to `"month"` |

### Backend

| File | Change |
|------|--------|
| `app/http/controllers/api/leaderboard_controller.go` | Remove `"all"` from period validation |
| `app/services/api/leaderboard_service.go` | Remove `getAllTimeExp()`, `getAllTimePlayTime()`, and the `period == "all"` branch in `GetLeaderboard()` |

## 2. Scope 我的战绩 to Last 30 Days

Add a rolling 30-day filter to game stats and display a badge indicating the scope.

### Frontend

| File | Change |
|------|--------|
| `features/web/games/components/my-stats-card.tsx` | Add `近一月` teal pill badge to the right end of the 我的战绩 title row |

No label changes. The 4 stats remain: 最高得分, 已玩次数, 累计得分, 总经验值.

### Backend

| File | Change |
|------|--------|
| `app/services/api/game_stats_service.go` | Add `AND ended_at >= NOW() - INTERVAL '30 days'` to WHERE clause |

The SQL becomes:
```sql
SELECT
  COUNT(*)::int AS total_sessions,
  COALESCE(MAX(score), 0)::int AS highest_score,
  COALESCE(SUM(score), 0)::int AS total_scores,
  COALESCE(SUM(exp), 0)::int AS total_exp,
  COALESCE(SUM(play_time), 0)::int AS total_play_time,
  COUNT(*)::int AS completion_count,
  EXTRACT(EPOCH FROM MIN(ended_at))::bigint AS first_completed
FROM game_sessions
WHERE user_id = ? AND game_id = ? AND ended_at IS NOT NULL
  AND ended_at >= NOW() - INTERVAL '30 days'
```

No API shape change — same `GameStatsData` struct, same endpoint.

## 3. Heatmap Compound Index

### Problem

The heatmap query on `game_records` filters by `user_id` and `created_at` range. Only `user_id` is indexed — PostgreSQL must scan all of a user's records to apply the date filter.

### Solution

Add a compound index directly in the existing migration `20260405000004_create_game_records_table.go` (DB will be reset fresh):

```go
table.Index("user_id", "created_at")
```

No query rewrite needed. The existing `TO_CHAR` grouping is fine — it operates on a small result set (max ~365 rows) after the index narrows the scan.

A standalone `created_at` index is not added — no current query filters by date alone, and `game_records` is high-write. Add it later if needed.

## Out of Scope

- No changes to leaderboard type tabs (时长/经验 remain)
- No changes to heatmap frontend components
- No new API endpoints
