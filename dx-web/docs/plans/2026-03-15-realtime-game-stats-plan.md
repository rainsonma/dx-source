# Realtime Game Stats Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Update GameStatsTotal and GameStatsLevel on each level completion instead of only at session end, so users see real-time progress. Add `totalScores` field to GameStatsLevel.

**Architecture:** Move stats accumulation (`totalPlayTime`, `totalScores`, `totalExp`, `highestScore`) from `endSession()` into the existing `completeLevel()` transaction. Add a new `updateGameStatsOnLevelComplete()` model function. Slim down `endSession()` to only handle session-count fields. Add `totalScores` column to `game_stats_levels` table.

**Tech Stack:** Prisma 7, PostgreSQL, Next.js server actions, Zustand (client — no changes needed)

**Design doc:** `docs/plans/2026-03-15-realtime-game-stats-design.md`

---

### Task 1: Add `totalScores` to GameStatsLevel schema

**Files:**
- Modify: `prisma/schema/game-stats-level.prisma:5-6`

**Step 1: Add field to schema**

In `prisma/schema/game-stats-level.prisma`, add `totalScores` after `highestScore` (line 5):

```prisma
  highestScore     Int       @default(0) @map("highest_score")
  totalScores      Int       @default(0) @map("total_scores")
  expEarned        Boolean   @default(false) @map("exp_earned")
```

**Step 2: Create migration**

Run: `npx prisma migrate dev --name add-total-scores-to-game-stats-level`
Expected: Migration succeeds, `game_stats_levels` table gains `total_scores INT NOT NULL DEFAULT 0`

**Step 3: Generate Prisma client**

Run: `npm run prisma:generate`
Expected: Client regenerated with `totalScores` on `GameStatsLevel`

**Step 4: Verify build**

Run: `npm run build`
Expected: Build succeeds — no code references `totalScores` yet, so no type errors

**Step 5: Commit**

```bash
git add prisma/
git commit -m "feat: add totalScores column to game_stats_levels table"
```

---

### Task 2: Add `updateGameStatsOnLevelComplete()` model function

**Files:**
- Modify: `src/models/game-stats-total/game-stats-total.mutation.ts`

**Step 1: Add new function**

Append to `src/models/game-stats-total/game-stats-total.mutation.ts` (after `updateHighestScoreIfNeeded`):

```typescript
/** Update game-level stats when a level completes (realtime accumulation) */
export async function updateGameStatsOnLevelComplete(
  userId: string,
  gameId: string,
  data: {
    levelScore: number;
    playTimeSeconds: number;
    expEarned: number;
  },
  tx?: TxClient
) {
  const client = tx ?? db;

  const stats = await client.gameStatsTotal.findUnique({
    where: { userId_gameId: { userId, gameId } },
    select: { highestScore: true },
  });

  if (!stats) throw new Error("Game stats not found");

  return client.gameStatsTotal.update({
    where: { userId_gameId: { userId, gameId } },
    data: {
      totalPlayTime: { increment: data.playTimeSeconds },
      totalScores: { increment: data.levelScore },
      totalExp: { increment: data.expEarned },
      ...(data.levelScore > stats.highestScore && {
        highestScore: data.levelScore,
      }),
    },
    select: { id: true },
  });
}
```

**Step 2: Add `TxClient` type and `Prisma` import**

The file currently imports `{ db }` from `@/lib/db`. Update to:

```typescript
import { db, Prisma } from "@/lib/db";

type TxClient = Prisma.TransactionClient;
```

**Step 3: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add src/models/game-stats-total/game-stats-total.mutation.ts
git commit -m "feat: add updateGameStatsOnLevelComplete model function"
```

---

### Task 3: Update `completeLevelStats()` to increment `totalScores`

**Files:**
- Modify: `src/models/game-stats-level/game-stats-level.mutation.ts:44-86`

**Step 1: Update function signature and body**

In `completeLevelStats()`, add `totalScores` increment to the update data. The `data` parameter already includes `score`. Add this line inside the `data` object of the `update()` call, after `totalPlayTime`:

```typescript
      totalScores: { increment: data.score },
