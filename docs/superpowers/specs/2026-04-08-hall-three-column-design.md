# Hall Page Three-Column Layout Redesign

**Date:** 2026-04-08
**Status:** Approved

## Summary

Redesign the hall dashboard's main content row from a 2-column unequal layout (我的学习进度 + 今日挑战) into a 3-column equal-width layout with a new 今日明星榜 leaderboard card.

## Current State

```
[ GameProgressCard (flex-1) ][ DailyChallengeCard (w-80 shrink-0) ]
```

- Left: 我的学习进度 — paginated game session progress list
- Right: 今日挑战 — teal-600 solid card with play button, fixed 320px width

## New Layout

### Desktop (lg+)

```
[ 我的学习进度 (1/3) ][ 今日打卡 (1/3) ][ 今日明星榜 (1/3) ]
```

- `grid grid-cols-3 gap-5` — three equal columns

### Mobile

Stacked vertically in this order:
1. 今日明星榜 (shows first on mobile)
2. 我的学习进度
3. 今日打卡

Achieved via `order-first lg:order-none` on the right column.

## Column Details

### Left Column — 我的学习进度

No functional changes. Only layout change: remove `flex-1`, let grid handle width. Same paginated game session list with progress bars.

### Center Column — 今日打卡

One card titled "今日打卡" containing two task items:

- **Card header**: Flame icon + "今日打卡" title, styled consistently with GameProgressCard header
- **Task 1**: Teal-600 solid block — keeps original DailyChallengeCard text and Play button linking to `/hall/games`
- **Task 2**: Teal gradient block (`bg-gradient-to-br from-teal-500 to-teal-700`) — "前往「斗学社」发表一条英文动态贴，进步需要坚持不懈!" with "去发帖" button linking to `/hall/community`
- Tasks stacked with `gap-4` inside card body

### Right Column — 今日明星榜

A leaderboard card showing today's top 50 users:

- **Card header**: Trophy/Star icon + "今日明星榜" title + "查看全部" link to `/hall/leaderboard`
- **Type tabs**: Two `TabPill` buttons — 经验 (Zap icon) / 时长 (Clock icon), default to 经验
- **Podium**: Mini version of existing `leaderboard-podium.tsx` for top 3 — same 2nd/1st/3rd arrangement, same medal colors (silver/gold/bronze), scaled down to fit column
- **List**: Scrollable list for rank 4–50, each row: rank number, avatar, name, value
- **My Rank**: Sticky bar at bottom showing current user's rank with teal-600 border
- **Loading state**: Spinner consistent with existing leaderboard
- **Period**: Always `day` (no period tabs)

## Data Strategy

**Approach**: Reuse existing leaderboard API, no backend changes.

- Fetch: `GET /api/leaderboard?type=exp&period=day` (default)
- Switch: `GET /api/leaderboard?type=playtime&period=day` (on tab change)
- Slice response to first 50 entries on the client (API returns up to 100)

## File Changes

### Modified Files

| File | Change |
|------|--------|
| `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` | Change main content row from flex to 3-column grid |
| `dx-web/src/features/web/hall/components/daily-challenge-card.tsx` | Refactor into grouped card: header "今日打卡" + two task items |
| `dx-web/src/features/web/hall/components/game-progress-card.tsx` | Remove `flex-1`, let grid handle width |

### New Files

| File | Purpose |
|------|---------|
| `dx-web/src/features/web/hall/components/today-stars-card.tsx` | 今日明星榜 card with mini podium, list, my-rank, type tabs |
| `dx-web/src/features/web/hall/hooks/use-today-stars.ts` | Hook to fetch leaderboard data with `period=day`, type switching, loading state |

### Reused (import, not copy)

- `formatValue` / `formatPlayTime` from leaderboard helpers / `@/lib/format`
- `getAvatarColor` from `@/lib/avatar`
- `Avatar` from `@/components/ui/avatar`
- `TabPill` from `@/components/in/tab-pill`
- `LeaderboardEntry` / `LeaderboardResult` types from leaderboard feature

### No Backend Changes

Existing `GET /api/leaderboard` endpoint covers all requirements.

## Styling

- Teal theme throughout: `teal-600` accents, `teal-50` backgrounds, teal gradients
- Podium colors: `amber-400` (gold/1st), `slate-300` (silver/2nd), `amber-600` (bronze/3rd)
- Card style: `rounded-[14px] border border-border bg-card` consistent with existing cards
- Must pass ESLint with zero warnings

## Constraints

- No breaking changes to existing functionality
- No backend modifications
- No lint issues
- Mobile-first responsive design
