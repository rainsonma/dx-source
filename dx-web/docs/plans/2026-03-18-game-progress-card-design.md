# Game Progress Card — Design

Wire real data into the hall dashboard's "我的游戏进度" card.

## Requirements

- List all user's `GameSessionTotal` records in `lastPlayedAt` desc order
- Progress % = `playedLevelsCount / totalLevelsCount`
- Client-side pagination using `DataTablePagination`
- "查看全部" links to `/hall/games/mine`

## Data Flow

```
page.tsx (SSR)
  → hall.service.ts: fetchDashboardStats() includes session totals
    → getUserSessionTotals(userId) query
      → GameSessionTotal + Game (name, mode)
      → ordered by lastPlayedAt desc
  → pass sessions[] to GameProgressCard (client component)
```

## Query

New `getUserSessionTotals` in `game-session-total.query.ts`:

- Joins `game.name`, `game.mode`
- Selects: `id`, `playedLevelsCount`, `totalLevelsCount`, `lastPlayedAt`, `endedAt`, `score`, `exp`, `degree`, `pattern`, `gameId`
- Orders by `lastPlayedAt` desc

## Component

`GameProgressCard` (client component):

- Props: `sessions[]` from SSR page
- Each row: game name, mode label, progress bar, percentage
- Pagination via `DataTablePagination`
- "查看全部" → `<Link href="/hall/games/mine">`
- Empty state when no sessions

## Files Changed

| File | Change |
|------|--------|
| `game-session-total.query.ts` | Add `getUserSessionTotals` |
| `hall.service.ts` | Include sessions in `fetchDashboardStats` |
| `game-progress-card.tsx` | Rewrite with real data + pagination |
| `page.tsx` (home) | Pass sessions to `GameProgressCard` |
