# Leaderboard Cleanup, Stats Scoping & Heatmap Index — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove 总榜 from leaderboard (default to 月榜), scope 我的战绩 to rolling 30 days with a badge, and add a compound index for heatmap query efficiency.

**Architecture:** Three independent changes touching frontend (Next.js) and backend (Go/Goravel). No new endpoints, no API shape changes. One migration file modified (DB will be reset fresh).

**Tech Stack:** Next.js 16, React, TypeScript, Zod, TailwindCSS, Go, Goravel, PostgreSQL, GORM

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `dx-web/src/features/web/leaderboard/types/leaderboard.types.ts` | Remove `"all"` from period union |
| Modify | `dx-web/src/features/web/leaderboard/schemas/leaderboard.schema.ts` | Remove `"all"` from Zod enum |
| Modify | `dx-web/src/features/web/leaderboard/components/leaderboard-content.tsx` | Remove 总榜 tab |
| Modify | `dx-web/src/features/web/leaderboard/hooks/use-leaderboard.ts` | Default period → `"month"` |
| Modify | `dx-web/src/features/web/leaderboard/actions/leaderboard.action.ts` | No change needed (passes through) |
| Modify | `dx-web/src/lib/api-client.ts` | Update leaderboardApi JSDoc comment |
| Modify | `dx-api/app/http/controllers/api/leaderboard_controller.go` | Remove `"all"` from validation |
| Modify | `dx-api/app/services/api/leaderboard_service.go` | Remove all-time functions and branch |
| Modify | `dx-web/src/features/web/games/components/my-stats-card.tsx` | Add 近一月 badge |
| Modify | `dx-api/app/services/api/game_stats_service.go` | Add 30-day filter to WHERE clause |
| Modify | `dx-api/database/migrations/20260405000004_create_game_records_table.go` | Add compound index |

---

### Task 1: Remove 总榜 — Backend

**Files:**
- Modify: `dx-api/app/services/api/leaderboard_service.go:38-49` (remove all-time branch)
- Modify: `dx-api/app/services/api/leaderboard_service.go:51-89` (delete two functions)
- Modify: `dx-api/app/http/controllers/api/leaderboard_controller.go:25,30-31`

- [ ] **Step 1: Remove all-time functions and branch from service**

In `dx-api/app/services/api/leaderboard_service.go`, replace `GetLeaderboard` and delete `getAllTimeExp` + `getAllTimePlayTime`:

```go
// GetLeaderboard returns a ranked list by type (exp|playtime) and period (day|week|month).
func GetLeaderboard(lbType, period, userID string) (*LeaderboardResult, error) {
	if lbType == "exp" {
		return getWindowedExp(period, userID)
	}
	return getWindowedPlayTime(period, userID)
}
```

Delete the entire `getAllTimeExp` function (lines 51-67) and `getAllTimePlayTime` function (lines 69-89).

- [ ] **Step 2: Update controller validation**

In `dx-api/app/http/controllers/api/leaderboard_controller.go`, change default period and validation:

```go
	period := ctx.Request().Query("period", "month")
```

```go
	if period != "day" && period != "week" && period != "month" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "时间范围必须是日、周或月")
	}
```

- [ ] **Step 3: Verify backend compiles**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/leaderboard_service.go dx-api/app/http/controllers/api/leaderboard_controller.go
git commit -m "refactor(api): remove all-time leaderboard period"
```

---

### Task 2: Remove 总榜 — Frontend

**Files:**
- Modify: `dx-web/src/features/web/leaderboard/types/leaderboard.types.ts:2`
- Modify: `dx-web/src/features/web/leaderboard/schemas/leaderboard.schema.ts:5`
- Modify: `dx-web/src/features/web/leaderboard/components/leaderboard-content.tsx:22`
- Modify: `dx-web/src/features/web/leaderboard/hooks/use-leaderboard.ts:17`
- Modify: `dx-web/src/lib/api-client.ts:432`

- [ ] **Step 1: Update type**

In `dx-web/src/features/web/leaderboard/types/leaderboard.types.ts`, line 2:

```typescript
export type LeaderboardPeriod = "day" | "week" | "month";
```

- [ ] **Step 2: Update Zod schema**

In `dx-web/src/features/web/leaderboard/schemas/leaderboard.schema.ts`, line 5:

```typescript
  period: z.enum(["day", "week", "month"]),
