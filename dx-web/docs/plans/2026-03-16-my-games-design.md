# My Games Page & Sidebar Reorder Design

## Overview

Add a "我的游戏" page listing all games the logged-in user has played (deduplicated), and reorder the sidebar navigation.

## Data Layer

### Query: `getUserPlayedGames(userId)`

Location: `src/models/game-stats-total/game-stats-total.query.ts`

Source table: `GameStatsTotal` — one record per user+game, natural deduplication.

Order by `lastPlayedAt desc`.

Select: game info (id, name, description, mode, cover, category, user) + stats (highestScore, totalPlayTime).

### Return Type: `PlayedGameCard`

```ts
{
  id: string;
  name: string;
  description: string | null;
  mode: string;
  cover: { url: string } | null;
  category: { name: string } | null;
  user: { username: string } | null;
  highestScore: number;
  totalPlayTime: number;
}
```

## Page

Route: `src/app/(web)/hall/(main)/games/mine/page.tsx`

- Server component, calls `auth()` for userId
- Calls `getUserPlayedGames(userId)`
- Uses `PageTopBar` with title "我的游戏", subtitle "你玩过的所有课程游戏"
- Count badge + total count text (same pattern as favorites)
- 5-column grid of `PlayedGameCard` components
- Empty state: `Gamepad2` icon + "还没有玩过游戏，去发现课程游戏吧"

## Card Component

File: `src/features/web/hall/components/played-game-card.tsx`

Based on `FavoriteCard` layout with an added stats row:

- Same gradient cover fallback for games without a cover image
- Game name + description (2-line clamp)
- Stats row between description and play button:
  - `Trophy` icon + "最高 {score} 分" (highest score)
  - `Clock` icon + formatted play time (minutes/hours)
- Category label + play button at bottom

## Sidebar Changes

File: `src/features/web/hall/components/hall-sidebar.tsx`

First nav section reordered from:

```
LayoutDashboard  学习主页  /hall
Gamepad2         课程游戏  /hall/games
Star             我的收藏  /hall/favorites
Bell             消息通知  /hall/notifications
```

To:

```
Gamepad2         课程游戏  /hall/games
LayoutDashboard  我的主页  /hall
Gamepad2         我的游戏  /hall/games/mine
Star             我的收藏  /hall/favorites
Bell             消息通知  /hall/notifications
```

Changes:
- "学习主页" renamed to "我的主页", moved below "课程游戏"
- New "我的游戏" item added above "我的收藏" with `Gamepad2` icon
