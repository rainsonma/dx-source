# Session Progress Tracking Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `totalLevelsCount`/`playedLevelsCount` to GameSessionTotal and `totalItemsCount`/`playedItemsCount` to GameSessionLevel, so session progress is explicitly tracked in the database.

**Architecture:** Four new Int columns with `@default(0)`. Totals are set as snapshots at creation time; played counts are incremented atomically in existing transactions. `totalItemsCount` is filtered by degree content types via `DEGREE_CONTENT_TYPES`.

**Tech Stack:** Prisma 7, PostgreSQL, Next.js server actions

---

### Task 1: Add Prisma Schema Fields

**Files:**
- Modify: `prisma/schema/game-session-total.prisma:12-18` (add after `playTime`)
- Modify: `prisma/schema/game-session-level.prisma:12-17` (add after `playTime`)

**Step 1: Add fields to GameSessionTotal**

In `prisma/schema/game-session-total.prisma`, add after the `playTime` line (line 18):

```prisma
  totalLevelsCount   Int       @default(0) @map("total_levels_count")
  playedLevelsCount  Int       @default(0) @map("played_levels_count")
```

**Step 2: Add fields to GameSessionLevel**

In `prisma/schema/game-session-level.prisma`, add after the `playTime` line (line 17):

```prisma
  totalItemsCount    Int       @default(0) @map("total_items_count")
  playedItemsCount   Int       @default(0) @map("played_items_count")
```

**Step 3: Create migration**

Run: `npx prisma migrate dev --name add_session_progress_tracking`

**Step 4: Generate Prisma client**

Run: `npm run prisma:generate`

**Step 5: Commit**

```
feat: add session progress tracking schema fields
```

---

### Task 2: Add Count Queries to Model Layer

**Files:**
- Modify: `src/models/game-level/game-level.query.ts`
- Modify: `src/models/content-item/content-item.query.ts`

**Step 1: Add `countActiveLevelsByGameId` to game-level.query.ts**

Add at the end of `src/models/game-level/game-level.query.ts`:

```typescript
/** Count active levels for a game */
export async function countActiveLevelsByGameId(gameId: string) {
  return db.gameLevel.count({
    where: { gameId, isActive: true },
  });
}
```

**Step 2: Add `countActiveContentItems` to content-item.query.ts**

Add at the end of `src/models/content-item/content-item.query.ts`:

```typescript
/** Count active content items for a level, optionally filtered by content types */
export async function countActiveContentItems(
  gameLevelId: string,
  contentTypes?: string[]
) {
  return db.contentItem.count({
    where: {
      gameLevelId,
      isActive: true,
      ...(contentTypes && { contentType: { in: contentTypes } }),
    },
  });
}
```

**Step 3: Commit**

```
feat: add count queries for levels and content items
```

---

### Task 3: Set `totalLevelsCount` on Session Creation

**Files:**
- Modify: `src/models/game-session-total/game-session-total.mutation.ts` — `createSession()`
- Modify: `src/features/web/play/services/session.service.ts` — `startGameSession()`

**Step 1: Add `totalLevelsCount` param to `createSession` mutation**

In `src/models/game-session-total/game-session-total.mutation.ts`, update the `createSession` function signature and data:

```typescript
/** Create a new game session */
export async function createSession(
  userId: string,
  gameId: string,
  currentLevelId: string,
  degree: string = "intermediate",
  pattern?: string,
  totalLevelsCount?: number
) {
  const id = ulid();

  return db.$transaction(async (tx) => {
    await assertFK(tx, [
      { table: "users", id: userId },
      { table: "games", id: gameId },
      { table: "game_levels", id: currentLevelId },
    ]);

    return tx.gameSessionTotal.create({
      data: {
        id,
        userId,
        gameId,
        currentLevelId,
        degree,
        pattern: pattern ?? null,
        totalLevelsCount: totalLevelsCount ?? 0,
      },
      select: { id: true, currentLevelId: true, startedAt: true },
    });
  });
}
```

**Step 2: Query level count in `startGameSession` and pass it**

In `src/features/web/play/services/session.service.ts`, add import:

```typescript
import { countActiveLevelsByGameId } from "@/models/game-level/game-level.query";
```

Then in `startGameSession()`, before the `createSession` call (around line 103-105), add the count query and pass it:

```typescript
  const totalLevelsCount = await countActiveLevelsByGameId(gameId);

  await upsertGameStats(userId, gameId);

  const session = await createSession(userId, gameId, resolvedLevelId, degree, pattern, totalLevelsCount);
```

**Step 3: Commit**

```
feat: set totalLevelsCount on game session creation
```

---

### Task 4: Set `totalItemsCount` on Level Session Creation

**Files:**
- Modify: `src/models/game-session-level/game-session-level.mutation.ts` — `createSessionLevel()`
- Modify: `src/features/web/play/services/session.service.ts` — `startSessionLevel()`

**Step 1: Add `totalItemsCount` param to `createSessionLevel` mutation**

In `src/models/game-session-level/game-session-level.mutation.ts`, update the `createSessionLevel` function's data type and the create call:

```typescript
/** Create a session level entry, or resume existing incomplete one */
export async function createSessionLevel(
  data: {
    gameSessionTotalId: string;
    gameLevelId: string;
    degree: string;
    pattern?: string;
    totalItemsCount?: number;
  },
  tx?: TxClient
) {
```