```

The full update `data` block becomes:

```typescript
    data: {
      completionCount: { increment: 1 },
      lastCompletedAt: now,
      totalPlayTime: { increment: data.playTimeSeconds },
      totalScores: { increment: data.score },
      ...(existing.firstCompletedAt === null && {
        firstCompletedAt: now,
      }),
      ...(data.score > existing.highestScore && {
        highestScore: data.score,
      }),
      ...(data.grantExp && { expEarned: true }),
    },
```

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add src/models/game-stats-level/game-stats-level.mutation.ts
git commit -m "feat: increment totalScores on level completion in completeLevelStats"
```

---

### Task 4: Add `totalScores` to level stats queries

**Files:**
- Modify: `src/models/game-stats-level/game-stats-level.query.ts`

**Step 1: Add `totalScores` to both query selects**

In `getLevelStats()` (line 9-16), add `totalScores: true` to the select:

```typescript
    select: {
      id: true,
      highestScore: true,
      totalScores: true,
      totalPlayTime: true,
      completionCount: true,
      firstCompletedAt: true,
      lastCompletedAt: true,
    },
```

In `getAllLevelStats()` (line 27-34), add the same:

```typescript
    select: {
      id: true,
      gameLevelId: true,
      highestScore: true,
      totalScores: true,
      totalPlayTime: true,
      completionCount: true,
      firstCompletedAt: true,
      lastCompletedAt: true,
    },
```

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add src/models/game-stats-level/game-stats-level.query.ts
git commit -m "feat: include totalScores in level stats queries"
```

---

### Task 5: Wire up GameStatsTotal updates in `completeLevel()` service

**Files:**
- Modify: `src/features/web/play/services/session.service.ts:300-379`

**Step 1: Add import**

Add `updateGameStatsOnLevelComplete` to the import from game-stats-total mutations (line 32):

```typescript
import {
  upsertGameStats,
  updateGameStatsAfterSession,
  markGameFirstCompletion,
  updateHighestScoreIfNeeded,
  updateGameStatsOnLevelComplete,
} from "@/models/game-stats-total/game-stats-total.mutation";
```

**Step 2: Get `gameId` inside the transaction**

Inside `completeLevel()`, after the `sessionLevel` query (line 313-320), add a query to get `gameId` from the session:

```typescript
    const session = await tx.gameSessionTotal.findUnique({
      where: { id: sessionId },
      select: { gameId: true },
    });

    if (!session) throw new Error("Session not found");
```

**Step 3: Add `updateGameStatsOnLevelComplete()` call**

After the existing `completeLevelStats()` call (line 357-366) and before the `incrementUserExp` call, add:

```typescript
    await updateGameStatsOnLevelComplete(
      userId,
      session.gameId,
      {
        levelScore: data.score,
        playTimeSeconds: sessionLevel.playTime,
        expEarned: expAmount,
      },
      tx
    );
```

**Step 4: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add src/features/web/play/services/session.service.ts
git commit -m "feat: update GameStatsTotal on each level completion"
```

---

### Task 6: Slim down `endSession()` service

**Files:**
- Modify: `src/features/web/play/services/session.service.ts:437-485`
- Modify: `src/models/game-stats-total/game-stats-total.mutation.ts:40-65`

**Step 1: Replace `updateGameStatsAfterSession()` with a slimmed-down version**

In `src/models/game-stats-total/game-stats-total.mutation.ts`, replace the existing `updateGameStatsAfterSession` function (lines 40-65) with:

