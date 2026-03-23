# Game Record: Add userId and duration Fields

## Summary

Add `userId` and `duration` fields to the `GameRecord` model to enable direct per-user queries and per-item answer timing analytics.

## Schema Change

Add two fields to `GameRecord`:

- `userId` (Char(26), indexed) — denormalized from `GameSessionTotal.userId` for direct query access
- `duration` (Int, default 0) — seconds spent on the content item from first display to answer completion

No Prisma relation added (project uses code-level `assertFK()`).

## Data Flow

### Client-side timing (use-lsrw.ts)

- Add `itemStartTimeRef = useRef<number>(Date.now())`
- Reset on `currentIndex` change in the existing useEffect
- On item completion: `Math.round((Date.now() - itemStartTimeRef.current) / 1000)`
- Pass `duration` to `recordAnswerAction()`

### Server-side userId (session.service.ts)

- `recordAnswer()` already calls `requireUserId()` — pass the resolved `userId` into `upsertRecord()`
- No extra DB lookup needed

### Action layer (session.action.ts)

- Add `duration: number` to `recordAnswerAction` params — passes through to service

### Mutation layer (game-record.mutation.ts)

- `upsertRecord()` accepts `userId` and `duration` in its data param
- `assertFK` adds check for `users` table with the userId

## Files Changed

1. `prisma/schema/game-record.prisma` — add userId, duration fields + index
2. New migration — two new columns
3. `src/models/game-record/game-record.mutation.ts` — accept userId and duration
4. `src/features/web/play/services/session.service.ts` — pass userId and duration through
5. `src/features/web/play/actions/session.action.ts` — add duration to recordAnswerAction params
6. `src/features/web/play/hooks/use-lsrw.ts` — track item start time, compute duration on submission

## Design Decisions

- **Denormalized userId**: Enables direct GameRecord queries without joining through GameSessionTotal. Consistent with GameStatsTotal/GameStatsLevel which also carry userId.
- **Client-side timing**: More accurate than server-side timestamp diffs — correctly excludes pauses, overlays, and network latency.
- **Integer seconds**: Consistent with existing `playTime` fields on session models. Sufficient precision for learning analytics.
- **Duration only for answered items**: Skipped items don't create GameRecords per game rules, so no skip duration tracking needed.
