# Play Streak Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Track users' consecutive playing days with daily streak counting via cron job.

**Architecture:** Three new fields on User model (`currentPlayStreak`, `maxPlayStreak`, `lastPlayedAt`). `lastPlayedAt` updated inside `recordAnswer` transaction on first answer each day. Standalone cron script runs 3 bulk SQL updates at 2 AM. Scripts share a dedicated Prisma client separate from the Next.js `server-only` one.

**Tech Stack:** Prisma 7, PostgreSQL, tsx for script execution, raw SQL for bulk updates

---

### Task 1: Add streak fields to User schema

**Files:**
- Modify: `prisma/schema/user.prisma:14-17`

**Step 1: Add three fields after the `inviteCode` line**

Insert these fields between `inviteCode` (line 15) and `vipDueAt` (line 16):

```prisma
currentPlayStreak  Int       @default(0) @map("current_play_streak")
maxPlayStreak      Int       @default(0) @map("max_play_streak")
lastPlayedAt       DateTime? @map("last_played_at") @db.Timestamptz
```

**Step 2: Add index for cron query performance**

Add to the `@@index` block (before `@@map("users")`):

```prisma
@@index([lastPlayedAt])
```

**Step 3: Generate Prisma client**

Run: `npm run prisma:generate`
Expected: Prisma client regenerated with new fields.

**Step 4: Create migration**

Run: `npx prisma migrate dev --schema prisma/schema --name add_play_streak_fields`
Expected: Migration created and applied. Three new columns on `users` table.

**Step 5: Verify migration**

Run: `npm run prisma:studio`
Check: `users` table shows `current_play_streak` (0), `max_play_streak` (0), `last_played_at` (null) columns.

**Step 6: Commit**

```
feat: add play streak fields to User model
```

---

### Task 2: Add `touchUserLastPlayedAt` mutation

**Files:**
- Modify: `src/models/user/user.mutation.ts`

**Step 1: Add the mutation function**

Append to `src/models/user/user.mutation.ts`:

```typescript
/** Touch user's lastPlayedAt to today (once per day, supports transactions) */
export async function touchUserLastPlayedAt(
  userId: string,
  tx?: TxClient
) {
  const client = tx ?? db;

  await client.$executeRawUnsafe(
    `UPDATE users SET last_played_at = now(), updated_at = now()
     WHERE id = $1
       AND (last_played_at IS NULL OR last_played_at::date < CURRENT_DATE)`,
    userId
  );
}
```

This uses a conditional raw UPDATE so it only writes once per day. The `CURRENT_DATE` comparison uses the database server's timezone.

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds.

**Step 3: Commit**

```
feat: add touchUserLastPlayedAt mutation
```

---

### Task 3: Call `touchUserLastPlayedAt` in `recordAnswer`

**Files:**
- Modify: `src/features/web/play/services/session.service.ts:1-2` (imports)
- Modify: `src/features/web/play/services/session.service.ts:209-249` (recordAnswer transaction)

**Step 1: Add import**

In `session.service.ts`, update the import from `@/models/user/user.mutation` (line 37):

```typescript
import { incrementUserExp, touchUserLastPlayedAt } from "@/models/user/user.mutation";
```

**Step 2: Add call inside recordAnswer transaction**

Inside the `db.$transaction(async (tx) => { ... })` block in `recordAnswer()`, add after the `updateSessionStatsOnAnswer` call (after line 248):

```typescript
    await touchUserLastPlayedAt(userId, tx);
```

The full transaction block becomes:
```typescript
  return db.$transaction(async (tx) => {
    await upsertRecord(/* ... existing ... */, tx);
    await updateSessionLevelStats(/* ... existing ... */, tx);
    await updateSessionStatsOnAnswer(/* ... existing ... */, tx);
    await touchUserLastPlayedAt(userId, tx);
  });
```

**Step 3: Verify build**

Run: `npm run build`
Expected: Build succeeds.

**Step 4: Manual smoke test**

1. Start dev server: `npm run dev`
2. Log in, start a game, answer one question
3. Check Prisma Studio: user's `last_played_at` should be set to now
4. Answer another question: `last_played_at` should not change (same day)

**Step 5: Commit**

```
feat: record lastPlayedAt on first answer each day
```

---

### Task 4: Create standalone Prisma client for scripts

**Files:**
- Create: `scripts/lib/db.ts`

**Step 1: Create `scripts/lib/` directory**

Run: `mkdir -p scripts/lib`

**Step 2: Write `scripts/lib/db.ts`**