```typescript
/** Update session-count stats after a session ends */
export async function updateGameStatsAfterSession(
  userId: string,
  gameId: string,
  data: {
    allCompleted: boolean;
  }
) {
  return db.gameStatsTotal.update({
    where: { userId_gameId: { userId, gameId } },
    data: {
      totalSessions: { increment: 1 },
      ...(data.allCompleted && {
        completionCount: { increment: 1 },
        lastCompletedAt: new Date(),
      }),
    },
    select: { id: true },
  });
}
```

**Step 2: Update `endSession()` in service**

In `src/features/web/play/services/session.service.ts`, simplify `endSession()`. Remove the `playTime` query, the `updateHighestScoreIfNeeded` call, and update the `updateGameStatsAfterSession` call:

```typescript
/** End a game session and update progress */
export async function endSession(
  sessionId: string,
  data: {
    gameId: string;
    score: number;
    exp: number;
    maxCombo: number;
    correctCount: number;
    wrongCount: number;
    skipCount: number;
    allLevelsCompleted: boolean;
  }
) {
  const userId = await requireUserId();

  await updateSessionStats(sessionId, {
    score: data.score,
    exp: data.exp,
    maxCombo: data.maxCombo,
    correctCount: data.correctCount,
    wrongCount: data.wrongCount,
    skipCount: data.skipCount,
  });

  await completeSession(sessionId);

  await updateGameStatsAfterSession(userId, data.gameId, {
    allCompleted: data.allLevelsCompleted,
  });

  if (data.allLevelsCompleted) {
    await markGameFirstCompletion(userId, data.gameId);
  }

  return { completed: true };
}
```

**Step 3: Remove unused import**

Remove `updateHighestScoreIfNeeded` from the import in `session.service.ts` if no other code uses it (it's now called from `updateGameStatsOnLevelComplete` instead).

**Step 4: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add src/models/game-stats-total/game-stats-total.mutation.ts src/features/web/play/services/session.service.ts
git commit -m "refactor: slim endSession to only update session-count fields"
```

---

### Task 7: Update rules doc

**Files:**
- Modify: `rules/GameLSRWRule.md:177-194`

**Step 1: Update Progress Tracking section**

Replace lines 177-194 with:

```markdown
## Progress Tracking

- **Per session** (`GameSessionTotal`): score, exp earned, max combo, correct/wrong/skip counts, play time
- **Per level session** (`GameSessionLevel`): score, exp, max combo, correct/wrong/skip counts, resume point (`currentContentItemId`), degree, pattern, play time
- **Per record** (`GameRecord`): linked to `gameSessionLevelId`, stores individual answer per content item per level session attempt
- **Per level lifetime** (`GameStatsLevel`): highest score, total scores, EXP earned flag, total play time, completion count — updated on each level completion
- **Per game lifetime** (`GameStatsTotal`): highest score, total scores, total sessions, total EXP, total play time, completion count — `totalPlayTime`, `totalScores`, `totalExp`, `highestScore` updated per level completion; `totalSessions`, `completionCount` updated on session end
- **Per user global**: total EXP (`User.exp`)
```

**Step 2: Commit**

```bash
git add rules/GameLSRWRule.md
git commit -m "docs: update progress tracking rules for realtime stats"
```

---

### Task 8: Final verification

**Step 1: Run full build**

Run: `npm run build`
Expected: Build succeeds with no errors

**Step 2: Run lint**

Run: `npm run lint`
Expected: No lint errors

**Step 3: Manual verification checklist**

- [ ] `completeLevel()` updates both GameStatsLevel and GameStatsTotal in one transaction
- [ ] `endSession()` no longer touches `totalPlayTime`, `totalScores`, `totalExp`, `highestScore`
- [ ] `GameStatsLevel.totalScores` incremented per level completion
- [ ] `GameStatsTotal.totalScores` incremented per level completion
- [ ] No double-counting — fields moved from session-end to level-completion, not duplicated
- [ ] `updateHighestScoreIfNeeded` import removed from service if unused
- [ ] Client code unchanged — `use-game-result.ts`, action signatures all preserved
