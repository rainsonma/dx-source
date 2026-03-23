# User Favorites Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a toggleable favorite button on the game detail hero section that persists to `user_favorites`, with toast feedback, and wire real favorite data into the favorites page.

**Architecture:** Server action pattern — model layer (query + mutation) → service → server action → client hook with `useTransition`. Favorites page is a server component fetching real data.

**Tech Stack:** Prisma (user_favorites model already exists), Next.js server actions, React `useTransition`, sonner toasts, lucide-react icons, TailwindCSS v4.

---

### Task 1: Model — Query Functions

**Files:**
- Create: `src/models/user-favorite/user-favorite.query.ts`

**Step 1: Create the query file**

```typescript
import "server-only";

import { db } from "@/lib/db";

/** Check whether a user has favorited a specific game */
export async function isGameFavorited(
  userId: string,
  gameId: string
): Promise<boolean> {
  const row = await db.userFavorite.findUnique({
    where: { userId_gameId: { userId, gameId } },
    select: { id: true },
  });

  return row !== null;
}

export type FavoriteGameCard = {
  id: string;
  name: string;
  description: string | null;
  mode: string;
  cover: { url: string } | null;
  category: { name: string } | null;
  user: { username: string } | null;
};

/** Get all games a user has favorited, newest first */
export async function getUserFavorites(
  userId: string
): Promise<FavoriteGameCard[]> {
  const favorites = await db.userFavorite.findMany({
    where: { userId },
    select: {
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
    orderBy: { createdAt: "desc" },
  });

  return favorites.map((f) => f.game);
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors related to user-favorite query.

**Step 3: Commit**

```
feat: add user-favorite query model
```

---

### Task 2: Model — Mutation Function

**Files:**
- Create: `src/models/user-favorite/user-favorite.mutation.ts`

**Step 1: Create the mutation file**

```typescript
import "server-only";

import { ulid } from "ulid";
import { db } from "@/lib/db";
import { assertFK } from "@/lib/assert-fk";

