# Search Modal Text Filter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a "search text" top item to the Cmd+K modal that filters the games list page by name, and show active search text in the top bar trigger with a clear button.

**Architecture:** Zustand store shares the search text between the search modal (in the top bar) and the games page content. Backend adds a `q` ILIKE filter to the existing `/api/games` list endpoint. cmdk controlled `value` keeps the top search item default-selected.

**Tech Stack:** Next.js 16, React 19, cmdk v1, zustand, shadcn/ui, Go/Goravel, PostgreSQL

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `dx-web/src/features/web/games/stores/game-search-store.ts` | Zustand store: `q`, `setQ`, `clearQ` |
| Modify | `dx-api/app/http/requests/api/game_request.go:14-20` | Add `Q` field to `ListGamesRequest` |
| Modify | `dx-api/app/http/controllers/api/game_controller.go:32` | Pass `req.Q` to service |
| Modify | `dx-api/app/services/api/game_service.go:66-84` | Add `q` param with ILIKE filter |
| Modify | `dx-web/src/features/web/games/hooks/use-infinite-public-games.ts:8-10,24` | Add `q` to Filters and URL params |
| Modify | `dx-web/src/features/web/games/components/games-page-content.tsx:12,23-24` | Read `q` from store, pass to hook |
| Modify | `dx-web/src/features/web/hall/components/game-search-dialog.tsx` | Add top search item, controlled selection, new group heading |
| Modify | `dx-web/src/features/web/hall/hooks/use-game-search.ts:84-86` | Change group label to "猜你想学" |
| Modify | `dx-web/src/features/web/hall/components/game-search-trigger.tsx` | Show active search text with clear button |

---

### Task 1: Add `q` parameter to Go backend

**Files:**
- Modify: `dx-api/app/http/requests/api/game_request.go:14-20,39-50`
- Modify: `dx-api/app/http/controllers/api/game_controller.go:32`
- Modify: `dx-api/app/services/api/game_service.go:66-84`

- [ ] **Step 1: Add `Q` field to `ListGamesRequest`**

In `dx-api/app/http/requests/api/game_request.go`, add `Q` to the struct and update `Rules` and `Filters`:

```go
// ListGamesRequest holds query parameters for listing published games.
type ListGamesRequest struct {
	Cursor      string `form:"cursor" json:"cursor"`
	Limit       int    `form:"limit" json:"limit"`
	CategoryIDs string `form:"categoryIds" json:"categoryIds"`
	PressID     string `form:"pressId" json:"pressId"`
	Mode        string `form:"mode" json:"mode"`
	Q           string `form:"q" json:"q"`
}
```

Update `Rules` to add validation for `q`:

```go
func (r *ListGamesRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"pressId": "uuid",
		"mode":    helpers.InEnum("mode"),
		"q":       "max_len:50",
	}
}
```

Update `Filters` to trim `q`:

```go
func (r *ListGamesRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"cursor":  "trim",
		"pressId": "trim",
		"q":       "trim",
	}
}
```

Add message for `q`:

```go
func (r *ListGamesRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"limit.min":    "每页数量不能小于1",
		"limit.max":    "每页数量不能超过50",
		"pressId.uuid": "无效的出版社ID",
		"mode.in":      "无效的游戏模式",
		"q.max_len":    "搜索关键词不能超过50个字符",
	}
}
```

- [ ] **Step 2: Add `q` parameter to `ListPublishedGames` service function**

In `dx-api/app/services/api/game_service.go`, change the signature and add the ILIKE filter:

```go
func ListPublishedGames(cursor string, limit int, categoryIDs []string, pressID string, mode string, q string) ([]GameCardData, string, bool, error) {
	if limit <= 0 {
		limit = 12
	}

	query := facades.Orm().Query().
		Where("status", consts.GameStatusPublished).
		Where("is_active", true).
		Where("is_private", false)

	if len(categoryIDs) > 0 {
		query = query.Where("game_category_id IN ?", categoryIDs)
	}
	if pressID != "" {
		query = query.Where("game_press_id", pressID)
	}
	if mode != "" {
		query = query.Where("mode", mode)
	}
	if q != "" {
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}

	// ... rest unchanged
```

- [ ] **Step 3: Pass `req.Q` in controller**

