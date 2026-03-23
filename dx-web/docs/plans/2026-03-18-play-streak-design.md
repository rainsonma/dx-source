# Play Streak Design

Track users' consecutive playing days with `currentPlayStreak`, `maxPlayStreak`, and `lastPlayedAt` on the User model. A daily cron job at 2 AM computes streaks.

## Definition

A "day played" = the user recorded at least one answer (GameRecord created/updated) on that calendar day.

## Schema Changes

Add 3 fields to `User` in `prisma/schema/user.prisma`:

```prisma
currentPlayStreak  Int       @default(0) @map("current_play_streak")
maxPlayStreak      Int       @default(0) @map("max_play_streak")
lastPlayedAt       DateTime? @map("last_played_at") @db.Timestamptz
```

## Recording `lastPlayedAt`

Set `lastPlayedAt = now()` inside the `recordAnswer()` transaction in `session.service.ts`. Use a conditional update to write only once per day:

```sql
UPDATE users SET last_played_at = now()
WHERE id = $1 AND (last_played_at IS NULL OR last_played_at::date < CURRENT_DATE)
```

New mutation: `touchUserLastPlayedAt(userId, tx)` in `src/models/user/user.mutation.ts`.

## Cron Job: Streak Calculation

**Schedule:** Daily at 2 AM
**Script:** `scripts/cron/update-play-streaks.ts`
**Command:** `npm run cron:play-streaks` → `npx tsx scripts/cron/update-play-streaks.ts`
**Docker:** `0 2 * * * docker exec your_container npm run cron:play-streaks`

### Logic

```
yesterday = date(now - 1 day)
today = date(now)

For each user where lastPlayedAt IS NOT NULL:
  playedDate = date(lastPlayedAt)

  if playedDate == today     → skip (already counted)
  if playedDate == yesterday → currentPlayStreak += 1
                                maxPlayStreak = GREATEST(currentPlayStreak + 1, maxPlayStreak)
  otherwise                  → currentPlayStreak = 1
```

### Implementation

Three bulk SQL updates (no per-user loops):

1. **Streak continues** — `WHERE last_played_at::date = yesterday`
   ```sql
   UPDATE users
   SET current_play_streak = current_play_streak + 1,
       max_play_streak = GREATEST(current_play_streak + 1, max_play_streak)
   WHERE last_played_at::date = $yesterday
   ```

2. **Streak broken** — `WHERE last_played_at::date < yesterday`
   ```sql
   UPDATE users
   SET current_play_streak = 1
   WHERE last_played_at IS NOT NULL
     AND last_played_at::date < $yesterday
     AND current_play_streak != 1
   ```

3. **Already counted today** — skip (no update needed)

Script logs summary and exits with code 0 on success, 1 on failure.

## Standalone DB Client

`scripts/lib/db.ts` — creates a Prisma client without `server-only` import (which blocks non-Next.js usage). Same connection setup as `src/lib/db.ts` but without the global singleton caching. Future cron scripts reuse this.

## File Changes

| File | Change |
|------|--------|
| `prisma/schema/user.prisma` | Add 3 fields |
| New migration | `current_play_streak`, `max_play_streak`, `last_played_at` |
| `src/models/user/user.mutation.ts` | Add `touchUserLastPlayedAt(userId, tx)` |
| `src/features/web/play/services/session.service.ts` | Call `touchUserLastPlayedAt` in `recordAnswer` |
| `scripts/lib/db.ts` | Standalone Prisma client for scripts |
| `scripts/cron/update-play-streaks.ts` | Streak calculation script |
| `package.json` | Add `cron:play-streaks` script |
