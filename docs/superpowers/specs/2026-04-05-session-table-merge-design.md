# Session Table Merge & Level-Based Game Play

**Date:** 2026-04-05
**Status:** Draft

## Overview

Restructure the game session and stats architecture to target individual game levels instead of entire games. Three interconnected changes:

1. Remove `game_stats_totals` and `game_stats_levels` — derive all stats from sessions
2. Merge `game_session_totals` and `game_session_levels` into a single `game_sessions` table
3. PK and group play target a single game level per session, with first-to-complete-wins strategy

## 1. New `game_sessions` Table

Replaces both `game_session_totals` and `game_session_levels`. One row = one level play.

### Schema

```
game_sessions
├── id                          PK, ULID
├── user_id                     FK users
├── game_id                     FK games (denormalized for direct queries)
├── game_level_id               FK game_levels — THE level being played
├── degree                      string (beginner/intermediate/advanced)
├── pattern                     string nullable (listen/speak/read/write)
├── current_content_item_id     FK content_items nullable — resume point
├── started_at                  timestamptz
├── last_played_at              timestamptz
├── ended_at                    timestamptz nullable — NULL = active
├── score                       int default 0
├── exp                         int default 0
├── max_combo                   int default 0
├── correct_count               int default 0
├── wrong_count                 int default 0
├── skip_count                  int default 0
├── play_time                   int default 0 — milliseconds
├── total_items_count           int default 0 — content items at session start
├── played_items_count          int default 0 — items answered
├── game_group_id               FK game_groups nullable
├── game_subgroup_id            FK game_subgroups nullable
├── game_pk_id                  FK game_pks nullable
├── created_at                  timestamptz
├── updated_at                  timestamptz
```

### Indexes

```sql
-- Active single-play session uniqueness
CREATE UNIQUE INDEX idx_game_sessions_active_single
  ON game_sessions (user_id, game_level_id, degree, pattern)
  WHERE ended_at IS NULL AND game_group_id IS NULL AND game_pk_id IS NULL;

-- Active group session uniqueness
CREATE UNIQUE INDEX idx_game_sessions_active_group
  ON game_sessions (user_id, game_level_id, degree, pattern, game_group_id)
  WHERE ended_at IS NULL AND game_group_id IS NOT NULL;

-- Active PK session uniqueness
CREATE UNIQUE INDEX idx_game_sessions_active_pk
  ON game_sessions (user_id, game_level_id, degree, pattern, game_pk_id)
  WHERE ended_at IS NULL AND game_pk_id IS NOT NULL;

-- Group session lookup
CREATE INDEX idx_game_sessions_group
  ON game_sessions (game_group_id) WHERE game_group_id IS NOT NULL;

-- PK session lookup
CREATE INDEX idx_game_sessions_pk
  ON game_sessions (game_pk_id) WHERE game_pk_id IS NOT NULL;

-- Leaderboard queries (windowed)
CREATE INDEX idx_game_sessions_leaderboard
  ON game_sessions (user_id, last_played_at);

-- Stats queries by game
CREATE INDEX idx_game_sessions_user_game
  ON game_sessions (user_id, game_id);
```

### Dropped Fields (from game_session_totals)

- `total_levels_count` — derivable: `COUNT(*) FROM game_levels WHERE game_id = ? AND is_active`
- `played_levels_count` — derivable: `COUNT(DISTINCT game_level_id) FROM game_sessions WHERE user_id = ? AND game_id = ? AND ended_at IS NOT NULL`
- `current_level_id` — no longer needed, each session IS a level
- `game_session_total_id` — no parent-child relationship

## 2. Stats Tables Removal

**Delete tables:** `game_stats_totals`, `game_stats_levels`
**Delete file:** `dx-api/app/services/api/stats_service.go`
**Delete models:** `GameStatsTotal`, `GameStatsLevel`

### Stats Derivation Queries

**Game detail stats card** (`GetGameStats`):
```sql
SELECT
  COUNT(*) AS total_sessions,
  COALESCE(MAX(score), 0) AS highest_score,
  COALESCE(SUM(score), 0) AS total_scores,
  COALESCE(SUM(exp), 0) AS total_exp,
  COALESCE(SUM(play_time), 0) AS total_play_time,
  COUNT(*) FILTER (WHERE ended_at IS NOT NULL) AS completion_count,
  MIN(ended_at) AS first_completed_at
FROM game_sessions
WHERE user_id = ? AND game_id = ? AND ended_at IS NOT NULL
```

**Recently played games** (`GetPlayedGames`):
```sql
SELECT DISTINCT ON (game_id) game_id, last_played_at
FROM game_sessions
WHERE user_id = ?
ORDER BY game_id, last_played_at DESC
```
Then sort by `last_played_at DESC`, limit 10.

