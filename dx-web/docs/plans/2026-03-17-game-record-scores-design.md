# Game Record Scores Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `baseScore` and `comboScore` fields to `GameRecord` so each answer record stores the points earned for that submission.

**Architecture:** Thread two new Int fields through the existing 4-layer pipeline: schema → model mutation → service → client hook. No new files or functions needed.

**Tech Stack:** Prisma schema, TypeScript, Next.js server actions, Zustand store

---

## Context

- `baseScore`: 1 if correct, 0 if wrong (base point per game rules in `rules/GameLSRWRule.md`)
- `comboScore`: combo bonus earned on this answer (3, 5, or 10 at combo thresholds), 0 otherwise
- `baseScore + comboScore` = total points earned for this record
- Scoring constants live in `src/consts/scoring.ts` (`SCORING.CORRECT_ANSWER = 1`)
- `processAnswer()` in `src/features/web/play/helpers/scoring.ts` already returns `comboBonus` separately
- The upsert uses `update: {}` — once a record exists for a level-session+item, it won't be overwritten

## Data Flow

```
use-lsrw.ts (client)
  → recordAnswerAction (server action)
    → recordAnswer (service)
      → upsertRecord (model mutation)
```

---

### Task 1: Add fields to Prisma schema

**Files:**
- Modify: `prisma/schema/game-record.prisma:7-9`

**Step 1: Add `baseScore` and `comboScore` fields**

After line 9 (`userAnswer`), add:

```prisma
  baseScore        Int     @default(0) @map("base_score")
  comboScore       Int     @default(0) @map("combo_score")
```

**Step 2: Generate Prisma client**

Run: `npm run prisma:generate`
Expected: Success, no errors

**Step 3: Create migration**

Run: `npx prisma migrate dev --name add-base-score-combo-score-to-game-record`
Expected: Migration created and applied successfully. Existing rows get default 0.

**Step 4: Commit**

```
feat: add baseScore and comboScore fields to GameRecord schema
```

---

### Task 2: Update model mutation

**Files:**
- Modify: `src/models/game-record/game-record.mutation.ts:11-18`

**Step 1: Add fields to `upsertRecord` data param**

Add `baseScore` and `comboScore` to the data type (after `sourceAnswer`):

```ts
export async function upsertRecord(
  data: {
    gameSessionTotalId: string;
    gameSessionLevelId: string;
    gameLevelId: string;
    contentItemId: string;
    isCorrect: boolean;
    userAnswer: string;
    sourceAnswer: string;
    baseScore: number;
    comboScore: number;
  },
  tx?: TxClient
)
```

No other changes needed — the `create: { id, ...data }` spread already includes all fields.

**Step 2: Verify build**

