# Game Record: userId & duration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `userId` and `duration` fields to `GameRecord` so each answer records who answered and how long they spent.

**Architecture:** Denormalize `userId` onto `GameRecord` (sourced from `requireUserId()` already called in the service). Track per-item duration client-side via a `useRef` timestamp reset on each content item, computed on submission. Pass both new fields through the existing action → service → mutation chain.

**Tech Stack:** Prisma 7, Next.js 16 server actions, Zustand, React hooks

---

### Task 1: Schema migration

**Files:**
- Modify: `prisma/schema/game-record.prisma`

**Step 1: Add userId and duration fields to the Prisma schema**

```prisma
model GameRecord {
  id                 String  @id @db.Char(26)
  userId             String  @map("user_id") @db.Char(26)
  gameSessionTotalId String  @map("game_session_total_id") @db.Char(26)
  gameSessionLevelId String  @map("game_session_level_id") @db.Char(26)
  gameLevelId        String  @map("game_level_id") @db.Char(26)
  contentItemId      String  @map("content_item_id") @db.Char(26)
  isCorrect          Boolean @map("is_correct")
  sourceAnswer       String  @map("source_answer")
  userAnswer         String  @map("user_answer")
  baseScore          Int     @default(0) @map("base_score")
  comboScore         Int     @default(0) @map("combo_score")
  duration           Int     @default(0)

  createdAt DateTime @default(now()) @map("created_at") @db.Timestamptz
  updatedAt DateTime @updatedAt @map("updated_at") @db.Timestamptz

  sessionTotal GameSessionTotal @relation(fields: [gameSessionTotalId], references: [id], onDelete: Restrict, onUpdate: Restrict)
  level        GameLevel        @relation(fields: [gameLevelId], references: [id], onDelete: Restrict, onUpdate: Restrict)
  contentItem  ContentItem      @relation(fields: [contentItemId], references: [id], onDelete: Restrict, onUpdate: Restrict)

  @@unique([gameSessionLevelId, contentItemId])
  @@index([userId])
  @@index([gameSessionTotalId])
  @@index([gameSessionLevelId])
  @@index([gameLevelId])
  @@index([contentItemId])
  @@index([isCorrect])
  @@map("game_records")
}
```

**Step 2: Run migration**

Run: `npm run prisma:migrate -- --name add-game-record-user-duration`
Expected: Migration created and applied successfully.

**Step 3: Generate Prisma client**

Run: `npm run prisma:generate`
Expected: Prisma client regenerated with new fields.

**Step 4: Commit**

```
docs: add game record userId and duration plan
feat: add userId and duration fields to GameRecord schema
```

---

### Task 2: Update mutation layer

**Files:**
- Modify: `src/models/game-record/game-record.mutation.ts`

**Step 1: Add userId and duration to upsertRecord data param and assertFK**

```typescript
import "server-only";

import { ulid } from "ulid";
import { db, Prisma } from "@/lib/db";
import { assertFK } from "@/lib/assert-fk";

type TxClient = Prisma.TransactionClient;

/** Upsert a game answer record — skip if already exists for this level-session+item */
export async function upsertRecord(
  data: {
    userId: string;
    gameSessionTotalId: string;
    gameSessionLevelId: string;
    gameLevelId: string;
    contentItemId: string;
    isCorrect: boolean;
    userAnswer: string;
    sourceAnswer: string;
    baseScore: number;
    comboScore: number;
    duration: number;
  },
  tx?: TxClient
) {
  const doWork = async (client: TxClient) => {
    await assertFK(client, [
      { table: "users", id: data.userId },
      { table: "game_session_totals", id: data.gameSessionTotalId },
      { table: "game_session_levels", id: data.gameSessionLevelId },
      { table: "game_levels", id: data.gameLevelId },
      { table: "content_items", id: data.contentItemId },
    ]);

    const id = ulid();

    return client.gameRecord.upsert({
      where: {
        gameSessionLevelId_contentItemId: {
          gameSessionLevelId: data.gameSessionLevelId,
          contentItemId: data.contentItemId,
        },
      },
      create: { id, ...data },
      update: {},
      select: { id: true, isCorrect: true },
    });
  };

  if (tx) return doWork(tx);
  return db.$transaction(doWork);
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: Build errors in `session.service.ts` (missing userId/duration) — expected, fixed in next task.

**Step 3: Commit**

```
feat: add userId and duration to upsertRecord
```

---

### Task 3: Update service layer

**Files:**
- Modify: `src/features/web/play/services/session.service.ts`

**Step 1: Add duration to recordAnswer params and pass userId + duration to upsertRecord**

In `recordAnswer()`, change the data type to include `duration: number`. Then pass `userId` (already resolved) and `duration` into the `upsertRecord` call:

```typescript
/** Record a single answer and update session + session-level stats atomically */
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
  duration: number;
}) {
  const userId = await requireUserId();
  const { allowed } = await checkRateLimit(
    `ratelimit:record-answer:${userId}`,
    RATE_LIMIT.limit,
    RATE_LIMIT.windowSeconds
  );
  if (!allowed) throw new Error(RATE_ERROR);

  return db.$transaction(async (tx) => {
    await upsertRecord(
      {
        userId,
        gameSessionTotalId: data.gameSessionTotalId,
        gameSessionLevelId: data.gameSessionLevelId,
        gameLevelId: data.gameLevelId,
        contentItemId: data.contentItemId,
        isCorrect: data.isCorrect,
        userAnswer: data.userAnswer,
        sourceAnswer: data.sourceAnswer,
        baseScore: data.baseScore,
        comboScore: data.comboScore,
        duration: data.duration,
      },
      tx
    );

    // ... rest unchanged
  });
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: Build errors in `session.action.ts` (missing duration) — expected, fixed in next task.

**Step 3: Commit**

```
feat: pass userId and duration through session service
```

---

### Task 4: Update action layer

**Files:**
- Modify: `src/features/web/play/actions/session.action.ts`

**Step 1: Add duration to recordAnswerAction params**

```typescript
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
  duration: number;
}) {
  try {
    return { data: await recordAnswer(data), error: null };
  } catch {
    return { data: null, error: "记录失败" };
  }
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: Build errors in `use-lsrw.ts` (missing duration) — expected, fixed in next task.

**Step 3: Commit**

```
feat: add duration to recordAnswerAction
```

---

### Task 5: Track item duration on client

**Files:**
- Modify: `src/features/web/play/hooks/use-lsrw.ts`

**Step 1: Add itemStartTimeRef and reset on currentIndex change**

Add ref at the top of `useLsrw()`:

```typescript
const itemStartTimeRef = useRef<number>(Date.now());
```

In the existing `useEffect` that resets on `currentIndex` change (around line 72), add the reset:

```typescript
itemStartTimeRef.current = Date.now();
```

**Step 2: Compute duration on item completion and pass to recordAnswerAction**

In `submitWord()`, where the item is complete (`nextIndex >= items.length`), compute duration before calling `recordAnswerAction`:

```typescript
const duration = Math.round((Date.now() - itemStartTimeRef.current) / 1000);
```

Add `duration` to the `recordAnswerAction` call:

```typescript
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
  duration,
});
```

**Step 3: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 4: Verify dev server**

Run: `npm run dev`
Expected: Dev server starts without errors.

**Step 5: Commit**

```
feat: track per-item answer duration in LSRW game
```
