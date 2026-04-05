# Session Table Merge & Level-Based Game Play — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Merge game session tables into a single `game_sessions` table, remove stats tables, and make all game modes (single, PK, group) level-based with first-to-complete-wins for competitive modes.

**Architecture:** Replace `game_session_totals` + `game_session_levels` with one `game_sessions` table (1 row = 1 level play). Remove `game_stats_totals` + `game_stats_levels` entirely — derive all stats from `game_sessions`. PK and group play become single-level, first-to-complete wins. `game_records` FK simplified to single `game_session_id`.

**Tech Stack:** Go/Goravel backend, Next.js 16 frontend, PostgreSQL, Redis, SSE for real-time

**Spec:** `docs/superpowers/specs/2026-04-05-session-table-merge-design.md`

---

## Task 1: Database Migration + New Model

**Files:**
- Create: `dx-api/app/models/game_session.go`
- Delete: `dx-api/app/models/game_session_total.go`
- Delete: `dx-api/app/models/game_session_level.go`
- Delete: `dx-api/app/models/game_stats_total.go`
- Delete: `dx-api/app/models/game_stats_level.go`
- Modify: `dx-api/app/models/game_record.go`
- Modify: `dx-api/app/models/game_pk.go`
- Create: `dx-api/database/migrations/20260405000001_create_game_sessions_table.go`
- Modify: `dx-api/bootstrap/migrations.go`

- [ ] **Step 1: Create `game_session.go` model**

```go
package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameSession struct {
	orm.Timestamps
	ID                   string     `gorm:"column:id;primaryKey" json:"id"`
	UserID               string     `gorm:"column:user_id" json:"user_id"`
	GameID               string     `gorm:"column:game_id" json:"game_id"`
	GameLevelID          string     `gorm:"column:game_level_id" json:"game_level_id"`
	Degree               string     `gorm:"column:degree" json:"degree"`
	Pattern              *string    `gorm:"column:pattern" json:"pattern"`
	CurrentContentItemID *string    `gorm:"column:current_content_item_id" json:"current_content_item_id"`
	StartedAt            time.Time  `gorm:"column:started_at" json:"started_at"`
	LastPlayedAt         time.Time  `gorm:"column:last_played_at" json:"last_played_at"`
	EndedAt              *time.Time `gorm:"column:ended_at" json:"ended_at"`
	Score                int        `gorm:"column:score" json:"score"`
	Exp                  int        `gorm:"column:exp" json:"exp"`
	MaxCombo             int        `gorm:"column:max_combo" json:"max_combo"`
	CorrectCount         int        `gorm:"column:correct_count" json:"correct_count"`
	WrongCount           int        `gorm:"column:wrong_count" json:"wrong_count"`
	SkipCount            int        `gorm:"column:skip_count" json:"skip_count"`
	PlayTime             int        `gorm:"column:play_time" json:"play_time"`
	TotalItemsCount      int        `gorm:"column:total_items_count" json:"total_items_count"`
	PlayedItemsCount     int        `gorm:"column:played_items_count" json:"played_items_count"`
	GameGroupID          *string    `gorm:"column:game_group_id" json:"game_group_id"`
	GameSubgroupID       *string    `gorm:"column:game_subgroup_id" json:"game_subgroup_id"`
	GamePkID             *string    `gorm:"column:game_pk_id" json:"game_pk_id"`
}

func (g *GameSession) TableName() string {
	return "game_sessions"
}
```

- [ ] **Step 2: Update `game_record.go`** — replace two FK fields with one

Replace `GameSessionTotalID` and `GameSessionLevelID` with single `GameSessionID`:

```go
package models

import "github.com/goravel/framework/database/orm"

type GameRecord struct {
	orm.Timestamps
	ID            string `gorm:"column:id;primaryKey" json:"id"`
	UserID        string `gorm:"column:user_id" json:"user_id"`
	GameSessionID string `gorm:"column:game_session_id" json:"game_session_id"`
	GameLevelID   string `gorm:"column:game_level_id" json:"game_level_id"`
	ContentItemID string `gorm:"column:content_item_id" json:"content_item_id"`
	IsCorrect     bool   `gorm:"column:is_correct" json:"is_correct"`
	SourceAnswer  string `gorm:"column:source_answer" json:"source_answer"`
	UserAnswer    string `gorm:"column:user_answer" json:"user_answer"`
	BaseScore     int    `gorm:"column:base_score" json:"base_score"`
	ComboScore    int    `gorm:"column:combo_score" json:"combo_score"`
	Duration      int    `gorm:"column:duration" json:"duration"`
}

func (g *GameRecord) TableName() string {
	return "game_records"
}
```

- [ ] **Step 3: Update `game_pk.go`** — replace `current_level_id` with `game_level_id`

```go
package models

import "github.com/goravel/framework/database/orm"

type GamePk struct {
	orm.Timestamps
	ID              string  `gorm:"column:id;primaryKey" json:"id"`
	UserID          string  `gorm:"column:user_id" json:"user_id"`
	OpponentID      string  `gorm:"column:opponent_id" json:"opponent_id"`
	GameID          string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID     string  `gorm:"column:game_level_id" json:"game_level_id"`
	Degree          string  `gorm:"column:degree" json:"degree"`
	Pattern         *string `gorm:"column:pattern" json:"pattern"`
	RobotDifficulty string  `gorm:"column:robot_difficulty" json:"robot_difficulty"`
	IsPlaying       bool    `gorm:"column:is_playing" json:"is_playing"`
	LastWinnerID    *string `gorm:"column:last_winner_id" json:"last_winner_id"`
}

func (g *GamePk) TableName() string {
	return "game_pks"
}
```

- [ ] **Step 4: Delete old model files**

Delete these 4 files:
- `dx-api/app/models/game_session_total.go`
- `dx-api/app/models/game_session_level.go`
- `dx-api/app/models/game_stats_total.go`
- `dx-api/app/models/game_stats_level.go`

- [ ] **Step 5: Create migration `20260405000001_create_game_sessions_table.go`**

This migration:
1. Drops old tables: `game_stats_levels`, `game_stats_totals`, `game_session_levels`, `game_session_totals`
2. Creates new `game_sessions` table
3. Recreates `game_records` table with `game_session_id` FK
4. Recreates `game_pks` table with `game_level_id` replacing `current_level_id`

Use the Goravel schema builder. Follow existing migration patterns in the codebase (see any existing migration for the structure). Register the new migration in `dx-api/bootstrap/migrations.go`.

The `game_sessions` table columns match the GameSession struct from Step 1. Add these raw SQL indexes after table creation:

```sql
CREATE UNIQUE INDEX idx_game_sessions_active_single ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, '')) WHERE ended_at IS NULL AND game_group_id IS NULL AND game_pk_id IS NULL;
CREATE UNIQUE INDEX idx_game_sessions_active_group ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_group_id) WHERE ended_at IS NULL AND game_group_id IS NOT NULL;
CREATE UNIQUE INDEX idx_game_sessions_active_pk ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_pk_id) WHERE ended_at IS NULL AND game_pk_id IS NOT NULL;
CREATE INDEX idx_game_sessions_group ON game_sessions (game_group_id) WHERE game_group_id IS NOT NULL;
CREATE INDEX idx_game_sessions_pk ON game_sessions (game_pk_id) WHERE game_pk_id IS NOT NULL;
CREATE INDEX idx_game_sessions_leaderboard ON game_sessions (user_id, last_played_at);
CREATE INDEX idx_game_sessions_user_game ON game_sessions (user_id, game_id);
```

- [ ] **Step 6: Verify compilation**

Run: `cd dx-api && go build ./...`

This will fail with many errors because services still reference old models. That's expected — we'll fix those in subsequent tasks. The goal here is just to verify the new model and migration compile.

To verify just the models and migration compile, temporarily comment out imports in files that reference the deleted models, or just confirm the new files are syntactically correct.

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/models/ dx-api/database/migrations/ dx-api/bootstrap/migrations.go
git commit -m "feat: create game_sessions model and migration, delete old session/stats models"
```

---

## Task 2: Delete Stats Service + Rewrite Game Stats Service

**Files:**
- Delete: `dx-api/app/services/api/stats_service.go`
- Modify: `dx-api/app/services/api/game_stats_service.go`

- [ ] **Step 1: Delete `stats_service.go`**

Delete `dx-api/app/services/api/stats_service.go` entirely. All 6 functions (`UpsertGameStats`, `UpsertLevelStats`, `completeLevelStatsInTx`, `updateGameStatsOnLevelCompleteInTx`, `UpdateGameStatsAfterSession`, `MarkGameFirstCompletion`) are no longer needed.

- [ ] **Step 2: Rewrite `game_stats_service.go`** — derive from `game_sessions`

The current `GetGameStats()` queries `game_stats_totals`. Replace with aggregation from `game_sessions`:

```go
package api

import (
	"github.com/goravel/framework/facades"
)

type GameStatsData struct {
	HighestScore    int    `json:"highestScore"`
	TotalSessions   int    `json:"totalSessions"`
	TotalScores     int    `json:"totalScores"`
	TotalExp        int    `json:"totalExp"`
	TotalPlayTime   int    `json:"totalPlayTime"`
	CompletionCount int    `json:"completionCount"`
	FirstCompleted  *int64 `json:"firstCompleted"`
}