In `dx-api/app/http/controllers/api/game_controller.go`, update the call on line 32:

```go
games, nextCursor, hasMore, err := services.ListPublishedGames(req.Cursor, limit, req.ParseCategoryIDs(), req.PressID, req.Mode, req.Q)
```

- [ ] **Step 4: Verify backend compiles**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/http/requests/api/game_request.go dx-api/app/http/controllers/api/game_controller.go dx-api/app/services/api/game_service.go
git commit -m "feat(api): add q text filter to game list endpoint"
```

---

### Task 2: Create zustand store for search text

**Files:**
- Create: `dx-web/src/features/web/games/stores/game-search-store.ts`

- [ ] **Step 1: Create the stores directory and store file**

Create `dx-web/src/features/web/games/stores/game-search-store.ts`:

```ts
import { create } from "zustand";

type GameSearchTextStore = {
  q: string;
  setQ: (q: string) => void;
  clearQ: () => void;
};

export const useGameSearchText = create<GameSearchTextStore>((set) => ({
  q: "",
  setQ: (q) => set({ q }),
  clearQ: () => set({ q: "" }),
}));
```

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/games/stores/game-search-store.ts
git commit -m "feat(web): add zustand store for game search text"
```

---

### Task 3: Wire `q` filter into games list data fetching

**Files:**
- Modify: `dx-web/src/features/web/games/hooks/use-infinite-public-games.ts:8-10,24`
- Modify: `dx-web/src/features/web/games/components/games-page-content.tsx:12,23-24`

- [ ] **Step 1: Add `q` to `Filters` type and URL params in `use-infinite-public-games.ts`**

In `dx-web/src/features/web/games/hooks/use-infinite-public-games.ts`, add `q` to the type and include it in params:

```ts
type Filters = {
  categoryIds?: string[]
  pressId?: string
  mode?: string
  q?: string
}
```

Inside `getKey`, add after the `mode` param line:

```ts
if (filters.q) params.set("q", filters.q)
```

- [ ] **Step 2: Read `q` from store in `games-page-content.tsx`**

In `dx-web/src/features/web/games/components/games-page-content.tsx`, import the store and pass `q` to the hook:

Add import:

```ts
import { useGameSearchText } from "@/features/web/games/stores/game-search-store"
```

Inside the `GamesPageContent` function, read `q` from store and include in the hook call:

```ts
export function GamesPageContent({ categories, presses }: GamesPageContentProps) {
  const [filters, setFilters] = useState<Filters>({})
  const q = useGameSearchText((s) => s.q)
  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfinitePublicGames({ ...filters, q })
```

- [ ] **Step 3: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/games/hooks/use-infinite-public-games.ts dx-web/src/features/web/games/components/games-page-content.tsx
git commit -m "feat(web): wire q filter into games list data fetching"
```

---

### Task 4: Add top search item to search dialog

**Files:**
- Modify: `dx-web/src/features/web/hall/components/game-search-dialog.tsx`
- Modify: `dx-web/src/features/web/hall/hooks/use-game-search.ts:85`

- [ ] **Step 1: Change group label in `use-game-search.ts`**

In `dx-web/src/features/web/hall/hooks/use-game-search.ts`, change line 85:

```ts
const groupLabel = query.trim() ? "猜你想学" : "最近玩过";
```

- [ ] **Step 2: Rewrite `game-search-dialog.tsx` with top search item and controlled selection**

Replace the full content of `dx-web/src/features/web/hall/components/game-search-dialog.tsx`:

```tsx
"use client";