/** Toggle a game favorite — add if absent, remove if present */
export async function toggleFavorite(
  userId: string,
  gameId: string
): Promise<{ favorited: boolean }> {
  return db.$transaction(async (tx) => {
    await assertFK(tx, [
      { table: "users", id: userId },
      { table: "games", id: gameId },
    ]);

    const existing = await tx.userFavorite.findUnique({
      where: { userId_gameId: { userId, gameId } },
      select: { id: true },
    });

    if (existing) {
      await tx.userFavorite.delete({ where: { id: existing.id } });
      return { favorited: false };
    }

    await tx.userFavorite.create({
      data: { id: ulid(), userId, gameId },
    });

    return { favorited: true };
  });
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: add user-favorite toggle mutation
```

---

### Task 3: Service

**Files:**
- Create: `src/features/web/games/services/favorite.service.ts`

**Step 1: Create the service file**

```typescript
import "server-only";

import { toggleFavorite } from "@/models/user-favorite/user-favorite.mutation";

type ToggleResult =
  | { success: true; favorited: boolean }
  | { error: string };

/** Toggle the favorite state for a game */
export async function toggleFavoriteService(
  userId: string,
  gameId: string
): Promise<ToggleResult> {
  try {
    const result = await toggleFavorite(userId, gameId);
    return { success: true, favorited: result.favorited };
  } catch {
    return { error: "操作失败，请重试" };
  }
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: add favorite toggle service
```

---

### Task 4: Server Action

**Files:**
- Create: `src/features/web/games/actions/favorite.action.ts`

**Step 1: Create the action file**

```typescript
"use server";

import { auth } from "@/lib/auth";
import { toggleFavoriteService } from "@/features/web/games/services/favorite.service";

export type ToggleFavoriteResult =
  | { success: true; favorited: boolean }
  | { error: string };

/** Server action to toggle a game's favorite state */
export async function toggleFavoriteAction(
  gameId: string
): Promise<ToggleFavoriteResult> {
  const session = await auth();
  if (!session?.user?.id) {
    return { error: "请先登录" };
  }

  return toggleFavoriteService(session.user.id, gameId);
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: add toggle-favorite server action
```

---

### Task 5: Client Hook

**Files:**
- Create: `src/features/web/games/hooks/use-favorite-toggle.ts`

**Step 1: Create the hook file**

```typescript
"use client";

import { useState, useRef, useTransition, useCallback } from "react";
import { toast } from "sonner";
import { toggleFavoriteAction } from "@/features/web/games/actions/favorite.action";

const COOLDOWN_MS = 2000;

/** Toggle favorite with optimistic UI, 2-second cooldown, and toast feedback */
export function useFavoriteToggle(
  gameId: string,
  gameName: string,
  initialFavorited: boolean
) {
  const [favorited, setFavorited] = useState(initialFavorited);
  const [isPending, startTransition] = useTransition();
  const lastToggleRef = useRef(0);

  const toggle = useCallback(() => {
    const now = Date.now();
    if (now - lastToggleRef.current < COOLDOWN_MS) {
      toast.warning("操作频繁，请稍后再试");
      return;
    }

    if (isPending) return;

    const prev = favorited;
    setFavorited(!prev);
    lastToggleRef.current = now;

    startTransition(async () => {
      const result = await toggleFavoriteAction(gameId);

      if ("error" in result) {
        setFavorited(prev);
        toast.error(result.error);
        return;
      }

      setFavorited(result.favorited);
      toast.success(
        result.favorited
          ? `已收藏「${gameName}」`
          : `已取消收藏「${gameName}」`
      );
    });
  }, [gameId, gameName, favorited, isPending, startTransition]);

  return { favorited, toggle, isPending };
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: add use-favorite-toggle hook
```

---

### Task 6: Wire into Hero Card

**Files:**
- Modify: `src/features/web/games/components/hero-card.tsx:19-109`

**Step 1: Update HeroCard props and button**

Add three new props to `HeroCard`:
- `isFavorited: boolean`
- `onFavoriteToggle: () => void`
- `isFavoritePending: boolean`

Update the 收藏 button (lines 97-103):
- `onClick={onFavoriteToggle}`
- `disabled={isFavoritePending}`
- When favorited: filled `Heart` with `fill-current text-rose-500` + text "已收藏"
- When not favorited: outline `Heart` + text "收藏"

```typescript
<button
  type="button"
  onClick={onFavoriteToggle}
  disabled={isFavoritePending}
  className={`flex items-center gap-2 rounded-[10px] border px-5 py-3 text-sm font-medium transition-colors ${
    isFavorited
      ? "border-rose-200 bg-rose-50 text-rose-500 hover:bg-rose-100"
      : "border-slate-200 bg-white text-slate-500 hover:bg-slate-50"
  } disabled:opacity-50`}
>
  <Heart
    className={`h-4 w-4 ${isFavorited ? "fill-current" : ""}`}
  />
  {isFavorited ? "已收藏" : "收藏"}
</button>
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: add favorite toggle to hero card
```

---

### Task 7: Wire into GameDetailContent

**Files:**
- Modify: `src/features/web/games/components/game-detail-content.tsx:10-68`

**Step 1: Add isFavorited prop and hook**

Add `isFavorited: boolean` to `GameDetailContentProps`.

Import and call `useFavoriteToggle`:

```typescript
import { useFavoriteToggle } from "@/features/web/games/hooks/use-favorite-toggle";
```

Inside the component:

```typescript
const { favorited, toggle, isPending: isFavoritePending } =
  useFavoriteToggle(game.id, game.name, isFavorited);
```

Pass to `HeroCard`:

```typescript
<HeroCard
  {...existingProps}
  isFavorited={favorited}
  onFavoriteToggle={toggle}
  isFavoritePending={isFavoritePending}
/>
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: wire favorite toggle into game detail content
```

---

### Task 8: Wire into Game Detail Page (Server)

**Files:**
- Modify: `src/app/(web)/hall/(main)/games/[id]/page.tsx:1-133`

**Step 1: Add favorite query to server fetch**

Import:

```typescript
import { isGameFavorited } from "@/models/user-favorite/user-favorite.query";
```

Add to the existing `Promise.all` (line 79-84):

```typescript
const [activeSession, myStats, isFavorited] = userId
  ? await Promise.all([
      getAnyActiveSession(userId, game.id),
      getGameStats(userId, game.id),
      isGameFavorited(userId, game.id),
    ])
  : [null, null, false];
```

Pass to `GameDetailContent`:

```typescript
<GameDetailContent
  game={...}
  heroSession={heroSession}
  isFavorited={isFavorited}
  rules={...}
  stats={...}
/>
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: fetch and pass favorite state in game detail page
```

---

### Task 9: Favorite Card Component

**Files:**
- Create: `src/features/web/hall/components/favorite-card.tsx`

**Step 1: Create the component**

```typescript
import Link from "next/link";
import { Play } from "lucide-react";
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";
import type { FavoriteGameCard } from "@/models/user-favorite/user-favorite.query";

const GRADIENT_COVERS = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

function getGradient(id: string) {
  let hash = 0;
  for (const ch of id) hash = (hash * 31 + ch.charCodeAt(0)) | 0;
  return GRADIENT_COVERS[Math.abs(hash) % GRADIENT_COVERS.length];
}

export function FavoriteCard({ game }: { game: FavoriteGameCard }) {
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
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-[11px] font-medium text-slate-500">
              {game.category?.name ?? modeLabel}
            </span>
          </div>
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

Run: `npx tsc --noEmit`
Expected: No errors.

**Step 3: Commit**

```
feat: add favorite card component
```

---

### Task 10: Favorites Page — Real Data

**Files:**
- Modify: `src/app/(web)/hall/(main)/favorites/page.tsx:1-112`

**Step 1: Rewrite with real data**

```typescript
import { Heart } from "lucide-react";

import { auth } from "@/lib/auth";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { FavoriteCard } from "@/features/web/hall/components/favorite-card";
import { getUserFavorites } from "@/models/user-favorite/user-favorite.query";

export default async function FavoritesPage() {
  const session = await auth();
  const userId = session?.user?.id;

  const favorites = userId ? await getUserFavorites(userId) : [];

  return (
    <div className="flex h-full flex-col gap-6 px-8 py-7">
      <PageTopBar
        title="我的收藏"
        subtitle="收藏你喜欢的课程游戏和学习内容"
        searchPlaceholder="搜索收藏..."
      />

      {/* Filter row */}
      <div className="flex items-center justify-between">
        <span className="rounded-full bg-teal-600 px-5 py-2 text-[13px] font-semibold text-white">
          全部 ({favorites.length})
        </span>
        <span className="text-[13px] text-slate-400">
          共 {favorites.length} 个收藏
        </span>
      </div>

      <div className="h-px w-full bg-slate-200" />

      {favorites.length > 0 ? (
        <div className="grid grid-cols-5 gap-4">
          {favorites.map((game) => (
            <FavoriteCard key={game.id} game={game} />
          ))}
        </div>
      ) : (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 text-slate-400">
          <Heart className="h-12 w-12 stroke-1" />
          <p className="text-sm">还没有收藏，去发现喜欢的游戏吧</p>
        </div>
      )}
    </div>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`
Expected: Build succeeds with no errors.

**Step 3: Commit**

```
feat: wire favorites page to real data
```

---

### Task 11: Final Verification

**Step 1: Full build check**

Run: `npm run build`
Expected: Build succeeds.

**Step 2: Lint check**

Run: `npm run lint`
Expected: No lint errors.

**Step 3: Manual smoke test**

1. Navigate to `/hall/games/[id]` — verify 收藏 button shows outline heart
2. Click 收藏 — toast shows "已收藏「游戏名」", heart fills rose, text changes to "已收藏"
3. Click 已收藏 — toast shows "已取消收藏「游戏名」", heart reverts
4. Rapid click within 2s — toast shows "操作频繁，请稍后再试"
5. Navigate to `/hall/favorites` — favorited games appear with real data
6. Remove all favorites — empty state shows "还没有收藏，去发现喜欢的游戏吧"