And in the `client.gameSessionLevel.create` data block, add `totalItemsCount`:

```typescript
    return client.gameSessionLevel.create({
      data: {
        id,
        gameSessionTotalId: data.gameSessionTotalId,
        gameLevelId: data.gameLevelId,
        degree: data.degree,
        pattern: data.pattern ?? null,
        totalItemsCount: data.totalItemsCount ?? 0,
      },
      select: { id: true, gameSessionTotalId: true, gameLevelId: true, currentContentItemId: true },
    });
```

**Step 2: Query content item count in `startSessionLevel` and pass it**

In `src/features/web/play/services/session.service.ts`, add imports:

```typescript
import { countActiveContentItems } from "@/models/content-item/content-item.query";
import { DEGREE_CONTENT_TYPES, type GameDegree } from "@/consts/game-degree";
```

Then update `startSessionLevel()`:

```typescript
/** Create a session-level entry and upsert level stats when the player starts a level */
export async function startSessionLevel(
  sessionId: string,
  gameLevelId: string,
  degree: string,
  pattern?: string
) {
  const userId = await requireUserId();

  await upsertLevelStats(userId, gameLevelId);

  const contentTypes = DEGREE_CONTENT_TYPES[degree as GameDegree] ?? undefined;
  const totalItemsCount = await countActiveContentItems(gameLevelId, contentTypes);

  return createSessionLevel({
    gameSessionTotalId: sessionId,
    gameLevelId,
    degree,
    pattern,
    totalItemsCount,
  });
}
```

**Step 3: Commit**

```
feat: set totalItemsCount on level session creation
```

---

### Task 5: Increment `playedLevelsCount` on Level Completion

**Files:**
- Modify: `src/features/web/play/services/session.service.ts` — `completeLevel()`

**Step 1: Add increment in `completeLevel` transaction**

In `src/features/web/play/services/session.service.ts`, inside the `completeLevel` function's `db.$transaction` block, update the existing `tx.gameSessionTotal.update` call (around line 354-360) to also increment `playedLevelsCount`:

```typescript
    await tx.gameSessionTotal.update({
      where: { id: sessionId },
      data: {
        score: data.score,
        maxCombo: data.maxCombo,
        playedLevelsCount: { increment: 1 },
        ...(meetsThreshold && { exp: { increment: expAmount } }),
      },
    });
```

**Step 2: Commit**

```
feat: increment playedLevelsCount on level completion
```

---

### Task 6: Increment `playedItemsCount` on Answer Recording

**Files:**
- Modify: `src/models/game-session-level/game-session-level.mutation.ts` — `updateSessionLevelStats()`

**Step 1: Add `playedItemsCount` increment in `updateSessionLevelStats`**

In `src/models/game-session-level/game-session-level.mutation.ts`, inside the `updateSessionLevelStats` function's update call (around line 124-141), add `playedItemsCount: { increment: 1 }`:

```typescript
    return client.gameSessionLevel.update({
      where: { id: sessionLevel.id },
      data: {
        correctCount: data.isCorrect ? { increment: 1 } : undefined,
        wrongCount: data.isCorrect ? undefined : { increment: 1 },
        playedItemsCount: { increment: 1 },
        score: data.score,
        maxCombo: data.maxCombo,
        playTime: data.playTime,
        currentContentItemId: data.currentContentItemId,
      },
      select: {
        id: true,
        correctCount: true,
        wrongCount: true,
        score: true,
        maxCombo: true,
      },
    });
```

This runs inside the existing `recordAnswer` transaction in `session.service.ts` — no changes needed there.

**Step 2: Commit**

```
feat: increment playedItemsCount on answer recording
```

---

### Task 7: Update GameLSRWRule.md

**Files:**
- Modify: `rules/GameLSRWRule.md`

**Step 1: Update the Progress Tracking section**

In `rules/GameLSRWRule.md`, update the "Progress Tracking" section (lines 176-183) to:

```markdown
## Progress Tracking

- **Per session** (`GameSessionTotal`): score, exp earned, max combo, correct/wrong/skip counts, play time, total levels count (snapshot at creation), played levels count (incremented on each level completion)
- **Per level session** (`GameSessionLevel`): score, exp, max combo, correct/wrong/skip counts, resume point (`currentContentItemId`), degree, pattern, play time, total items count (snapshot at creation, filtered by degree), played items count (incremented on each answer, not on skip)
- **Per record** (`GameRecord`): linked to `gameSessionLevelId`, stores individual answer per content item per level session attempt
- **Per level lifetime** (`GameStatsLevel`): highest score, total scores, total play time, completion count — updated on each level completion
- **Per game lifetime** (`GameStatsTotal`): highest score, total scores, total sessions, total EXP, total play time, completion count — `totalPlayTime`, `totalScores`, `totalExp`, `highestScore` updated per level completion; `totalSessions` updated on session creation; `completionCount` updated on session end
- **Per user global**: total EXP (`User.exp`)
```

**Step 2: Commit**

```
docs: update GameLSRWRule with session progress tracking fields
```

---

### Task 8: Verify Build

**Step 1: Run build**

Run: `npm run build`

Expected: Build succeeds with no errors.

**Step 2: Run lint**

Run: `npm run lint`

Expected: No lint errors.
