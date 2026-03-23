# User Favorites Design

Toggle-favorite for games: click the 收藏 button in the game detail hero section to add/remove from `user_favorites`, with toast feedback. Wire real favorite data into the favorites page.

## Constraints

- All hall routes are auth-protected by `proxy.ts` middleware — no unauthenticated guard needed
- Code-level FK via `assertFK()` (relationMode = "prisma", no DB-level FK constraints)
- Abuse prevention: client-side 2-second cooldown per effective click + DB composite unique constraint as idempotency guard
- Rapid clicks within cooldown show `toast.warning("操作频繁，请稍后再试")`

## Data Layer

### `src/models/user-favorite/user-favorite.query.ts`

- `isGameFavorited(userId, gameId)` → `boolean`
- `getUserFavorites(userId)` → list of favorited games (id, name, description, mode, cover, category, creator)

### `src/models/user-favorite/user-favorite.mutation.ts`

- `toggleFavorite(userId, gameId)` → `{ favorited: boolean }`
  - Check existence via composite unique `userId_gameId`
  - Delete if exists, create with `ulid()` + `assertFK` if not
  - Single transaction

## Service + Action + Hook

### `src/features/web/games/services/favorite.service.ts`

- `toggleFavoriteService(userId, gameId)` → `{ success: true; favorited: boolean } | { error: string }`

### `src/features/web/games/actions/favorite.action.ts`

- `toggleFavoriteAction(gameId)` → gets userId from `auth()`, delegates to service
- Returns `{ success: true; favorited: boolean } | { error: string }`

### `src/features/web/games/hooks/use-favorite-toggle.ts`

- `useFavoriteToggle(gameId, gameName, initialFavorited)`
- Optimistic local state toggle
- `useTransition` for `isPending` (natural debounce)
- 2-second cooldown after each effective click; clicks within cooldown show `toast.warning("操作频繁，请稍后再试")`
- Success: `toast.success("已收藏「游戏名」")` / `toast.success("已取消收藏「游戏名」")`
- Failure: `toast.error(error)` with state rollback

## Game Detail Page Wiring

### `src/app/(web)/hall/(main)/games/[id]/page.tsx`

- Add `isGameFavorited(userId, gameId)` to existing `Promise.all`
- Pass `isFavorited` to `GameDetailContent`

### `src/features/web/games/components/game-detail-content.tsx`

- Accept `isFavorited` prop
- Initialize `useFavoriteToggle(game.id, game.name, isFavorited)` hook
- Pass `isFavorited`, `onFavoriteToggle`, `isFavoritePending` to `HeroCard`

### `src/features/web/games/components/hero-card.tsx`

- New props: `isFavorited`, `onFavoriteToggle`, `isFavoritePending`
- Favorited: filled Heart (rose tint) + "已收藏"
- Not favorited: outline Heart + "收藏"
- `disabled={isFavoritePending}`

## Favorites Page

### `src/app/(web)/hall/(main)/favorites/page.tsx`

- Convert to server component fetching real data
- `auth()` → `getUserFavorites(userId)`
- Total count in filter row
- Cards: cover (image or gradient fallback), name, description, category, creator
- Each card links to `/hall/games/[id]`
- Empty state: "还没有收藏，去发现喜欢的游戏吧"

### `src/features/web/hall/components/favorite-card.tsx`

- Extracted component with typed props from query result