**Leaderboard — all-time exp:**
```sql
WITH ranked AS (
  SELECT u.id, u.username, u.nickname, u.avatar_id,
         COALESCE(SUM(s.exp), 0)::int AS value,
         RANK() OVER (ORDER BY COALESCE(SUM(s.exp), 0) DESC)::int AS rank
  FROM users u
  INNER JOIN game_sessions s ON s.user_id = u.id
  WHERE u.is_active = true
  GROUP BY u.id, u.username, u.nickname, u.avatar_id
  HAVING COALESCE(SUM(s.exp), 0) > 0
)
SELECT * FROM ranked WHERE rank <= 100 ORDER BY rank
```

**Leaderboard — windowed (day/week/month):**
Same pattern but add `AND s.last_played_at >= $1 AND s.last_played_at < NOW()`.

**Leaderboard — all-time playtime:**
Same as exp but `SUM(s.play_time)` instead of `SUM(s.exp)`.

**User rank:** If user's ID not in top 100 results, frontend shows "未上榜".

### Deleted Write Functions

All functions in `stats_service.go` are deleted:
- `UpsertGameStats()`
- `UpsertLevelStats()`
- `completeLevelStatsInTx()`
- `updateGameStatsOnLevelCompleteInTx()`
- `UpdateGameStatsAfterSession()`
- `MarkGameFirstCompletion()`

These are no longer called from any play service.

## 3. `game_records` Table Change

Replace two foreign keys with one:

```
Before:
  game_session_total_id    FK game_session_totals
  game_session_level_id    FK game_session_levels

After:
  game_session_id          FK game_sessions
```

Keep `game_level_id` and `content_item_id` as-is.

## 4. `game_pks` Model Change

```
Before:
  current_level_id    nullable — advances through levels

After:
  game_level_id       required — the specific level being played
  (current_level_id removed)
```

All other fields stay: `id`, `user_id`, `opponent_id`, `game_id`, `degree`, `pattern`, `robot_difficulty`, `is_playing`, `last_winner_id`.

## 5. Single-Play: Level-Based

### Flow

1. User selects a level (from game detail hero section or levels grid)
2. `StartSession(userID, gameID, gameLevelID, degree, pattern)` — creates one `game_session` row
3. User answers questions — `RecordAnswer()` updates `game_session` row
4. Skip is available in single-play — `RecordSkip()` updates `game_session` row
5. `CompleteLevel()` — sets `ended_at`, calculates exp
6. If next level exists in game: show "Next Level" button → starts new session

### Removed Functions

- `AdvanceLevel()` — no multi-level within session
- `CheckActiveLevelSession()` — no level-within-session concept
- `StartLevel()` — merged into `StartSession()`

### Changed Functions

- `StartSession()` — now takes `gameLevelID` as required parameter
- `CheckActiveSession()` — checks for active session by `(user_id, game_level_id, degree, pattern)`
- `EndSession()` — simplified, no "all levels completed" logic
- `CompleteLevel()` — no longer updates `played_levels_count` on parent, just sets `ended_at`
- `RestoreSessionData()` — returns data from single `game_sessions` row

## 6. PK Play: Level-Based, First-to-Complete Wins

### Flow

1. User starts PK for a specific level: `StartPk(userID, gameID, gameLevelID, degree, pattern, difficulty)`
2. Creates `game_pks` row + `game_session` for human + spawns robot goroutine for the level
3. **Skip and answer buttons disabled** — pure competition, no shortcuts
4. Both players race to complete all content items
5. **First to complete wins** — the moment a player finishes:
   - Their `game_session.ended_at` is set
   - SSE broadcasts result to opponent immediately
   - Opponent's game ends — their `game_session.ended_at` is set with partial stats
   - Robot goroutine cancelled if human wins first
6. Result panel shown to all players immediately
7. If next level exists: "Next Level" button → starts new PK for next level (same opponent/difficulty)

### New Function

- `NextPkLevel(userID, pkID)` — finds next active level by order, creates new `game_pks` + `game_session` for the next level with same opponent/difficulty settings

### Removed Functions/Routes

- `StartPkLevel()` — merged into `StartPk()`
- Score comparison logic — winner is first to complete, not highest score

### SSE Events

- `pk_player_complete` — someone finished first (includes winner info)
- `pk_player_action` — combo streak updates during play
- `pk_timeout` / `pk_timeout_warning` — keep as-is
- `pk_force_end` — keep as-is
- **Remove:** `pk_next_level`, `pk_level_complete`

### Robot Simplification

- Robot spawns once for the single level
- Robot state management simplified: one goroutine per PK, no multi-level lifecycle
- Robot completion triggers same first-to-complete logic

## 7. Group Play: Level-Based, First-to-Complete Wins

### Flow