func GetGameStats(userID, gameID string) (*GameStatsData, error) {
	var result struct {
		TotalSessions   int    `gorm:"column:total_sessions"`
		HighestScore    int    `gorm:"column:highest_score"`
		TotalScores     int    `gorm:"column:total_scores"`
		TotalExp        int    `gorm:"column:total_exp"`
		TotalPlayTime   int    `gorm:"column:total_play_time"`
		CompletionCount int    `gorm:"column:completion_count"`
		FirstCompleted  *int64 `gorm:"column:first_completed"`
	}

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
		WHERE user_id = ? AND game_id = ? AND ended_at IS NOT NULL`,
		userID, gameID,
	).Scan(&result)
	if err != nil {
		return nil, err
	}

	return &GameStatsData{
		HighestScore:    result.HighestScore,
		TotalSessions:   result.TotalSessions,
		TotalScores:     result.TotalScores,
		TotalExp:        result.TotalExp,
		TotalPlayTime:   result.TotalPlayTime,
		CompletionCount: result.CompletionCount,
		FirstCompleted:  result.FirstCompleted,
	}, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add -A dx-api/app/services/api/stats_service.go dx-api/app/services/api/game_stats_service.go
git commit -m "feat: delete stats_service, derive game stats from game_sessions"
```

---

## Task 3: Rewrite Leaderboard Service

**Files:**
- Modify: `dx-api/app/services/api/leaderboard_service.go`

- [ ] **Step 1: Rewrite leaderboard queries**

All 4 leaderboard functions now query `game_sessions` instead of `game_stats_totals` / `game_session_totals`. Remove the user's own rank logic — if not in top 100, frontend shows "未上榜".

Key changes:
- `getAllTimeExp()`: Keep as-is (already queries `users.exp` directly, not stats tables)
- `getAllTimePlayTime()`: Change from joining `game_stats_totals` to joining `game_sessions`
- `getWindowedExp()`: Change from joining `game_session_totals` to joining `game_sessions`
- `getWindowedPlayTime()`: Same change
- All functions: Remove the `OR id = ?` clause from the WHERE. Just `WHERE rank <= 100`.
- Remove `myRank` computation. Return `nil` for `MyRank` if user not in top 100.

For `getAllTimePlayTime`, replace the SQL:
```sql
-- Old: JOIN game_stats_totals
-- New:
WITH ranked AS (
  SELECT u.id, u.username, u.nickname, u.avatar_id,
         COALESCE(SUM(s.play_time), 0)::int AS value,
         RANK() OVER (ORDER BY COALESCE(SUM(s.play_time), 0) DESC)::int AS rank
  FROM users u
  INNER JOIN game_sessions s ON s.user_id = u.id
  WHERE u.is_active = true
  GROUP BY u.id, u.username, u.nickname, u.avatar_id
  HAVING COALESCE(SUM(s.play_time), 0) > 0
)
SELECT * FROM ranked WHERE rank <= 100 ORDER BY rank
```

For windowed queries, replace `game_session_totals` with `game_sessions`:
```sql
INNER JOIN game_sessions s ON s.user_id = u.id
  AND s.last_played_at >= $1 AND s.last_played_at < NOW()
```

For all queries: scan results into `[]LeaderboardEntry`. Find user's entry in the results — if present, set as `MyRank`; if absent, `MyRank` is nil.

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/services/api/leaderboard_service.go
git commit -m "feat: leaderboard queries from game_sessions, top 100 only"
```

---

## Task 4: Rewrite Hall Service + Game Service

**Files:**
- Modify: `dx-api/app/services/api/hall_service.go`
- Modify: `dx-api/app/services/api/game_service.go`

- [ ] **Step 1: Update `hall_service.go` — `getSessionProgress()`**

The `getSessionProgress()` function currently queries `game_session_totals`. Change to query `game_sessions` and compute progress differently.

Replace the SQL in `getSessionProgress()`:
```sql
SELECT
  s.game_id,
  g.name AS game_name,
  g.mode AS game_mode,
  COUNT(DISTINCT s.game_level_id) FILTER (WHERE s.ended_at IS NOT NULL)::int AS completed_levels,
  (SELECT COUNT(*)::int FROM game_levels gl WHERE gl.game_id = s.game_id AND gl.is_active = true) AS total_levels,
  COALESCE(SUM(s.score), 0)::int AS score,
  COALESCE(SUM(s.exp), 0)::int AS exp,
  MAX(s.last_played_at) AS last_played_at,
  BOOL_AND(s.ended_at IS NOT NULL) AS all_ended
FROM game_sessions s
INNER JOIN games g ON g.id = s.game_id
WHERE s.user_id = ? AND s.game_group_id IS NULL AND s.game_pk_id IS NULL
GROUP BY s.game_id, g.name, g.mode
ORDER BY MAX(s.last_played_at) DESC
LIMIT 20
```

Update the `SessionProgressItem` struct to match: replace `PlayedLevelsCount`/`TotalLevelsCount` with `CompletedLevels`/`TotalLevels`. Remove `ID`, `Degree`, `Pattern`, `EndedAt` fields (no longer per-session, now per-game aggregated).

Note: `getTodayAnswerCount()` and `GetHeatmap()` query `game_records` directly and need no changes (game_records still exists, just with `game_session_id` instead of the two old FK fields — but these functions only use `user_id` and `created_at`).

- [ ] **Step 2: Update `game_service.go` — `GetPlayedGames()`**

Replace the query that reads `game_stats_totals` with a query on `game_sessions`:

```go
// GetPlayedGames returns recently played games for a user.
func GetPlayedGames(userID string) ([]PlayedGameData, error) {
	var results []struct {
		GameID       string    `gorm:"column:game_id"`
		LastPlayedAt time.Time `gorm:"column:last_played_at"`
	}

	if err := facades.Orm().Query().Raw(
		`SELECT DISTINCT ON (game_id) game_id, last_played_at
		 FROM game_sessions
		 WHERE user_id = ?
		 ORDER BY game_id, last_played_at DESC`,
		userID,
	).Scan(&results); err != nil {
		return nil, err
	}

	// Sort by last_played_at DESC and limit to 10
	sort.Slice(results, func(i, j int) bool {
		return results[i].LastPlayedAt.After(results[j].LastPlayedAt)
	})
	if len(results) > 10 {
		results = results[:10]
	}

	// ... fetch game details for each game_id (keep existing logic)
}
```

Remove the import of `models.GameStatsTotal` from this file.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/hall_service.go dx-api/app/services/api/game_service.go
git commit -m "feat: hall and game services derive data from game_sessions"
```

---

## Task 5: Rewrite Single-Play Service

**Files:**
- Modify: `dx-api/app/services/api/game_play_single_service.go`

This is the largest service change. Key structural changes:

- [ ] **Step 1: Update DTOs**

Replace all DTOs at the top of the file. Key changes:
- `StartSessionResult`: Remove `LevelID` (replaced by `GameLevelID`). Remove reference to `currentLevelId`.
- `ActiveSessionData`: Remove `CurrentLevelID`. Add `GameLevelID`.
- Delete `ActiveLevelSessionData` (no longer needed).
- Delete `StartLevelResult` (merged into StartSession).
- `SessionRestoreData`: Replace `Session`+`SessionLevel` with single flat struct.
- `RecordAnswerInput`: Replace `GameSessionTotalID`+`GameSessionLevelID` with single `GameSessionID`.
- `RecordSkipInput`: Replace `GameSessionTotalID` with `GameSessionID`.
- `EndSessionInput`: Simplify — remove `AllLevelsCompleted`, `GameID`.

Updated DTOs:

```go
type StartSessionResult struct {
	ID                   string    `json:"id"`
	GameLevelID          string    `json:"gameLevelId"`
	Degree               string    `json:"degree"`
	Pattern              *string   `json:"pattern"`
	Score                int       `json:"score"`
	Exp                  int       `json:"exp"`
	MaxCombo             int       `json:"maxCombo"`
	CorrectCount         int       `json:"correctCount"`
	WrongCount           int       `json:"wrongCount"`
	StartedAt            time.Time `json:"startedAt"`
	CurrentContentItemID *string   `json:"currentContentItemId"`
}

type ActiveSessionData struct {
	ID                   string  `json:"id"`
	GameLevelID          string  `json:"gameLevelId"`
	Degree               string  `json:"degree"`
	Pattern              *string `json:"pattern"`
	CurrentContentItemID *string `json:"currentContentItemId"`
}

type CompleteLevelResult struct {
	ExpEarned      int     `json:"expEarned"`
	Accuracy       float64 `json:"accuracy"`
	MeetsThreshold bool    `json:"meetsThreshold"`
	NextLevelID    *string `json:"nextLevelId"`
	NextLevelName  *string `json:"nextLevelName"`
}

type SessionRestoreData struct {
	Score        int `json:"score"`
	MaxCombo     int `json:"maxCombo"`
	CorrectCount int `json:"correctCount"`
	WrongCount   int `json:"wrongCount"`
	SkipCount    int `json:"skipCount"`
	PlayTime     int `json:"playTime"`
}

type RecordAnswerInput struct {
	GameSessionID string
	GameLevelID   string
	ContentItemID string
	IsCorrect     bool
	UserAnswer    string
	SourceAnswer  string
	BaseScore     int
	ComboScore    int
	Score         int
	MaxCombo      int
	PlayTime      int
	NextContentItemID *string
	Duration      int
}

type RecordSkipInput struct {
	GameSessionID     string
	GameLevelID       string
	PlayTime          int
	NextContentItemID *string
}

type EndSessionInput struct {
	Score        int
	Exp          int
	MaxCombo     int
	CorrectCount int
	WrongCount   int
	SkipCount    int
}
```

- [ ] **Step 2: Rewrite `StartSession()`**

Now takes `gameLevelID` as required parameter instead of optional `levelID`. Creates a `GameSession` instead of `GameSessionTotal`. No longer calls `UpsertGameStats()`.

```go
func StartSession(userID, gameID, gameLevelID, degree string, pattern *string) (*StartSessionResult, error) {
	query := facades.Orm().Query()

	// VIP guard: non-first levels require active VIP
	var firstLevel models.GameLevel
	if err := query.Where("game_id", gameID).Where("is_active", true).
		Order("\"order\" asc").First(&firstLevel); err != nil || firstLevel.ID == "" {
		return nil, ErrNoGameLevels
	}
	if gameLevelID != firstLevel.ID {
		if err := requireVip(userID); err != nil {
			return nil, err
		}
	}

	// Check for existing active session
	existing, err := findActiveSession(query, userID, gameLevelID, degree, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to check active session: %w", err)
	}
	if existing != nil {
		if _, err := query.Model(&models.GameSession{}).Where("id", existing.ID).
			Update("last_played_at", time.Now()); err != nil {
			return nil, fmt.Errorf("failed to touch session: %w", err)
		}
		return &StartSessionResult{
			ID: existing.ID, GameLevelID: existing.GameLevelID,
			Degree: existing.Degree, Pattern: existing.Pattern,
			Score: existing.Score, Exp: existing.Exp, MaxCombo: existing.MaxCombo,
			CorrectCount: existing.CorrectCount, WrongCount: existing.WrongCount,
			StartedAt: existing.StartedAt, CurrentContentItemID: existing.CurrentContentItemID,
		}, nil
	}

	// Count content items for this level+degree
	totalItems, err := countLevelItems(query, gameLevelID, degree)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := models.GameSession{
		ID: newID(), UserID: userID, GameID: gameID, GameLevelID: gameLevelID,
		Degree: degree, Pattern: pattern,
		TotalItemsCount: int(totalItems),
		StartedAt: now, LastPlayedAt: now,
	}
	if err := query.Create(&session); err != nil {
		// Concurrent create — return existing
		existing, findErr := findActiveSession(query, userID, gameLevelID, degree, pattern)
		if findErr != nil || existing == nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		return &StartSessionResult{
			ID: existing.ID, GameLevelID: existing.GameLevelID,
			Degree: existing.Degree, Pattern: existing.Pattern,
			StartedAt: existing.StartedAt,
		}, nil
	}

	return &StartSessionResult{
		ID: session.ID, GameLevelID: session.GameLevelID,
		Degree: session.Degree, Pattern: session.Pattern, StartedAt: session.StartedAt,
	}, nil
}
```

- [ ] **Step 3: Rewrite `findActiveSession()`**

Change from querying `game_session_totals` to `game_sessions`:

```go
func findActiveSession(query orm.Query, userID, gameLevelID, degree string, pattern *string) (*models.GameSession, error) {
	var session models.GameSession
	q := query.Where("user_id", userID).Where("game_level_id", gameLevelID).
		Where("degree", degree).Where("ended_at IS NULL").
		Where("game_group_id IS NULL").Where("game_pk_id IS NULL")

	if pattern != nil {
		q = q.Where("pattern", *pattern)
	} else {
		q = q.Where("pattern IS NULL")
	}

	if err := q.First(&session); err != nil || session.ID == "" {
		return nil, nil
	}
	return &session, nil
}
```

- [ ] **Step 4: Add `countLevelItems()` helper**

```go
func countLevelItems(query orm.Query, gameLevelID, degree string) (int64, error) {
	q := query.Model(&models.ContentItem{}).
		Where("game_level_id", gameLevelID).Where("is_active", true)

	allowedTypes, ok := consts.DegreeContentTypes[degree]
	if ok {
		q = q.WhereIn("content_type", allowedTypes)
	}

	return q.Count()
}
```

- [ ] **Step 5: Simplify remaining functions**

- `CheckActiveSession()`: Update to use `gameLevelID` instead of `gameID`, return `ActiveSessionData` with `GameLevelID`
- `CheckAnyActiveSession()`: Query `game_sessions` instead of `game_session_totals`
- Delete `CheckActiveLevelSession()` — no longer needed
- Delete `StartLevel()` — merged into `StartSession()`
- Delete `AdvanceLevel()` — no multi-level advancing
- `RecordAnswer()`: Update to use `GameSession` instead of `GameSessionTotal`+`GameSessionLevel`. Update one table instead of two. Create `GameRecord` with `GameSessionID` instead of two FK fields.
- `RecordSkip()`: Same simplification — update one `game_sessions` row
- `CompleteLevel()`: Simplified — sets `ended_at` on the `game_sessions` row, calculates exp, increments user exp. Add `NextLevelID`/`NextLevelName` to result by finding next active level by order.
- `EndSession()`: Simplified — just sets `ended_at` with final stats. No `UpdateGameStatsAfterSession` or `MarkGameFirstCompletion` calls.
- `RestoreSessionData()`: Returns flat `SessionRestoreData` from single `game_sessions` row
- `RestartLevel()`: Resets the `game_sessions` row (clear scores, reset position)
- `ForceCompleteSession()`: Sets `ended_at` on the session
- `SyncPlayTime()`: Updates single `game_sessions` row
- `UpdateContentItem()`: Updates single `game_sessions` row

- [ ] **Step 6: Add `findNextLevel()` helper**

Used by `CompleteLevel()` to return next level info:

```go
func findNextLevel(gameID, currentLevelID string) (*string, *string, error) {
	var currentLevel models.GameLevel
	if err := facades.Orm().Query().Where("id", currentLevelID).First(&currentLevel); err != nil || currentLevel.ID == "" {
		return nil, nil, nil
	}

	var nextLevel models.GameLevel
	if err := facades.Orm().Query().Where("game_id", gameID).Where("is_active", true).
		Where("\"order\" > ?", currentLevel.Order).
		Order("\"order\" asc").First(&nextLevel); err != nil || nextLevel.ID == "" {
		return nil, nil, nil
	}
	return &nextLevel.ID, &nextLevel.Name, nil
}
```

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/services/api/game_play_single_service.go
git commit -m "feat: rewrite single-play service for level-based game_sessions"
```

---

## Task 6: Update Single-Play Controller + Routes

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_play_single_controller.go`
- Modify: `dx-api/app/http/requests/api/session_request.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Update request structs**

In `session_request.go`:
- `StartSessionRequest`: Add `GameLevelID` as required field. Remove optional `LevelID`.
- Delete `CheckActiveLevelSessionRequest` — no longer needed
- Delete `StartLevelRequest` — merged into start session
- Delete `AdvanceLevelRequest` — no longer needed
- `RecordAnswerRequest`: Replace `GameSessionTotalId`+`GameSessionLevelId` with `GameSessionId`
- `RecordSkipRequest`: Replace `GameSessionTotalId` with `GameSessionId`
- `EndSessionRequest`: Remove `AllLevelsCompleted`, `GameID`

- [ ] **Step 2: Update controller**

In `game_play_single_controller.go`:
- `Start()`: Pass `GameLevelID` to `StartSession()`
- Delete `StartLevel()` method
- Delete `AdvanceLevel()` method
- Delete `CheckActiveLevel()` method
- `RecordAnswer()`: Map `GameSessionId` instead of two fields
- `RecordSkip()`: Map `GameSessionId`
- `CompleteLevel()`: Return `nextLevelId` and `nextLevelName` from result
- `End()`: Simplified input

- [ ] **Step 3: Update routes**

In `routes/api.go`, under `/play-single`:
- Remove: `sessions.Get("/active-level", ...)` — deleted
- Remove: `sessions.Post("/{id}/levels/start", ...)` — deleted
- Remove: `sessions.Post("/{id}/levels/{levelId}/advance", ...)` — deleted
- Keep all other routes

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/api/game_play_single_controller.go dx-api/app/http/requests/api/session_request.go dx-api/routes/api.go
git commit -m "feat: update single-play controller and routes for level-based sessions"
```

---

## Task 7: Rewrite PK Service + PK Winner Service

**Files:**
- Modify: `dx-api/app/services/api/game_play_pk_service.go`
- Modify: `dx-api/app/services/api/pk_winner_service.go`

- [ ] **Step 1: Rewrite PK DTOs**

Update result types:
- `PkStartResult`: Remove `RobotCompleted` (no multi-level resume). Add `GameLevelID`.
- `PkPlayerCompleteEvent`: Add `Score` field (for result display).
- Remove `PkNextLevelEvent` type.

- [ ] **Step 2: Rewrite `StartPk()`**

Now takes `gameLevelID` as required. Creates PK + session in one step:

```go
func StartPk(userID, gameID, gameLevelID, degree string, pattern *string, difficulty string) (*PkStartResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	query := facades.Orm().Query()

	// Check for existing active PK for this user/game/level
	var existingPk models.GamePk
	query.Where("user_id", userID).Where("game_id", gameID).
		Where("game_level_id", gameLevelID).Where("is_playing", true).First(&existingPk)
	if existingPk.ID != "" {
		// Return existing PK + session
		// ... (similar to current idempotent logic but using game_sessions)
	}

	// Verify game and level
	// ... (keep existing validation)

	// Find or create mock user
	mockUser, err := FindOrCreateMockUser()
	// ...

	// Count content items
	totalItems, err := countLevelItems(query, gameLevelID, degree)
	// ...

	// Create game_pks record
	pkID := newID()
	pk := models.GamePk{
		ID: pkID, UserID: userID, OpponentID: mockUser.ID,
		GameID: gameID, GameLevelID: gameLevelID,
		Degree: degree, Pattern: pattern,
		RobotDifficulty: difficulty, IsPlaying: true,
	}
	query.Create(&pk)

	// Create human player's session
	now := time.Now()
	session := models.GameSession{
		ID: newID(), UserID: userID, GameID: gameID, GameLevelID: gameLevelID,
		Degree: degree, Pattern: pattern,
		TotalItemsCount: int(totalItems),
		StartedAt: now, LastPlayedAt: now, GamePkID: &pkID,
	}
	query.Create(&session)

	// Spawn robot goroutine for this single level
	spawnRobotForLevel(pkID, gameID, gameLevelID, degree, pattern, mockUser.ID, difficulty, int(totalItems))

	return &PkStartResult{
		PkID: pkID, SessionID: session.ID,
		GameLevelID: gameLevelID,
		OpponentID: mockUser.ID, OpponentName: mockUserName(mockUser),
	}, nil
}
```

- [ ] **Step 3: Delete `StartPkLevel()`, simplify `CompletePkLevel()`**

- Delete `StartPkLevel()` — merged into `StartPk()`
- `CompletePkLevel()` → rename to `CompletePk()`. First to complete wins:
  1. Set `ended_at` on winner's `game_session`
  2. Calculate exp (same accuracy threshold logic)
  3. Broadcast `pk_player_complete` with winner info via SSE
  4. Force-end opponent's session (set `ended_at` with partial stats)
  5. Set `game_pks.is_playing = false`
  6. Cancel robot goroutine if human won
  7. Return `CompleteLevelResult` with `NextLevelID`/`NextLevelName`

- [ ] **Step 4: Rewrite `NextPkLevel()`**

Instead of advancing within a PK, create a new PK for the next level:

```go
func NextPkLevel(userID, pkID string) (*PkStartResult, error) {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil || pk.ID == "" {
		return nil, ErrPkNotFound
	}

	nextLevelID, _, err := findNextLevel(pk.GameID, pk.GameLevelID)
	if err != nil || nextLevelID == nil {
		return nil, fmt.Errorf("no next level available")
	}

	return StartPk(userID, pk.GameID, *nextLevelID, pk.Degree, pk.Pattern, pk.RobotDifficulty)
}
```

- [ ] **Step 5: Update `PkRecordAnswer()` and `PkRecordSkip()`**

Change from updating two tables (`game_session_totals` + `game_session_levels`) to updating one `game_sessions` row. Remove skip functionality for PK (skip button is disabled).

Actually, since skip is disabled in PK, `PkRecordSkip()` can be deleted entirely.

- [ ] **Step 6: Simplify robot goroutine**

`spawnRobotForLevel()`: Robot creates its own `game_session`, simulates answers, then calls the completion logic. On completion, broadcasts `pk_player_complete`. No multi-level robot lifecycle.

- [ ] **Step 7: Rewrite `pk_winner_service.go`**

`DeterminePkWinner()` is no longer needed as a separate function — the winner is the first to complete. The completion handler in `CompletePk()` handles this directly. Either delete the file or simplify to a helper that force-ends the loser's session.

```go
// ForceEndPkLoser sets ended_at on the loser's session with their partial stats.
func ForceEndPkLoser(pkID, winnerUserID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`UPDATE game_sessions SET ended_at = ?, updated_at = now()
		 WHERE game_pk_id = ? AND user_id != ? AND ended_at IS NULL`,
		now, pkID, winnerUserID,
	)
	return err
}
```

- [ ] **Step 8: Commit**

```bash
git add dx-api/app/services/api/game_play_pk_service.go dx-api/app/services/api/pk_winner_service.go
git commit -m "feat: rewrite PK service for level-based first-to-complete-wins"
```

---

## Task 8: Update PK Controller + Routes

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_play_pk_controller.go`
- Modify: `dx-api/app/http/requests/api/session_request.go` (PK-specific requests)
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Update PK request structs**

- `StartPkRequest`: Add `GameLevelID` as required
- Remove `StartPkLevelRequest` (merged into start)
- `PkRecordAnswerRequest`: Replace two session FKs with `GameSessionId`
- Delete `PkRecordSkipRequest` — skip is disabled in PK

- [ ] **Step 2: Update PK controller**

- `Start()`: Pass `GameLevelID` to `StartPk()`
- Delete `StartLevel()` method
- Delete `RecordSkip()` method — skip disabled in PK
- `CompleteLevel()`: Call `CompletePk()`, return result with `nextLevelId`
- `NextLevel()`: Call `NextPkLevel()`, return new PK result
- `RecordAnswer()`: Map single `GameSessionId`

- [ ] **Step 3: Update PK routes**

In `routes/api.go`, under `/play-pk`:
- Remove: `pk.Post("/{id}/levels/start", ...)` — merged into start
- Remove: `pk.Post("/{id}/skips", ...)` — skip disabled
- Keep: `pk.Post("/{id}/next-level", ...)` — now creates new PK

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/api/game_play_pk_controller.go dx-api/app/http/requests/api/session_request.go dx-api/routes/api.go
git commit -m "feat: update PK controller and routes for level-based play"
```

---

## Task 9: Rewrite Group Play + Group Winner + Group Game Services

**Files:**
- Modify: `dx-api/app/services/api/game_play_group_service.go`
- Modify: `dx-api/app/services/api/group_winner_service.go`
- Modify: `dx-api/app/services/api/group_game_service.go`

- [ ] **Step 1: Rewrite group play DTOs**

- `GroupPlayStartSessionResult`: Remove `LevelID` (now uses `GameLevelID`). Remove `CurrentContentItemID`.
- `GroupPlayStartLevelResult`: Delete (merged into start session)
- `GroupPlayerCompleteEvent`: Add `Score` field

- [ ] **Step 2: Rewrite `GroupPlayStartSession()`**

Now creates a `GameSession` directly. Takes `gameLevelID` as required. No longer calls `UpsertGameStats()`.

Key logic: verify user is group member, lookup subgroup for team mode, create `game_sessions` row with `game_group_id` and optional `game_subgroup_id`.

- [ ] **Step 3: Delete `GroupPlayStartLevel()` — merged into start session**

- [ ] **Step 4: Rewrite `GroupPlayCompleteLevel()` — first-to-complete wins**

When a player completes:
1. Set `ended_at` on their `game_sessions` row
2. Calculate exp
3. Call `DetermineGroupWinner()` — since first-to-complete wins, this is trivial
4. Broadcast `group_player_complete` with winner info to all group SSE clients
5. Force-end all other players' sessions
6. Return `CompleteLevelResult` with `NextLevelID`/`NextLevelName`

- [ ] **Step 5: Remove skip from group play**

Delete `GroupPlayRecordSkip()` — skip is disabled in group play.

- [ ] **Step 6: Update `GroupPlayRecordAnswer()`**

Change from updating two tables to updating single `game_sessions` row. Remove skip tracking.

- [ ] **Step 7: Simplify `GroupPlaySyncPlayTime()`, `GroupPlayRestoreSessionData()`, `GroupPlayUpdateContentItem()`**

All now operate on single `game_sessions` row instead of two tables.

- [ ] **Step 8: Rewrite `group_winner_service.go`**

Simplify dramatically. First-to-complete wins:

```go
// ForceEndGroupLosers sets ended_at on all sessions for this group level except the winner.
func ForceEndGroupLosers(gameGroupID, gameLevelID, winnerUserID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`UPDATE game_sessions SET ended_at = ?, updated_at = now()
		 WHERE game_group_id = ? AND game_level_id = ? AND user_id != ? AND ended_at IS NULL`,
		now, gameGroupID, gameLevelID, winnerUserID,
	)
	return err
}

// For team mode: ForceEndGroupLosersExceptTeam
func ForceEndGroupLosersExceptTeam(gameGroupID, gameLevelID, winnerSubgroupID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`UPDATE game_sessions SET ended_at = ?, updated_at = now()
		 WHERE game_group_id = ? AND game_level_id = ? AND (game_subgroup_id IS NULL OR game_subgroup_id != ?) AND ended_at IS NULL`,
		now, gameGroupID, gameLevelID, winnerSubgroupID,
	)
	return err
}
```

Delete `CheckAndDetermineWinner()`, `DetermineWinnerForLevel()`, `determineSoloWinner()`, `determineTeamWinner()` — all replaced by first-to-complete logic.

- [ ] **Step 9: Update `group_game_service.go`**

- `StartGroupGame()`: Now force-ends any existing active sessions for the group, then broadcasts start event. Members create their own sessions.
- `ForceEndGroupGame()`: Update to end `game_sessions` WHERE `game_group_id` AND `ended_at IS NULL`
- `NextGroupLevel()`: Rewrite — find next level by order, update `start_game_level_id` on `game_groups`, broadcast `group_next_level` SSE event. Any member can call this now, not just owner.

- [ ] **Step 10: Commit**

```bash
git add dx-api/app/services/api/game_play_group_service.go dx-api/app/services/api/group_winner_service.go dx-api/app/services/api/group_game_service.go
git commit -m "feat: rewrite group services for level-based first-to-complete-wins"
```

---

## Task 10: Update Group Controller + Routes

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_play_group_controller.go`
- Modify: `dx-api/app/http/controllers/api/group_game_controller.go` (if NextLevel is here)
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Update group play request structs**

- `GroupPlayStartRequest`: Add `GameLevelID` as required
- Delete `GroupPlayStartLevelRequest` — merged
- `GroupRecordAnswerRequest`: Replace two session FKs with `GameSessionId`
- Delete `GroupRecordSkipRequest` — skip disabled

- [ ] **Step 2: Update group play controller**

- `Start()`: Pass `GameLevelID` to `GroupPlayStartSession()`
- Delete `StartLevel()` method
- Delete `RecordSkip()` method
- `CompleteLevel()`: Return result with `nextLevelId`

- [ ] **Step 3: Update `NextLevel` in group game controller**

Change from owner-only to any member. Remove the ownership check. The endpoint finds next level automatically.

- [ ] **Step 4: Update group routes**

In `routes/api.go`:
- Under `/play-group`: remove `/{id}/levels/start` and `/{id}/skips`
- Under `/groups/{id}`: keep `next-level` route but it now accepts any member

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/http/controllers/api/game_play_group_controller.go dx-api/app/http/controllers/api/ dx-api/routes/api.go
git commit -m "feat: update group controller and routes for level-based play"
```

---

## Task 11: Backend Build Verification

- [ ] **Step 1: Fix any remaining compilation errors**

Run: `cd dx-api && go build ./...`

Fix any remaining references to:
- `GameSessionTotal` / `GameSessionLevel` — replace with `GameSession`
- `GameStatsTotal` / `GameStatsLevel` — remove
- `game_session_total_id` / `game_session_level_id` — replace with `game_session_id`
- Deleted functions like `UpsertGameStats`, `UpsertLevelStats`, etc.

Check: `game_stats_controller.go` still compiles (it calls `GetGameStats()` which was rewritten in Task 2).

- [ ] **Step 2: Run linter**

Run: `cd dx-api && go vet ./...`

Fix any lint issues.

- [ ] **Step 3: Commit fixes**

```bash
git add dx-api/
git commit -m "fix: resolve all build and lint errors after session table merge"
```

---

## Task 12: Frontend Single-Play Updates

**Files:**
- Modify: `dx-web/src/features/web/play-single/actions/session.action.ts`
- Modify: `dx-web/src/features/web/play-core/hooks/use-game-store.ts`

- [ ] **Step 1: Update single-play action types and calls**

In `session.action.ts`:
- `startSessionAction()`: Add `gameLevelId` parameter. Remove `levelId` from URL params. New API: `POST /api/play-single/start` with body `{ gameId, gameLevelId, degree, pattern }`
- Delete `checkActiveLevelSessionAction()` — no longer exists
- Delete `startSessionLevelAction()` — merged into startSession
- Delete `advanceSessionLevelAction()` — no more advancing
- `recordAnswerAction()`: Replace `gameSessionTotalId`+`gameSessionLevelId` with `gameSessionId`
- `recordSkipAction()`: Replace `gameSessionTotalId` with `gameSessionId`
- `completeLevelAction()`: Response now includes `nextLevelId` and `nextLevelName`
- `endSessionAction()`: Simplified body — remove `allLevelsCompleted`, `gameId`
- `fetchSessionRestoreDataAction()`: Returns flat object (no `session`/`sessionLevel` nesting)

- [ ] **Step 2: Update game store**

In `use-game-store.ts`:
- Remove `levelSessionId` from state and `initSession` — there's only one session ID now
- The `sessionId` IS the level session (they're the same thing now)
- Update `initSession` data type to remove `levelSessionId`
- Update `resetGame` to not clear `levelSessionId`

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-single/ dx-web/src/features/web/play-core/hooks/use-game-store.ts
git commit -m "feat: update frontend single-play for level-based sessions"
```

---

## Task 13: Frontend PK Updates

**Files:**
- Modify: `dx-web/src/features/web/play-pk/actions/session.action.ts`
- Modify: `dx-web/src/features/web/play-pk/hooks/use-pk-play-store.ts`
- Modify: `dx-web/src/features/web/play-pk/types/pk-play.ts`
- Modify: PK play shell and result panel components

- [ ] **Step 1: Update PK types**

In `pk-play.ts`:
- Remove `PkNextLevelEvent` type
- `PkPlayerCompleteEvent`: Add `score: number` field
- `PkLevelCompleteEvent`: Simplify — first-to-complete, no score comparison

- [ ] **Step 2: Update PK store**

In `use-pk-play-store.ts`:
- Remove `levelSessionId` from state — only `sessionId` exists
- Remove `skipCount` from state (skip disabled in PK)
- Remove `recordSkip` action
- Remove `setPkWaiting` — no waiting phase, result is immediate
- Remove `pkPhase: "waiting"` — only `"playing"` and `"result"`
- `initSession`: Remove `levelSessionId` param
- When opponent completes (SSE `pk_player_complete`), immediately set `pkPhase: "result"`

- [ ] **Step 3: Update PK actions**

In `session.action.ts`:
- `startPkAction()`: Add `gameLevelId` parameter
- Delete `startLevelAction()` — merged into startPk
- Delete `recordSkipAction()` — skip disabled
- Delete `nextLevelAction()` — replaced by creating new PK
- Add `nextPkLevelAction(pkId)` — calls `POST /api/play-pk/{id}/next-level`, returns new PK start result
- `recordAnswerAction()`: Replace two session IDs with single `gameSessionId`
- `completeLevelAction()`: Response includes `nextLevelId`

- [ ] **Step 4: Update PK play shell**

In the PK play shell component:
- Hide skip button entirely
- Hide answer/reveal button entirely
- When SSE `pk_player_complete` arrives (opponent completed first), immediately show result panel — disable answer buttons
- "Next Level" button on result panel calls `nextPkLevelAction()`

- [ ] **Step 5: Update PK result panel**

In `pk-play-result-panel.tsx`:
- Show winner (first to complete) and their score
- Show loser's partial progress
- "Next Level" button if `nextLevelId` exists
- "End PK" button always

- [ ] **Step 6: Update PK SSE event handler**

In `use-pk-play-events.ts`:
- Remove handler for `pk_next_level` event
- Remove handler for `pk_level_complete` event (merged into `pk_player_complete`)
- `pk_player_complete` handler: immediately set pkPhase to "result" and show result panel

- [ ] **Step 7: Commit**

```bash
git add dx-web/src/features/web/play-pk/
git commit -m "feat: update frontend PK for level-based first-to-complete play"
```

---

## Task 14: Frontend Group Updates

**Files:**
- Modify: `dx-web/src/features/web/play-group/actions/session.action.ts`
- Modify: `dx-web/src/features/web/play-group/hooks/use-group-play-store.ts`
- Modify: `dx-web/src/features/web/play-group/types/group-play.ts`
- Modify: Group play shell and result panel components

- [ ] **Step 1: Update group types**

In `group-play.ts`:
- Remove `GroupNextLevelEvent` type
- `GroupPlayerCompleteEvent`: Add `score: number` field
- `GroupLevelCompleteEvent`: Simplify for first-to-complete

- [ ] **Step 2: Update group store**

In `use-group-play-store.ts`:
- Remove `levelSessionId` from state
- Remove `skipCount` from state (skip disabled)
- Remove `recordSkip` action
- Remove `setGroupWaiting` — no waiting, result is immediate
- Remove `groupPhase: "waiting"` — only `"playing"` and `"result"`
- When any player completes (SSE `group_player_complete` with winner), immediately set `groupPhase: "result"`

- [ ] **Step 3: Update group actions**

In `session.action.ts`:
- `startSessionAction()`: Add `gameLevelId` parameter
- Delete `startLevelAction()` — merged
- Delete `recordSkipAction()` — skip disabled
- `recordAnswerAction()`: Replace two session IDs with single `gameSessionId`
- `completeLevelAction()`: Response includes `nextLevelId`
- Add `nextGroupLevelAction(groupId)` — calls `POST /api/groups/{id}/next-level`

- [ ] **Step 4: Update group play shell**

- Hide skip button entirely
- Hide answer/reveal button entirely
- When SSE `group_player_complete` arrives and current user is NOT the winner, immediately show result — disable answer buttons

- [ ] **Step 5: Update group result panel**

In `group-play-result-panel.tsx`:
- Show winner (first to complete) — podium with just 1st place
- Show all other players' partial progress
- "Next Level" button if `nextLevelId` exists — any member can click

- [ ] **Step 6: Update group SSE event handler**

In `use-group-play-events.ts`:
- Remove handler for `group_next_level` event
- `group_player_complete` handler: if this is the winning event, immediately show result panel

- [ ] **Step 7: Commit**

```bash
git add dx-web/src/features/web/play-group/
git commit -m "feat: update frontend group for level-based first-to-complete play"
```

---

## Task 15: Frontend Hall + Leaderboard Updates

**Files:**
- Modify: `dx-web/src/features/web/hall/components/game-progress-card.tsx`
- Modify: `dx-web/src/features/web/hall/actions/` (if progress types change)
- Modify: `dx-web/src/features/web/leaderboard/` (types and components)
- Modify: `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx`

- [ ] **Step 1: Update progress card**

In `game-progress-card.tsx`:
- Update `SessionProgressItem` type to match new API response:
  - Replace `playedLevelsCount`/`totalLevelsCount` with `completedLevels`/`totalLevels`
  - Remove `id`, `degree`, `pattern`, `endedAt` (now per-game aggregated)
- Update progress calculation: `(completedLevels / totalLevels) * 100`
- Progress is now per-game (how many levels completed) not per-session

- [ ] **Step 2: Update leaderboard — "未上榜" for unranked users**

In leaderboard types (`leaderboard.types.ts`):
- `LeaderboardResult.myRank` is now nullable
- When `myRank` is null, display "未上榜" instead of a rank number

In leaderboard component:
- Check if `myRank` is null → show "未上榜" text
- Otherwise show rank as before

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/hall/ dx-web/src/features/web/leaderboard/
git commit -m "feat: update hall progress and leaderboard for session merge"
```

---

## Task 16: Frontend Play Shell — Skip/Answer Button Visibility

**Files:**
- Modify: Play shell components for all game modes (the components that render game UI during play)

- [ ] **Step 1: Identify skip/answer button locations**

Search for skip and answer button components across:
- `dx-web/src/features/web/play-core/`
- `dx-web/src/features/web/play-single/`
- `dx-web/src/features/web/play-pk/`
- `dx-web/src/features/web/play-group/`

Find the game mode components (word-sentence, vocab-battle, vocab-match, vocab-elimination, listening-challenge) and identify where skip and answer buttons are rendered.

- [ ] **Step 2: Add mode-aware visibility**

In each game mode component that renders skip/answer buttons:
- Accept a `competitive` prop (or read game mode from store)
- When `competitive === true` (PK or group mode): hide skip button and answer/reveal button
- When `competitive === false` (single-play): show both as normal

The PK and group play shells should pass `competitive={true}` to the shared game components. Single-play shells pass `competitive={false}` or omit.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-core/ dx-web/src/features/web/play-single/ dx-web/src/features/web/play-pk/ dx-web/src/features/web/play-group/
git commit -m "feat: hide skip and answer buttons in competitive modes"
```

---

## Task 17: Full Build + Lint Verification

- [ ] **Step 1: Backend build**

```bash
cd dx-api && go build ./...
```

Fix any remaining errors.

- [ ] **Step 2: Backend lint**

```bash
cd dx-api && go vet ./...
```

Fix any warnings.

- [ ] **Step 3: Frontend build**

```bash
cd dx-web && npm run build
```

Fix any TypeScript errors.

- [ ] **Step 4: Frontend lint**

```bash
cd dx-web && npm run lint
```

Fix any ESLint issues.

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "fix: resolve all build and lint issues"
```

---

## File Impact Summary

### Backend — Delete (6 files)
- `dx-api/app/models/game_session_total.go`
- `dx-api/app/models/game_session_level.go`
- `dx-api/app/models/game_stats_total.go`
- `dx-api/app/models/game_stats_level.go`
- `dx-api/app/services/api/stats_service.go`
- Migrations for dropped tables (keep as history, add new migration)

### Backend — Create (2 files)
- `dx-api/app/models/game_session.go`
- `dx-api/database/migrations/20260405000001_create_game_sessions_table.go`

### Backend — Modify (14 files)
- `dx-api/app/models/game_record.go`
- `dx-api/app/models/game_pk.go`
- `dx-api/app/services/api/game_stats_service.go`
- `dx-api/app/services/api/game_play_single_service.go`
- `dx-api/app/services/api/game_play_pk_service.go`
- `dx-api/app/services/api/game_play_group_service.go`
- `dx-api/app/services/api/leaderboard_service.go`
- `dx-api/app/services/api/hall_service.go`
- `dx-api/app/services/api/game_service.go`
- `dx-api/app/services/api/pk_winner_service.go`
- `dx-api/app/services/api/group_winner_service.go`
- `dx-api/app/services/api/group_game_service.go`
- `dx-api/app/http/controllers/api/game_play_single_controller.go`
- `dx-api/app/http/controllers/api/game_play_pk_controller.go`
- `dx-api/app/http/controllers/api/game_play_group_controller.go`
- `dx-api/app/http/requests/api/session_request.go`
- `dx-api/routes/api.go`
- `dx-api/bootstrap/migrations.go`

### Frontend — Modify (~15 files)
- `dx-web/src/features/web/play-single/actions/session.action.ts`
- `dx-web/src/features/web/play-pk/actions/session.action.ts`
- `dx-web/src/features/web/play-pk/types/pk-play.ts`
- `dx-web/src/features/web/play-pk/hooks/use-pk-play-store.ts`
- `dx-web/src/features/web/play-pk/hooks/use-pk-play-events.ts`
- PK play shell + result panel components
- `dx-web/src/features/web/play-group/actions/session.action.ts`
- `dx-web/src/features/web/play-group/types/group-play.ts`
- `dx-web/src/features/web/play-group/hooks/use-group-play-store.ts`
- `dx-web/src/features/web/play-group/hooks/use-group-play-events.ts`
- Group play shell + result panel components
- `dx-web/src/features/web/play-core/hooks/use-game-store.ts`
- `dx-web/src/features/web/hall/components/game-progress-card.tsx`
- Leaderboard types + components
- Play mode components (skip/answer button visibility)
