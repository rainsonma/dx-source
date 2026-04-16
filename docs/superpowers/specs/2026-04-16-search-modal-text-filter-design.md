# Search Modal Text Filter Design

## Goal

Improve the Cmd+K search modal so users can either search the games list by text or jump directly to a specific game.

## Current Behavior

- Cmd+K opens a cmdk-based dialog
- Typing triggers debounced API search, results displayed as game items
- cmdk auto-selects the first game item
- Clicking/Enter on any item navigates to `/hall/games/{id}`
- The `/api/games` list endpoint has no text search (`q`) parameter

## New Behavior

### Search Modal

- When user types text, a **top item** appears above the game results:
  - Shows `搜索 "word"` with a search icon
  - Default-selected (highlighted) via controlled cmdk `value`
- Game results below are grouped under heading **"猜你想学"** (replaces "搜索结果")
- "最近玩过" heading unchanged when query is empty

### Top Item (Enter/Click)

1. Store typed text in zustand store via `setQ(query)`
2. Navigate to `/hall/games` (if not already there)
3. Close modal
4. Games list page reads `q` from store, passes to API alongside existing filters
5. Search trigger in top bar shows typed text with (X) clear button

### Game Item (Arrow Down + Enter/Click)

Navigate to `/hall/games/{id}` — unchanged from current behavior.

### Clear Button

- (X) on the search trigger calls `clearQ()`
- Games list refreshes without text filter
- Trigger reverts to default placeholder

## Architecture

### Shared State

Zustand store at `features/web/games/stores/game-search-store.ts`:

```
q: string        — current search text
setQ(q: string)  — set search text
clearQ()         — reset to ""
```

### Backend Changes

**`game_request.go`** — Add `Q string` field (`form:"q"`) to `ListGamesRequest`.

**`game_controller.go`** — Pass `req.Q` to `ListPublishedGames`.

**`game_service.go`** — Add `q string` parameter to `ListPublishedGames`; when non-empty, add `WHERE name ILIKE '%q%'`.

### Frontend Changes

**`game-search-dialog.tsx`**
- Add top `CommandItem` with `value="__search__"`, shows search icon + `搜索 "{query}"`
- Use controlled `value`/`onValueChange` on `Command` to default-select top item when query is non-empty
- Top item `onSelect`: call `setQ(query)`, navigate to `/hall/games`, close modal
- Game item `onSelect`: navigate to `/hall/games/{id}` (unchanged)
- Change group heading from `"搜索结果"` to `"猜你想学"`

**`game-search-trigger.tsx`**
- Read `q` from zustand store + `usePathname()`
- When `q` is set AND pathname is `/hall/games`: show text with (X) clear button
- (X) calls `clearQ()` with `stopPropagation` to avoid opening modal
- Otherwise: default placeholder

**`games-page-content.tsx`**
- Read `q` from zustand store
- Pass `q` to `useInfinitePublicGames` as part of filters

**`use-infinite-public-games.ts`**
- Add `q?: string` to `Filters` type
- Include `q` in URL params when non-empty

### Selection Logic (cmdk)

| Query State | Top Item | Group | Selection |
|-------------|----------|-------|-----------|
| Empty | Hidden | "最近玩过" (recent games) | cmdk default (first item) |
| Non-empty | Visible, `搜索 "{query}"` | "猜你想学" (search results) | Controlled: top item selected |

Arrow keys change selection via `onValueChange`. Enter triggers the selected item's `onSelect`.

## Edge Cases

- **Navigate away from `/hall/games`**: trigger reverts to default placeholder; store retains `q` so returning restores filter
- **Page refresh**: `q` lost (in-memory zustand) — consistent with existing filter behavior
- **Open modal with active `q`**: modal starts empty (existing reset-on-close); new query replaces old
- **Empty search results**: top item still shows `搜索 "{query}"` so user can still filter the games list
- **`q` combined with category/press/mode**: all sent together to API; backend applies all filters with AND logic
