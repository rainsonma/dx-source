# Game Fuzzy Search Design

## Overview

Add game fuzzy search to the hall top bar using ShadCN Command (cmdk) dialog. Users click the search bar or press `⌘K` to open a command palette that searches published games by name with debounced server-side queries.

## Requirements

- Cmd+K / Ctrl+K keyboard shortcut to open search dialog
- Debounced (300ms) server action on each keystroke
- Search games by `name` field (case-insensitive)
- Suggestions show: game name + category + mode label
- Selecting a suggestion navigates to `/hall/games/{id}`
- Default state (before typing): show user's recently played games (logged-in only)

## Architecture — Approach A: Two Server Actions

Chosen over single combined action (cleaner separation) and API route (project uses server actions pattern).

## Data Layer

### `searchPublishedGames(query, limit)` — `src/models/game/game.query.ts`

- Params: `query: string`, `limit: number` (default 8)
- Filters: `status: "published"`, `isActive: true`, `name` contains query (Prisma `contains` + `mode: "insensitive"`)
- Returns: `GameSearchResult[]` — `{ id, name, mode, category: { name } }`
- Ordered by `createdAt: "desc"`

### `getRecentlyPlayedGames(userId, limit)` — `src/models/game-stats-total/game-stats-total.query.ts`

- Params: `userId: string`, `limit: number` (default 5)
- Joins through `game` relation for name/mode/category
- Filters: game `status: "published"`, `isActive: true`
- Returns: same `GameSearchResult[]` shape
- Ordered by `lastPlayedAt: "desc"`

### Shared type

```ts
type GameSearchResult = {
  id: string;
  name: string;
  mode: string;
  category: { name: string } | null;
};
```

## Server Actions — `src/features/web/hall/actions/game-search.action.ts`

### `searchGamesAction(query: string)`

- Trims query, returns empty array if blank
- Calls `searchPublishedGames(query, 8)`
- Returns: `{ games: GameSearchResult[], error?: string }`

### `getRecentGamesAction()`

- Gets current user session; returns empty array if not logged in
- Calls `getRecentlyPlayedGames(userId, 5)`
- Returns: `{ games: GameSearchResult[], error?: string }`

## Client Components

### `useGameSearch` hook — `src/features/web/hall/hooks/use-game-search.ts`

- State: `query`, `results`, `recentGames`, `isLoading`, `isOpen`
- On dialog open: calls `getRecentGamesAction()` once, caches in state
- On query change: debounces 300ms, calls `searchGamesAction(query)`
- When query cleared: shows recent games again
- Exposes: `{ query, setQuery, results, isLoading, isOpen, setIsOpen }`

### `GameSearchDialog` — `src/features/web/hall/components/game-search-dialog.tsx`

- Client component using `CommandDialog`
- `CommandInput` — placeholder "输入课程名称搜索..."
- `CommandList` with:
  - `CommandEmpty` — "没有找到相关课程"
  - `CommandGroup` heading "最近玩过" (when no query, logged-in)
  - `CommandGroup` heading "搜索结果" (when query present)
- Each `CommandItem`: game name + category name + mode label (from `GAME_MODE_LABELS`)
- On select: `router.push(/hall/games/${id})`, close dialog
- Registers `⌘K` / `Ctrl+K` via `useEffect`

### `GameSearchTrigger` — `src/features/web/hall/components/game-search-trigger.tsx`

- Client component wrapping the search bar div
- On click: opens `GameSearchDialog`
- Shows `⌘K` shortcut hint in the bar
- Renders `GameSearchDialog` as a child

### Changes to `TopActions` — `src/features/web/hall/components/top-actions.tsx`

- Replace static search `<div>` with `<GameSearchTrigger />`
- `TopActions` remains a server component

## File Summary

| File | Change |
|------|--------|
| `src/models/game/game.query.ts` | Add `searchPublishedGames()` + `GameSearchResult` type |
| `src/models/game-stats-total/game-stats-total.query.ts` | Add `getRecentlyPlayedGames()` |
| `src/features/web/hall/actions/game-search.action.ts` | New — two server actions |
| `src/features/web/hall/hooks/use-game-search.ts` | New — search state + debounce |
| `src/features/web/hall/components/game-search-dialog.tsx` | New — CommandDialog UI |
| `src/features/web/hall/components/game-search-trigger.tsx` | New — clickable search bar |
| `src/features/web/hall/components/top-actions.tsx` | Replace static div with trigger |