```typescript
import { Pool } from "pg";
import { PrismaPg } from "@prisma/adapter-pg";
import { PrismaClient } from "../../src/generated/prisma/client.js";

/** Create a standalone Prisma client for scripts (no server-only restriction) */
export function createDb() {
  const connectionString = process.env.DATABASE_URL;

  if (!connectionString) {
    throw new Error("DATABASE_URL environment variable is not set");
  }

  const pool = new Pool({ connectionString });
  const adapter = new PrismaPg(pool);

  return new PrismaClient({ adapter });
}
```

Note: Uses relative import to generated Prisma client since `@/` alias may not resolve in standalone tsx scripts without extra config.

**Step 3: Verify it compiles**

Run: `npx tsx --env-file=.env -e "const { createDb } = require('./scripts/lib/db'); const db = createDb(); console.log('OK'); process.exit(0);"`

If path alias issues occur, adjust the import path.

**Step 4: Commit**

```
feat: add standalone Prisma client for scripts
```

---

### Task 5: Create play streak cron script

**Files:**
- Create: `scripts/cron/update-play-streaks.ts`

**Step 1: Create `scripts/cron/` directory**

Run: `mkdir -p scripts/cron`

**Step 2: Write `scripts/cron/update-play-streaks.ts`**

```typescript
import { createDb } from "../lib/db.js";

/**
 * Daily cron job: update play streaks for all users.
 *
 * Schedule: 0 2 * * * (2 AM daily)
 * Command:  docker exec your_container npm run cron:play-streaks
 *
 * Logic (based on last_played_at):
 *   - yesterday  → streak + 1, update max
 *   - today      → skip (already counted)
 *   - otherwise  → reset streak to 1
 */
async function main() {
  const start = Date.now();
  const db = createDb();

  try {
    // 1. Streak continues: played yesterday
    const continued = await db.$executeRaw`
      UPDATE users
      SET current_play_streak = current_play_streak + 1,
          max_play_streak = GREATEST(current_play_streak + 1, max_play_streak),
          updated_at = now()
      WHERE last_played_at IS NOT NULL
        AND last_played_at::date = CURRENT_DATE - INTERVAL '1 day'
    `;

    // 2. Streak broken: played before yesterday
    const reset = await db.$executeRaw`
      UPDATE users
      SET current_play_streak = 1,
          updated_at = now()
      WHERE last_played_at IS NOT NULL
        AND last_played_at::date < CURRENT_DATE - INTERVAL '1 day'
        AND current_play_streak != 1
    `;

    // 3. Played today: skip (no update needed)

    const elapsed = Date.now() - start;
    console.log(
      `[play-streaks] done in ${elapsed}ms — continued: ${continued}, reset: ${reset}`
    );
  } catch (error) {
    console.error("[play-streaks] failed:", error);
    process.exit(1);
  } finally {
    await db.$disconnect();
  }
}

main();
```

**Step 3: Add npm script to `package.json`**

Add to `"scripts"` section:

```json
"cron:play-streaks": "npx tsx --env-file=.env scripts/cron/update-play-streaks.ts"
```

**Step 4: Test the script locally**

Run: `npm run cron:play-streaks`
Expected: Outputs `[play-streaks] done in Xms — continued: 0, reset: 0` (no users have streaks yet).

**Step 5: Integration test with real data**

1. Start dev server, log in, answer one question (sets `last_played_at` to today)
2. Run: `npm run cron:play-streaks`
   Expected: `continued: 0, reset: 0` (today's play is skipped)
3. Manually backdate via Prisma Studio: set `last_played_at` to yesterday for that user
4. Run: `npm run cron:play-streaks`
   Expected: `continued: 1, reset: 0` — user's `current_play_streak` is now 1
5. Check Prisma Studio: `current_play_streak = 1`, `max_play_streak = 1`

**Step 6: Commit**

```
feat: add daily play streak cron script
```

---

### Task 6: Final verification and build check

**Step 1: Full build**

Run: `npm run build`
Expected: Build succeeds with no errors.

**Step 2: Lint**

Run: `npm run lint`
Expected: No lint errors.

**Step 3: Verify all new files**

- `prisma/schema/user.prisma` — 3 new fields + 1 new index
- `src/models/user/user.mutation.ts` — `touchUserLastPlayedAt` function
- `src/features/web/play/services/session.service.ts` — import + call in `recordAnswer`
- `scripts/lib/db.ts` — standalone Prisma client
- `scripts/cron/update-play-streaks.ts` — streak cron script
- `package.json` — `cron:play-streaks` script

**Step 4: Commit**

Only if any fixes were needed in previous steps.
