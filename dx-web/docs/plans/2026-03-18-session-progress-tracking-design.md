# Session Progress Tracking Fields

## Date: 2026-03-18

## Goal

Add progress tracking fields to GameSessionTotal and GameSessionLevel so that session progress (how many levels/items completed out of total) is explicitly stored in the database.

## New Fields

| Model | Field | Type | When Set |
|-------|-------|------|----------|
| GameSessionTotal | totalLevelsCount | Int @default(0) | On session creation — count of GameLevel for the game |
| GameSessionTotal | playedLevelsCount | Int @default(0) | Incremented on each completeLevel() |
| GameSessionLevel | totalItemsCount | Int @default(0) | On level session creation — count of ContentItem filtered by degree |
| GameSessionLevel | playedItemsCount | Int @default(0) | Incremented on each recordAnswer() (correct or wrong only, not skip) |

## Changes

### 1. Prisma Schema

- `game-session-total.prisma` — add `totalLevelsCount Int @default(0)` and `playedLevelsCount Int @default(0)`
- `game-session-level.prisma` — add `totalItemsCount Int @default(0)` and `playedItemsCount Int @default(0)`
- Create migration

### 2. Model Layer

- `game-session-total.mutation.ts` — update `createSession()` to accept and set `totalLevelsCount`; add helper to increment `playedLevelsCount`
- `game-session-level.mutation.ts` — update `createSessionLevel()` to accept and set `totalItemsCount`; add helper to increment `playedItemsCount`
- `game-level/game-level.query.ts` — add `countLevelsByGameId(gameId)` query
- `content-item/content-item.query.ts` — add `countContentItemsByLevelId(levelId, contentTypes?)` query

### 3. Service Layer (session.service.ts)

- `startGameSession()` — query level count, pass `totalLevelsCount` to `createSession()`
- `startSessionLevel()` — query content item count (filtered by degree), pass `totalItemsCount` to `createSessionLevel()`
- `completeLevel()` — inside existing transaction, increment `playedLevelsCount` on GameSessionTotal
- `recordAnswer()` — inside existing transaction, increment `playedItemsCount` on GameSessionLevel

### 4. Rules Update

- `rules/GameLSRWRule.md` — update Progress Tracking section to document the new fields

## Notes

- Skip does NOT increment `playedItemsCount` — only answered items (correct/wrong) count
- `totalItemsCount` is filtered by the degree's content types, matching what the player actually sees
- These are snapshot values set at creation time (totalLevelsCount, totalItemsCount) — they reflect the game state when the session started
- No client-side changes needed — these are server-side tracking fields
