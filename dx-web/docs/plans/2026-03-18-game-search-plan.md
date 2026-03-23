# Game Fuzzy Search Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Cmd+K game fuzzy search to the hall top bar using ShadCN Command dialog with debounced server-side search and recently played suggestions.

**Architecture:** Two model queries (search + recent), two server actions, one custom hook for debounce/state, and two client components (trigger + dialog). TopActions stays a server component.

**Tech Stack:** ShadCN Command (cmdk), Next.js server actions, Prisma `contains` search, React hooks

**Design doc:** `docs/plans/2026-03-18-game-search-design.md`

---

### Task 1: Add `GameSearchResult` type and `searchPublishedGames` query

**Files:**
- Modify: `src/models/game/game.query.ts`

**Step 1: Add the type and query function at the end of the file**

```ts
export type GameSearchResult = {
  id: string;
  name: string;
  mode: string;
  category: { name: string } | null;
};

/** Search published games by name (case-insensitive) */
export async function searchPublishedGames(
  query: string,
  limit: number = 8
): Promise<GameSearchResult[]> {
  return db.game.findMany({
    where: {
      status: "published",
      isActive: true,
      name: { contains: query, mode: "insensitive" },
    },
    select: {
      id: true,
      name: true,
      mode: true,
      category: { select: { name: true } },
    },
    orderBy: { createdAt: "desc" },
    take: limit,
  });
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: no errors related to game.query.ts

**Step 3: Commit**

```
feat: add searchPublishedGames query for game fuzzy search
```

---

### Task 2: Add `getRecentlyPlayedGames` query

**Files:**
- Modify: `src/models/game-stats-total/game-stats-total.query.ts`

**Step 1: Import the `GameSearchResult` type and add query**

Add import at top:
```ts
import type { GameSearchResult } from "@/models/game/game.query";
```

Add function at end of file:
```ts
/** Get user's recently played games for search suggestions */
export async function getRecentlyPlayedGames(
  userId: string,
  limit: number = 5
): Promise<GameSearchResult[]> {
  const stats = await db.gameStatsTotal.findMany({
    where: {
      userId,
      game: { status: "published", isActive: true },
    },
    select: {
      game: {
        select: {
          id: true,
          name: true,
          mode: true,
          category: { select: { name: true } },
        },
      },
    },
    orderBy: { lastPlayedAt: "desc" },
    take: limit,
  });

  return stats.map((s) => s.game);
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: no errors

**Step 3: Commit**

```
feat: add getRecentlyPlayedGames query for search suggestions
```

---

### Task 3: Create server actions

**Files:**
- Create: `src/features/web/hall/actions/game-search.action.ts`

**Step 1: Create the server actions file**

```ts
"use server";

import { auth } from "@/lib/auth";
import {
  searchPublishedGames,
  type GameSearchResult,
} from "@/models/game/game.query";
import { getRecentlyPlayedGames } from "@/models/game-stats-total/game-stats-total.query";

export type GameSearchActionResult = {
  games: GameSearchResult[];
  error?: string;
};

/** Search published games by name */
export async function searchGamesAction(
  query: string
): Promise<GameSearchActionResult> {
  try {
    const trimmed = query.trim();
    if (!trimmed) return { games: [] };

    const games = await searchPublishedGames(trimmed, 8);
    return { games };
  } catch {
    return { games: [], error: "搜索失败，请重试" };
  }
}

/** Get current user's recently played games */
export async function getRecentGamesAction(): Promise<GameSearchActionResult> {
  try {
    const session = await auth();
    if (!session?.user?.id) return { games: [] };

    const games = await getRecentlyPlayedGames(session.user.id, 5);
    return { games };
  } catch {
    return { games: [], error: "加载失败，请重试" };
  }
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: no errors

**Step 3: Commit**

```
feat: add game search server actions
```

---

### Task 4: Create `useGameSearch` hook

**Files:**
- Create: `src/features/web/hall/hooks/use-game-search.ts`

**Step 1: Create the hook**

```ts
"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import type { GameSearchResult } from "@/models/game/game.query";
import {
  searchGamesAction,
  getRecentGamesAction,
} from "@/features/web/hall/actions/game-search.action";

/** Manages game search dialog state with debounced server queries */
export function useGameSearch() {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<GameSearchResult[]>([]);
  const [recentGames, setRecentGames] = useState<GameSearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const recentLoadedRef = useRef(false);

  /** Load recent games once when dialog opens */
  useEffect(() => {
    if (!isOpen) return;
    if (recentLoadedRef.current) return;

    recentLoadedRef.current = true;
    getRecentGamesAction().then((res) => {
      if (!res.error) setRecentGames(res.games);
    });
  }, [isOpen]);

  /** Reset query when dialog closes */
  useEffect(() => {
    if (!isOpen) {
      setQuery("");
      setResults([]);
    }
  }, [isOpen]);

  /** Debounced search on query change */
  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current);

    const trimmed = query.trim();
    if (!trimmed) {
      setResults([]);
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    timerRef.current = setTimeout(async () => {
      const res = await searchGamesAction(trimmed);
      setResults(res.games);
      setIsLoading(false);
    }, 300);

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [query]);

  /** Register Cmd+K / Ctrl+K keyboard shortcut */
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setIsOpen((prev) => !prev);
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  /** The items to display: search results when typing, recent games otherwise */
  const displayItems = query.trim() ? results : recentGames;
  const groupLabel = query.trim() ? "搜索结果" : "最近玩过";
  const showGroup = displayItems.length > 0;

  return {
    isOpen,
    setIsOpen,
    query,
    setQuery,
    displayItems,
    groupLabel,
    showGroup,
    isLoading,
  };
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: no errors

**Step 3: Commit**

```
feat: add useGameSearch hook with debounced search
```

---

### Task 5: Create `GameSearchDialog` component

**Files:**
- Create: `src/features/web/hall/components/game-search-dialog.tsx`

**Step 1: Create the component**

```tsx
"use client";

import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";
import {
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from "@/components/ui/command";
import { GAME_MODE_LABELS } from "@/consts/game-mode";
import { useGameSearch } from "@/features/web/hall/hooks/use-game-search";

/** Cmd+K game search dialog with fuzzy matching and recent suggestions */
export function GameSearchDialog() {
  const router = useRouter();
  const {
    isOpen,
    setIsOpen,
    query,
    setQuery,
    displayItems,
    groupLabel,
    showGroup,
    isLoading,
  } = useGameSearch();

  /** Navigate to game detail and close dialog */
  function handleSelect(gameId: string) {
    router.push(`/hall/games/${gameId}`);
    setIsOpen(false);
  }

  /** Format the mode label for display */
  function getModeLabel(mode: string): string {
    return GAME_MODE_LABELS[mode as keyof typeof GAME_MODE_LABELS] ?? mode;
  }

  return (
    <CommandDialog
      open={isOpen}
      onOpenChange={setIsOpen}
      title="搜索课程游戏"
      description="输入课程名称搜索"
      showCloseButton={false}
    >
      <CommandInput
        placeholder="输入课程名称搜索..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList>
        {isLoading && (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
          </div>
        )}

        {!isLoading && query.trim() && displayItems.length === 0 && (
          <CommandEmpty>没有找到相关课程</CommandEmpty>
        )}

        {!isLoading && showGroup && (
          <CommandGroup heading={groupLabel}>
            {displayItems.map((game) => (
              <CommandItem
                key={game.id}
                value={game.name}
                onSelect={() => handleSelect(game.id)}
                className="cursor-pointer"
              >
                <div className="flex flex-col gap-0.5">
                  <span className="text-sm font-medium">{game.name}</span>
                  <span className="text-xs text-muted-foreground">
                    {getModeLabel(game.mode)}
                    {game.category && ` · ${game.category.name}`}
                  </span>
                </div>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
}
```

**Step 2: Verify build**

Run: `npx tsc --noEmit`
Expected: no errors

**Step 3: Commit**

```
feat: add GameSearchDialog component with Command palette
```

---

### Task 6: Create `GameSearchTrigger` and update `TopActions`

**Files:**
- Create: `src/features/web/hall/components/game-search-trigger.tsx`
- Modify: `src/features/web/hall/components/top-actions.tsx`

**Step 1: Create the trigger component**

```tsx
"use client";

import { Search } from "lucide-react";
import { GameSearchDialog } from "@/features/web/hall/components/game-search-dialog";

/** Clickable search bar that opens the game search dialog */
export function GameSearchTrigger({
  placeholder = "搜索课程游戏...",
}: {
  placeholder?: string;
}) {
  return (
    <>
      <GameSearchDialog />
      <button
        type="button"
        onClick={() => {
          document.dispatchEvent(
            new KeyboardEvent("keydown", { key: "k", metaKey: true })
          );
        }}
        className="flex h-10 w-80 items-center gap-2 rounded-[10px] border border-slate-200 bg-white px-3 hover:bg-slate-50"
      >
        <Search className="h-4 w-4 text-slate-400" />
        <span className="flex-1 text-left text-[13px] text-slate-400">
          {placeholder}
        </span>
        <kbd className="pointer-events-none hidden h-5 items-center gap-0.5 rounded border border-slate-200 bg-slate-100 px-1.5 font-mono text-[10px] font-medium text-slate-400 sm:inline-flex">
          ⌘K
        </kbd>
      </button>
    </>
  );
}
```

**Step 2: Update `TopActions` — replace static search div with trigger**

In `src/features/web/hall/components/top-actions.tsx`:

Remove the `Search` import from lucide-react (line 2), add the trigger import, and replace the static search div (lines 22-25) with `<GameSearchTrigger searchPlaceholder={searchPlaceholder} />`. Pass the prop through.

Updated file:
```tsx
import Link from "next/link";
import { Bell, Sun } from "lucide-react";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { UserProfileMenu } from "@/features/web/auth/components/user-profile-menu";
import { hasUnreadNotices } from "@/features/web/hall/helpers/has-unread-notices";
import { ContentSeekButton } from "@/features/web/hall/components/content-seek-button";
import { FeedbackButton } from "@/features/web/hall/components/feedback-button";
import { GameSearchTrigger } from "@/features/web/hall/components/game-search-trigger";

export async function TopActions({
  searchPlaceholder = "搜索课程游戏...",
}: {
  searchPlaceholder?: string;
} = {}) {
  const [profile, unread] = await Promise.all([
    fetchUserProfile(),
    hasUnreadNotices(),
  ]);

  return (
    <div className="flex items-center gap-3">
      {/* Search */}
      <GameSearchTrigger placeholder={searchPlaceholder} />

      {/* Request course */}
      <ContentSeekButton />

      {/* Icon buttons */}
      <FeedbackButton />
      <Link
        href="/hall/notices"
        className="relative flex h-10 w-10 items-center justify-center rounded-[10px] border border-slate-200 bg-white text-slate-500 hover:bg-slate-50"
      >
        <Bell className="h-[18px] w-[18px]" />
        {unread && (
          <span className="absolute top-2 right-2 h-2 w-2 rounded-full bg-red-500" />
        )}
      </Link>
      <button className="flex h-10 w-10 items-center justify-center rounded-[10px] border border-slate-200 bg-white text-slate-500 hover:bg-slate-50">
        <Sun className="h-[18px] w-[18px]" />
      </button>

      {/* Profile */}
      {profile && <UserProfileMenu profile={profile} />}
    </div>
  );
}
```

**Step 3: Verify build**

Run: `npx tsc --noEmit`
Expected: no errors

**Step 4: Manual test**

1. Open `http://localhost:3000/hall`
2. Click the search bar → dialog opens
3. Press `⌘K` → dialog toggles
4. Type a game name → suggestions appear after 300ms
5. Clear input → recent games show (if logged in)
6. Click a suggestion → navigates to `/hall/games/{id}`
7. Press Escape → dialog closes

**Step 5: Commit**

```
feat: wire game search trigger into TopActions with Cmd+K support
```