import { useState, useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { Loader2, Search } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from "@/components/ui/command";
import { GAME_MODE_LABELS } from "@/consts/game-mode";
import { useGameSearch } from "@/features/web/hall/hooks/use-game-search";
import { useGameSearchText } from "@/features/web/games/stores/game-search-store";

const SEARCH_ITEM_VALUE = "__search__";

/** Cmd+K game search dialog with server-side fuzzy matching and recent suggestions */
export function GameSearchDialog() {
  const router = useRouter();
  const pathname = usePathname();
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
  const setQ = useGameSearchText((s) => s.setQ);
  const [selectedValue, setSelectedValue] = useState("");

  const trimmedQuery = query.trim();

  /** Reset selection to top search item whenever query changes */
  useEffect(() => {
    if (trimmedQuery) {
      setSelectedValue(SEARCH_ITEM_VALUE);
    } else {
      setSelectedValue("");
    }
  }, [trimmedQuery]);

  /** Navigate to games list with text filter */
  function handleSearchSelect() {
    if (!trimmedQuery) return;
    setQ(trimmedQuery);
    if (pathname !== "/hall/games") {
      router.push("/hall/games");
    }
    setIsOpen(false);
  }

  /** Navigate to game detail and close dialog */
  function handleGameSelect(gameId: string) {
    router.push(`/hall/games/${gameId}`);
    setIsOpen(false);
  }

  /** Format the mode label for display */
  function getModeLabel(mode: string): string {
    return GAME_MODE_LABELS[mode as keyof typeof GAME_MODE_LABELS] ?? mode;
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogHeader className="sr-only">
        <DialogTitle>搜索课程游戏</DialogTitle>
        <DialogDescription>输入课程游戏名称搜索</DialogDescription>
      </DialogHeader>
      <DialogContent className="top-[10%] translate-y-0 overflow-hidden p-2" showCloseButton={false}>
        <Command
          shouldFilter={false}
          value={selectedValue}
          onValueChange={setSelectedValue}
          className="[&_[cmdk-group-heading]]:text-muted-foreground **:data-[slot=command-input-wrapper]:h-12 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group]]:px-2 [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:h-5 [&_[cmdk-item]_svg]:w-5"
        >
          <CommandInput
            placeholder="输入课程游戏名称搜索..."
            value={query}
            onValueChange={setQuery}
          />
          <CommandList>
            {trimmedQuery && (
              <CommandGroup>
                <CommandItem
                  value={SEARCH_ITEM_VALUE}
                  onSelect={handleSearchSelect}
                  className="cursor-pointer"
                >
                  <Search className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm">
                    搜索 &ldquo;{trimmedQuery}&rdquo;
                  </span>
                </CommandItem>
              </CommandGroup>
            )}

            {isLoading && (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            )}

            {!isLoading && displayItems.length === 0 && !trimmedQuery && (
              <CommandEmpty>暂无数据</CommandEmpty>
            )}

            {!isLoading && showGroup && (
              <CommandGroup heading={groupLabel}>
                {displayItems.map((game) => (
                  <CommandItem
                    key={game.id}
                    value={game.id}
                    onSelect={() => handleGameSelect(game.id)}
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
        </Command>
      </DialogContent>
    </Dialog>
  );
}
```

Key changes from original:
- Import `useState`, `useEffect`, `usePathname`, `Search` icon, `useGameSearchText`
- `SEARCH_ITEM_VALUE` constant for the top item
- Controlled `value`/`onValueChange` on `Command`
- `useEffect` resets selection to `SEARCH_ITEM_VALUE` when query changes
- New `CommandGroup` at top with the search item (only when `trimmedQuery` is non-empty)
- `handleSearchSelect`: calls `setQ`, navigates to `/hall/games` if not there, closes modal
- Game items use `game.id` as `value` (was `game.name`) to avoid cmdk value collisions
- Empty state only shows when no query and no items (removed "未找到相关数据" since the top search item is always available when typing)

- [ ] **Step 3: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/hall/components/game-search-dialog.tsx dx-web/src/features/web/hall/hooks/use-game-search.ts
git commit -m "feat(web): add top search item to game search dialog"
```

---

### Task 5: Show active search text in trigger with clear button

**Files:**
- Modify: `dx-web/src/features/web/hall/components/game-search-trigger.tsx`

- [ ] **Step 1: Rewrite `game-search-trigger.tsx`**

Replace the full content of `dx-web/src/features/web/hall/components/game-search-trigger.tsx`:

```tsx
"use client";

import { usePathname } from "next/navigation";
import { Search, X } from "lucide-react";
import { GameSearchDialog } from "@/features/web/hall/components/game-search-dialog";
import { useGameSearchText } from "@/features/web/games/stores/game-search-store";
import { KbdGroup } from "@/components/ui/kbd";

/** Clickable search bar that opens the game search dialog */
export function GameSearchTrigger({
  placeholder = "搜索课程游戏...",
}: {
  placeholder?: string;
}) {
  const pathname = usePathname();
  const q = useGameSearchText((s) => s.q);
  const clearQ = useGameSearchText((s) => s.clearQ);

  const showActiveSearch = q && pathname === "/hall/games";

  function openDialog() {
    document.dispatchEvent(
      new KeyboardEvent("keydown", { key: "k", metaKey: true })
    );
  }

  return (
    <>
      <GameSearchDialog />
      <button
        type="button"
        onClick={openDialog}
        className="flex h-10 w-52 items-center gap-2 rounded-[10px] border border-border bg-card px-3 hover:bg-accent"
      >
        <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
        {showActiveSearch ? (
          <>
            <span className="flex-1 truncate text-left text-[13px] text-foreground">
              {q}
            </span>
            <button
              type="button"
              aria-label="清除搜索"
              onClick={(e) => {
                e.stopPropagation();
                clearQ();
              }}
              className="flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground hover:text-foreground"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </>
        ) : (
          <>
            <span className="flex-1 text-left text-[13px] text-muted-foreground">
              {placeholder}
            </span>
            <KbdGroup>
              <kbd className="pointer-events-none hidden h-5 items-center gap-0.5 rounded border border-border bg-muted px-1.5 font-mono text-muted-foreground sm:inline-flex">
                ⌘
              </kbd>
              <kbd className="pointer-events-none hidden h-5 items-center gap-0.5 rounded border border-border bg-muted px-1.5 text-[10px] font-mono text-muted-foreground sm:inline-flex">
                k
              </kbd>
            </KbdGroup>
          </>
        )}
      </button>
    </>
  );
}
```

Key changes from original:
- Import `usePathname`, `X` icon, `useGameSearchText`
- Read `q` and `clearQ` from store
- `showActiveSearch` is true when `q` is set AND on `/hall/games`
- When active: show search text (truncated) + (X) clear button with `stopPropagation`
- When inactive: show default placeholder + keyboard shortcut (unchanged)

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npx tsc --noEmit`
Expected: No type errors.

- [ ] **Step 3: Lint check**

Run: `cd dx-web && npm run lint`
Expected: No lint errors.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/hall/components/game-search-trigger.tsx
git commit -m "feat(web): show active search text in trigger with clear button"
```

---

### Task 6: Manual verification

- [ ] **Step 1: Start backend and frontend**

Run in separate terminals:
```bash
cd dx-api && air
cd dx-web && npm run dev
```

- [ ] **Step 2: Test top search item behavior**

1. Open http://localhost:3000/hall/games
2. Press Cmd+K — dialog opens, shows "最近玩过" games
3. Type "word" — top item appears: `搜索 "word"` (highlighted by default), game results below under "猜你想学"
4. Press Enter — modal closes, games list filters by "word", trigger shows "word" with (X)
5. Click (X) on trigger — filter clears, trigger reverts to "搜索课程游戏...", games list resets

- [ ] **Step 3: Test game item selection**

1. Press Cmd+K, type a query that returns results
2. Press arrow-down to highlight a game item
3. Press Enter — modal closes, navigates to `/hall/games/{id}`
4. Verify game detail page loads correctly

- [ ] **Step 4: Test combined filters**

1. On `/hall/games`, select a category filter (e.g., "同步练习")
2. Press Cmd+K, type "word", press Enter on top item
3. Verify games list shows results matching BOTH the category filter AND text "word"
4. Click (X) to clear text — category filter should remain active

- [ ] **Step 5: Test cross-page navigation**

1. Navigate to `/hall` (dashboard)
2. Press Cmd+K, type "word", press Enter on top item
3. Verify navigation to `/hall/games` with filtered results
4. Navigate away to `/hall/favorites` — trigger shows default placeholder
5. Navigate back to `/hall/games` — filter is still active, trigger shows "word" with (X)

- [ ] **Step 6: Test edge cases**

1. Press Cmd+K, type spaces only — top search item should NOT appear
2. Press Cmd+K, type query with no results — top search item still appears, no game items below
3. Press Escape — dialog closes without changes
4. Press Cmd+K when search is active — dialog opens with empty input (fresh search)