1. Owner sets game + specific level + mode + time limit: `SetGroupGame()`
2. Members start session: `GroupPlayStartSession()` — creates one `game_session` per member
3. **Skip and answer buttons disabled** — pure competition
4. All members race to complete
5. **First to complete wins** — the moment a member finishes:
   - Their `game_session.ended_at` is set
   - SSE broadcasts result to all group members immediately
   - All other members' sessions force-ended with partial stats
6. Result panel shown to all players immediately
7. If next level exists: **any member** can trigger "Next Level" → new endpoint `NextGroupLevel(memberID, groupID)` auto-advances to next level by order and starts new round (no need for owner to call `SetGroupGame` again)

### Winner Determination

- **Solo mode:** First individual to complete wins
- **Team mode:** First team where any member completes wins (that member's completion triggers team victory)

### New Function

- `NextGroupLevel(memberID, groupID)` — any member can call; finds next active level by order, updates `start_game_level_id`, broadcasts SSE to start new round

### Removed Functions

- Score comparison logic — winner is first to complete

### SSE Events

- `group_player_complete` — someone finished first (includes winner info)
- `group_player_action` — combo streak updates
- `group_game_force_end` — keep
- `group_dismissed` — keep
- **Remove:** `group_next_level`

## 8. Frontend Changes

### Game Detail Page

- Hero section: "Start" button starts level 1 (or resumes active session)
- Levels grid: each level shows start/resume button
- Stats card: derived from `game_sessions` queries

### Play Core

- Skip and answer buttons: **hidden in PK and group modes**, visible in single-play
- Result panel: "Next Level" button shown across all modes if next level exists
- No more "advance to next level" within a session

### PK Play

- Start with specific level selection
- No "Next Level" mid-session flow
- Opponent completion → immediate result panel (no waiting screen)
- "Next Level" on result panel starts new PK
- Either player can trigger "Next Level"

### Group Play

- Members see specific level being played
- No "Next Level" mid-session flow
- First completion → immediate result panel for everyone
- "Next Level" on result panel: any member can trigger
- Owner can still force-end

### Hall Dashboard

- Session progress: derived from `game_sessions` data
  - Progress = `COUNT(DISTINCT game_level_id) WHERE ended_at IS NOT NULL` / total levels
- Heatmap: unchanged (queries `game_records` directly)
- Leaderboard: user not in top 100 shows "未上榜"

## 9. Tables Dropped

- `game_stats_totals`
- `game_stats_levels`
- `game_session_totals`
- `game_session_levels`

## 10. Files Affected

### Backend (dx-api)

**Models — delete:**
- `game_stats_total.go`
- `game_stats_level.go`
- `game_session_total.go`
- `game_session_level.go`

**Models — create:**
- `game_session.go`

**Models — modify:**
- `game_record.go` — `game_session_id` replaces two FK fields
- `game_pk.go` — `game_level_id` replaces `current_level_id`

**Services — delete:**
- `stats_service.go`

**Services — modify:**
- `game_play_single_service.go` — level-based sessions
- `game_play_pk_service.go` — level-based, first-to-complete
- `game_play_group_service.go` — level-based, first-to-complete
- `game_stats_service.go` — derive from game_sessions
- `game_service.go` — GetPlayedGames from game_sessions
- `leaderboard_service.go` — query game_sessions, no user rank outside top 100
- `pk_winner_service.go` — first-to-complete logic
- `group_winner_service.go` — first-to-complete logic
- `hall_service.go` — session progress from game_sessions

**Controllers — modify:**
- `game_play_single_controller.go` — remove advance/start-level routes
- `game_play_pk_controller.go` — remove next-level route, add level param
- `game_play_group_controller.go` — remove next-level handling

**Routes:**
- `api.go` — remove deleted routes, update params

**Migrations — create:**
- New migration: create `game_sessions`, drop old tables, alter `game_records`, alter `game_pks`

### Frontend (dx-web)

**Actions — modify:**
- `play-single/actions/session.action.ts` — level-based API
- `play-pk/actions/session.action.ts` — level-based API, no next-level
- `play-group/actions/session.action.ts` — level-based API, no next-level

**Components — modify:**
- `play-core/` — hide skip/answer buttons for PK/group modes
- `play-single/` — result panel with "Next Level"
- `play-pk/components/pk-play-result-panel.tsx` — first-to-complete result, "Next Level"
- `play-group/components/group-play-result-panel.tsx` — first-to-complete result, "Next Level"
- Game detail page — start/resume specific levels
- Hall dashboard — progress derivation

**Types — modify:**
- Session types across all play features
- Remove `gameSessionTotalId` / `gameSessionLevelId` distinction

**Stores — modify:**
- PK store — remove multi-level state
- Group store — remove multi-level state
- Game store — handle skip/answer button visibility by mode

## 11. Constraints

- No database-level FK constraints (code-level only, for PostgreSQL partitioning)
- No data migration — fresh database after implementation
- Must not break any existing functions not covered by these changes
- No lint issues
