# My Games Page Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a "我的游戏" page listing all games the user has played (deduplicated), and reorder the sidebar navigation.

**Architecture:** New query on `GameStatsTotal` (natural dedup — one row per user+game) joined with `Game`. New page at `/hall/games/mine` following the `favorites/page.tsx` pattern. New card component with play stats. Sidebar nav items reordered.

**Tech Stack:** Next.js 16 App Router, Prisma, TailwindCSS, Lucide icons

**Design doc:** `docs/plans/2026-03-16-my-games-design.md`

---

### Task 1: Add `getUserPlayedGames` Query

**Files:**
- Modify: `src/models/game-stats-total/game-stats-total.query.ts`

**Step 1: Add the type and query function**

Append to `src/models/game-stats-total/game-stats-total.query.ts`:

```ts
export type PlayedGameCard = {
  id: string;
  name: string;
  description: string | null;
  mode: string;
  cover: { url: string } | null;
  category: { name: string } | null;
  user: { username: string } | null;
  highestScore: number;
  totalPlayTime: number;
};

/** Get all games a user has played, newest first */
export async function getUserPlayedGames(
  userId: string
): Promise<PlayedGameCard[]> {
  const stats = await db.gameStatsTotal.findMany({
    where: { userId },
    select: {
      highestScore: true,
      totalPlayTime: true,
      game: {
        select: {
          id: true,
          name: true,
          description: true,
          mode: true,
          cover: { select: { url: true } },
          category: { select: { name: true } },
          user: { select: { username: true } },
        },
      },
    },
    orderBy: { lastPlayedAt: "desc" },
  });

  return stats.map((s) => ({
    ...s.game,
    highestScore: s.highestScore,
    totalPlayTime: s.totalPlayTime,
  }));
}
```

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 3: Commit**

```
feat: add getUserPlayedGames query for my-games page
```

---

### Task 2: Create `PlayedGameCard` Component

**Files:**
- Create: `src/features/web/hall/components/played-game-card.tsx`

**Step 1: Create the component**

Based on `src/features/web/hall/components/favorite-card.tsx` with a stats row added.

```tsx
import Link from "next/link";
import { Play, Trophy, Clock } from "lucide-react";
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";
import { formatPlayTime } from "@/lib/format";
import type { PlayedGameCard as PlayedGameCardType } from "@/models/game-stats-total/game-stats-total.query";

const GRADIENT_COVERS = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

/** Deterministic gradient based on id hash */
function getGradient(id: string) {
  let hash = 0;
  for (const ch of id) hash = (hash * 31 + ch.charCodeAt(0)) | 0;
  return GRADIENT_COVERS[Math.abs(hash) % GRADIENT_COVERS.length];
}

export function PlayedGameCard({ game }: { game: PlayedGameCardType }) {
  const modeLabel = GAME_MODE_LABELS[game.mode as GameMode] ?? game.mode;

  return (
    <Link
      href={`/hall/games/${game.id}`}
      className="flex w-full flex-col overflow-hidden rounded-xl border border-slate-200 bg-white transition-shadow hover:shadow-md"
    >
      {game.cover ? (
        <img
          src={game.cover.url}
          alt={game.name}
          className="h-[180px] w-full object-cover"
        />
      ) : (
        <div
          className={`flex h-[180px] w-full items-center justify-center ${getGradient(game.id)}`}
        >
          <span className="text-lg font-bold text-white/80">{modeLabel}</span>
        </div>
      )}

      <div className="flex flex-1 flex-col justify-between gap-2 px-3.5 py-3">
        <div className="flex flex-col gap-1">
          <h4 className="text-sm font-bold text-slate-900">{game.name}</h4>
          <p className="line-clamp-2 text-[11px] leading-[1.4] text-slate-500">
            {game.description}
          </p>
        </div>
        <div className="flex items-center gap-3 text-[11px] text-slate-400">
          <span className="flex items-center gap-1">
            <Trophy className="h-3 w-3" />
            最高 {game.highestScore} 分
          </span>
          <span className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {formatPlayTime(game.totalPlayTime)}
          </span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-medium text-slate-500">
            {game.category?.name ?? modeLabel}
          </span>
          <span className="flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1.5 text-[11px] font-semibold text-white">
            <Play className="h-3 w-3" />
            开始
          </span>
        </div>
      </div>
    </Link>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 3: Commit**

```
feat: add PlayedGameCard component with stats display
```

---

### Task 3: Create My Games Page

**Files:**
- Create: `src/app/(web)/hall/(main)/games/mine/page.tsx`

**Step 1: Create the page**

Follow the same pattern as `src/app/(web)/hall/(main)/favorites/page.tsx`.

```tsx
import { Gamepad2 } from "lucide-react";

import { auth } from "@/lib/auth";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { PlayedGameCard } from "@/features/web/hall/components/played-game-card";
import { getUserPlayedGames } from "@/models/game-stats-total/game-stats-total.query";

export default async function MyGamesPage() {
  const session = await auth();
  const userId = session?.user?.id;

  const games = userId ? await getUserPlayedGames(userId) : [];

  return (
    <div className="flex h-full flex-col gap-6 px-8 py-7">
      <PageTopBar
        title="我的游戏"
        subtitle="你玩过的所有课程游戏"
        searchPlaceholder="搜索游戏..."
      />

      {/* Filter row */}
      <div className="flex items-center justify-between">
        <span className="rounded-full bg-teal-600 px-5 py-2 text-[13px] font-semibold text-white">
          全部 ({games.length})
        </span>
        <span className="text-[13px] text-slate-400">
          共 {games.length} 个游戏
        </span>
      </div>

      {games.length > 0 ? (
        <div className="grid grid-cols-5 gap-4">
          {games.map((game) => (
            <PlayedGameCard key={game.id} game={game} />
          ))}
        </div>
      ) : (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 text-slate-400">
          <Gamepad2 className="h-12 w-12 stroke-1" />
          <p className="text-sm">还没有玩过游戏，去发现课程游戏吧</p>
        </div>
      )}
    </div>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 3: Commit**

```
feat: add my-games page listing all played games
```

---

### Task 4: Reorder Sidebar Navigation

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx`

**Step 1: Update `navSections` first section**

In `src/features/web/hall/components/hall-sidebar.tsx`, change the first section's `items` array from:

```ts
{ icon: LayoutDashboard, label: "学习主页", href: "/hall" },
{ icon: Gamepad2, label: "课程游戏", href: "/hall/games" },
{ icon: Star, label: "我的收藏", href: "/hall/favorites" },
{ icon: Bell, label: "消息通知", href: "/hall/notifications" },
```

To:

```ts
{ icon: Gamepad2, label: "课程游戏", href: "/hall/games" },
{ icon: LayoutDashboard, label: "我的主页", href: "/hall" },
{ icon: Gamepad2, label: "我的游戏", href: "/hall/games/mine" },
{ icon: Star, label: "我的收藏", href: "/hall/favorites" },
{ icon: Bell, label: "消息通知", href: "/hall/notifications" },
```

Changes:
- "课程游戏" moved to first position
- "学习主页" renamed to "我的主页", now second
- New "我的游戏" item added third with `Gamepad2` icon
- "我的收藏" and "消息通知" stay in order

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds

**Step 3: Verify in browser**

- Navigate to `/hall` — sidebar shows new order
- Click "我的游戏" — navigates to `/hall/games/mine`
- Page shows empty state if no games played, or card grid with stats if games exist

**Step 4: Commit**

```
feat: reorder sidebar nav, add my-games link, rename to 我的主页
```