Run: `npm run build`
Expected: Build fails at service layer (service doesn't pass the new fields yet). This is expected.

**Step 3: Commit**

```
feat: add baseScore and comboScore to upsertRecord params
```

---

### Task 3: Update service layer

**Files:**
- Modify: `src/features/web/play/services/session.service.ts:186-219`

**Step 1: Add fields to `recordAnswer` params**

Add `baseScore` and `comboScore` to the data type (after `sourceAnswer` at line 193):

```ts
export async function recordAnswer(data: {
  gameSessionTotalId: string;
  gameSessionLevelId: string;
  gameLevelId: string;
  contentItemId: string;
  isCorrect: boolean;
  userAnswer: string;
  sourceAnswer: string;
  baseScore: number;
  comboScore: number;
  score: number;
  maxCombo: number;
  playTime: number;
  nextContentItemId: string | null;
})
```

**Step 2: Pass fields through to `upsertRecord`**

In the `upsertRecord` call (lines 208-218), add the two new fields:

```ts
    await upsertRecord(
      {
        gameSessionTotalId: data.gameSessionTotalId,
        gameSessionLevelId: data.gameSessionLevelId,
        gameLevelId: data.gameLevelId,
        contentItemId: data.contentItemId,
        isCorrect: data.isCorrect,
        userAnswer: data.userAnswer,
        sourceAnswer: data.sourceAnswer,
        baseScore: data.baseScore,
        comboScore: data.comboScore,
      },
      tx
    );
```

**Step 3: Verify build**

Run: `npm run build`
Expected: Build fails at action layer (action doesn't pass the new fields yet). Expected.

**Step 4: Commit**

```
feat: thread baseScore and comboScore through recordAnswer service
```

---

### Task 4: Update server action

**Files:**
- Modify: `src/features/web/play/actions/session.action.ts:89-107`

**Step 1: Add fields to `recordAnswerAction` type**

Add `baseScore` and `comboScore` to the data type (after `sourceAnswer`):

```ts
export async function recordAnswerAction(data: {
  gameSessionTotalId: string;
  gameSessionLevelId: string;
  gameLevelId: string;
  contentItemId: string;
  isCorrect: boolean;
  userAnswer: string;
  sourceAnswer: string;
  baseScore: number;
  comboScore: number;
  score: number;
  maxCombo: number;
  playTime: number;
  nextContentItemId: string | null;
})
```

No other changes — `recordAnswer(data)` already passes the full object through.

**Step 2: Verify build**

Run: `npm run build`
Expected: Build fails at client layer (client doesn't pass the new fields yet). Expected.

**Step 3: Commit**

```
feat: add baseScore and comboScore to recordAnswerAction type
```

---

### Task 5: Update client hook

**Files:**
- Modify: `src/features/web/play/hooks/use-lsrw.ts:144-169`

**Step 1: Add SCORING import**

At the top of the file, add:

```ts
import { SCORING } from "@/consts/scoring";
```

**Step 2: Compute baseScore and comboScore before the server call**

Replace the block at lines 145-169:

```ts
      if (nextIndex >= items.length) {
        const isItemCorrect = !hadWrongAttemptRef.current;
        setIsRevealed(true);

        // Capture score before update to derive per-record scores
        const prevScore = useGameStore.getState().score;
        recordResult(isItemCorrect);

        // Derive per-record scores from state delta
        const latestState = useGameStore.getState();
        const pointsEarned = latestState.score - prevScore;
        const baseScore = isItemCorrect ? SCORING.CORRECT_ANSWER : 0;
        const comboScore = pointsEarned - baseScore;

        // Record to server (fire-and-forget)
        if (sessionId && levelSessionId && levelId && currentItem) {
          const nextItemId = contentItems?.[currentIndex + 1]?.id ?? null;
          recordAnswerAction({
            gameSessionTotalId: sessionId,
            gameSessionLevelId: levelSessionId!,
            gameLevelId: levelId,
            contentItemId: currentItem.id,
            isCorrect: isItemCorrect,
            userAnswer: allTypedWords
              .filter((w) => w.isAnswer)
              .map((w) => w.text)
              .join(" "),
            sourceAnswer: currentItem.content as string,
            baseScore,
            comboScore,
            score: latestState.score,
            maxCombo: latestState.combo.maxCombo,
            playTime: getElapsedSeconds(),
            nextContentItemId: nextItemId,
          });
        }

        // Save incorrect item for review (fire-and-forget)
        if (!isItemCorrect && currentItem && gameId && levelId) {
          markAsReviewAction({
            contentItemId: currentItem.id,
            gameId,
            gameLevelId: levelId,
          });
        }

        // Wait for user to press Enter/Space to advance (handled in handleKeyDown)
      }
```

**Step 3: Verify build**

Run: `npm run build`
Expected: Build succeeds. All layers now pass `baseScore` and `comboScore`.

**Step 4: Manual smoke test**

Run: `npm run dev`
Play a level, answer correctly and incorrectly. Check Prisma Studio (`npm run prisma:studio`) to verify:
- Correct answers: `baseScore = 1`, `comboScore = 0` (unless at combo threshold)
- At 3-combo: `baseScore = 1`, `comboScore = 3`
- Wrong answers: `baseScore = 0`, `comboScore = 0`

**Step 5: Commit**

```
feat: compute and send baseScore/comboScore from client to GameRecord
```

---

## Unchanged

- `processAnswer` in `src/features/web/play/helpers/scoring.ts`
- `useGameStore.recordResult` in `src/features/web/play/hooks/use-game-store.ts`
- Session/level stat updates (work with cumulative `score`, not per-record)
- Skip flow (no GameRecord created on skip, per game rules)