```

- [ ] **Step 3: Remove 总榜 tab**

In `dx-web/src/features/web/leaderboard/components/leaderboard-content.tsx`, replace `PERIOD_TABS` (lines 21-26):

```typescript
const PERIOD_TABS: { label: string; value: LeaderboardPeriod }[] = [
  { label: "日榜", value: "day" },
  { label: "周榜", value: "week" },
  { label: "月榜", value: "month" },
];
```

- [ ] **Step 4: Change default period to month**

In `dx-web/src/features/web/leaderboard/hooks/use-leaderboard.ts`, line 17:

```typescript
  const [period, setPeriod] = useState<LeaderboardPeriod>("month");
```

- [ ] **Step 5: Update api-client JSDoc**

In `dx-web/src/lib/api-client.ts`, line 432:

```typescript
  /** Get leaderboard by type (exp|playtime) and period (day|week|month) */
```

- [ ] **Step 6: Verify frontend compiles**

Run: `cd dx-web && npx tsc --noEmit`
Expected: no type errors.

- [ ] **Step 7: Commit**

```bash
git add dx-web/src/features/web/leaderboard/ dx-web/src/lib/api-client.ts
git commit -m "refactor(web): remove 总榜 tab, default to 月榜"
```

---

### Task 3: Scope 我的战绩 to Last 30 Days — Backend

**Files:**
- Modify: `dx-api/app/services/api/game_stats_service.go:30-42`

- [ ] **Step 1: Add 30-day filter to SQL query**

In `dx-api/app/services/api/game_stats_service.go`, replace the raw SQL (lines 30-42):

```go
	err := facades.Orm().Query().Raw(
		`SELECT
			COUNT(*)::int AS total_sessions,
			COALESCE(MAX(score), 0)::int AS highest_score,
			COALESCE(SUM(score), 0)::int AS total_scores,
			COALESCE(SUM(exp), 0)::int AS total_exp,
			COALESCE(SUM(play_time), 0)::int AS total_play_time,
			COUNT(*)::int AS completion_count,
			EXTRACT(EPOCH FROM MIN(ended_at))::bigint AS first_completed
		FROM game_sessions
		WHERE user_id = ? AND game_id = ? AND ended_at IS NOT NULL
			AND ended_at >= NOW() - INTERVAL '30 days'`,
		userID, gameID,
	).Scan(&result)
```

- [ ] **Step 2: Verify backend compiles**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/game_stats_service.go
git commit -m "feat(api): scope game stats to rolling 30 days"
```

---

### Task 4: Add 近一月 Badge — Frontend

**Files:**
- Modify: `dx-web/src/features/web/games/components/my-stats-card.tsx:10-13`

- [ ] **Step 1: Add badge to title row**

In `dx-web/src/features/web/games/components/my-stats-card.tsx`, replace the title row (lines 10-13):

```tsx
      <div className="flex items-center gap-2">
        <TrendingUp className="h-[18px] w-[18px] text-teal-600" />
        <h3 className="text-base font-bold text-foreground">我的战绩</h3>
        <span className="ml-auto rounded-full bg-teal-100 px-2 py-0.5 text-xs font-medium text-teal-700 dark:bg-teal-900/30 dark:text-teal-400">
          近一月
        </span>
      </div>
```

- [ ] **Step 2: Verify frontend compiles**

Run: `cd dx-web && npx tsc --noEmit`
Expected: no type errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/games/components/my-stats-card.tsx
git commit -m "feat(web): add 近一月 badge to 我的战绩 card"
```

---

### Task 5: Add Heatmap Compound Index

**Files:**
- Modify: `dx-api/database/migrations/20260405000004_create_game_records_table.go:35`

- [ ] **Step 1: Add compound index to migration**

In `dx-api/database/migrations/20260405000004_create_game_records_table.go`, add the following line after the existing `table.Index("is_correct")` (line 35):

```go
		table.Index("user_id", "created_at")
```

- [ ] **Step 2: Verify backend compiles**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors.

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260405000004_create_game_records_table.go
git commit -m "perf(api): add compound index on game_records(user_id, created_at)"
```

---

### Task 6: Manual Verification

- [ ] **Step 1: Start backend and frontend**

```bash
cd dx-api && air &
cd dx-web && npm run dev &
```

- [ ] **Step 2: Verify leaderboard**

Open `http://localhost:3000/hall/leaderboard`:
- Confirm only 3 period tabs visible: 日榜, 周榜, 月榜
- Confirm page loads on 月榜 by default
- Switch between all tabs — each should load data without errors

- [ ] **Step 3: Verify 我的战绩**

Open any game detail page (e.g., `http://localhost:3000/hall/games/{id}`):
- Confirm 近一月 teal badge appears at the right end of the 我的战绩 title row
- Confirm 4 stats display (may be 0 if no recent sessions — that's correct)

- [ ] **Step 4: Verify heatmap**

Open `http://localhost:3000/hall`:
- Confirm 学习热力图 loads without errors
- Switch years — each should load correctly
