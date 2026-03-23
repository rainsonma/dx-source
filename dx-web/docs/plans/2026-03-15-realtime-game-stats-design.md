# Realtime Game Stats Updates

## Problem

GameStatsTotal and GameStatsLevel are only updated at session boundaries:
- GameStatsLevel updates on level completion (OK)
- GameStatsTotal updates only on session end (in `endSession()`)

For games with many levels, a user who plays several levels without finishing the entire game sees stale GameStatsTotal values (`totalPlayTime=0`, `totalScores=0`, `totalExp=0`). This feels broken since their level-specific stats and `User.exp` already reflect progress in real time.

Additionally, `endSession()` currently only increments `totalScores` by the last level's score — intermediate levels' scores are silently lost.

## Solution

Move GameStatsTotal updates from session-end to level-completion, alongside the existing GameStatsLevel updates. Add a `totalScores` field to GameStatsLevel for symmetry.

## Schema Change

### GameStatsLevel — add `totalScores` column

```prisma
model GameStatsLevel {
  // ... existing fields ...
  totalScores      Int       @default(0) @map("total_scores")  // NEW
}
```

Migration: `ALTER TABLE game_stats_levels ADD COLUMN total_scores INT NOT NULL DEFAULT 0;`

## Update Timing

### On level completion (`completeLevel()`)

**GameStatsLevel** (existing `completeLevelStats()` — add `totalScores`):

| Field | Update |
|-------|--------|
| `completionCount` | += 1 |
| `totalPlayTime` | += level's playTime |
| `totalScores` | += level's score *(new)* |
| `highestScore` | update if beaten |
| `expEarned` | set true (once) |
| `firstCompletedAt` | set once |
| `lastCompletedAt` | set |

**GameStatsTotal** (new — add calls inside `completeLevel()` transaction):

| Field | Update |
|-------|--------|
| `totalPlayTime` | += level's playTime |
| `totalExp` | += expEarned |
| `totalScores` | += level's score |
| `highestScore` | update if beaten (compare session running total) |

### On session end (`endSession()`)

**GameStatsTotal** — only session-level aggregates remain:

| Field | Update |
|-------|--------|
| `totalSessions` | += 1 |
| `completionCount` | += 1 (if all levels completed) |
| `lastCompletedAt` | set |
| `firstCompletedAt` | set once |

Fields **removed** from `endSession()`:
- ~~`totalPlayTime`~~ — now incremented per level completion
- ~~`totalScores`~~ — now incremented per level completion
- ~~`totalExp`~~ — now incremented per level completion
- ~~`highestScore`~~ — now checked per level completion

## Files to Change

### Schema & Migration
- `prisma/schema/game-stats-level.prisma` — add `totalScores` field

### Model Layer
- `src/models/game-stats-level/game-stats-level.mutation.ts`
  - `completeLevelStats()` — add `totalScores` increment
- `src/models/game-stats-total/game-stats-total.mutation.ts`
  - `updateGameStatsAfterSession()` — remove `totalPlayTime`, `totalScores`, `totalExp` increments
  - `updateHighestScoreIfNeeded()` — no change (reused from `completeLevel()`)
  - Add new `updateGameStatsOnLevelComplete()` function for per-level updates

### Service Layer
- `src/features/web/play/services/session.service.ts`
  - `completeLevel()` — add GameStatsTotal updates inside the transaction
  - `endSession()` — remove redundant stats updates, keep only session-count fields

### Query Layer
- `src/models/game-stats-level/game-stats-level.query.ts`
  - `getLevelStats()` / `getAllLevelStats()` — add `totalScores` to select

### Client (no changes)
- `use-game-result.ts` — no changes needed; `completeLevelAction()` and `endSessionAction()` signatures unchanged

## Score Semantics

The game store's `score` is **level-specific** (resets to 0 per level). So:
- `data.score` passed to `completeLevel()` = the current level's score
- `GameStatsLevel.totalScores` += level's score (straightforward)
- `GameStatsTotal.totalScores` += level's score (accumulates across levels)
- `GameStatsTotal.highestScore` compares against per-level score (best single level score, not session total)

## Rules Doc Update

Update `rules/GameLSRWRule.md` Progress Tracking section to reflect the new timing.
