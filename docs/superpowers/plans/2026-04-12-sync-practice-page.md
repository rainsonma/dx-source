# Sync Practice Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a dedicated 同步练习 page (`/hall/sync`) to the hall, extracting the 同步练习 category subtree from the existing games page into its own page with category/press/mode filters and infinite-scroll game listing.

**Architecture:** Backend splits the category service into main (excluding sync) and sync (only sync subtree with adjusted depths). A new `GET /api/game-categories/sync` endpoint serves the sync subtree. Frontend creates a new sync feature with a page component that reuses the existing `FilterSection`, `GameCard`, and `useInfinitePublicGames` — the only difference is a `useMemo` that ensures the API always receives sync category IDs even when no filter is selected.

**Tech Stack:** Go/Goravel (backend), Next.js/React/SWR/Lucide (frontend)

**Spec:** `docs/superpowers/specs/2026-04-12-sync-practice-page-design.md`

---

### Task 1: Backend — Category Service Refactor

**Files:**
- Modify: `dx-api/app/services/api/game_category_service.go`

- [ ] **Step 1: Replace the full file with the refactored version**

The current file has a single `ListCategories()` function. Replace the entire file with:
- A `SyncCategoryName` constant
- A shared `categoryTree` struct and `loadCategoryTree()` helper
- A modified `ListCategories()` that skips the sync subtree
- A new `ListSyncCategories()` that returns only the sync subtree with adjusted depths

```go
package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// SyncCategoryName is the name of the top-level category shown on /hall/sync.
const SyncCategoryName = "同步练习"

// CategoryData represents a game category with hierarchy info.
type CategoryData struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Depth  int    `json:"depth"`
	IsLeaf bool   `json:"isLeaf"`
}

// categoryTree holds the loaded category hierarchy for shared use.
type categoryTree struct {
	parentMap   map[string][]models.GameCategory
	hasChildren map[string]bool
	syncID      string
}

// loadCategoryTree loads all enabled categories and builds the tree structure.
func loadCategoryTree() (*categoryTree, error) {
	var categories []models.GameCategory
	if err := facades.Orm().Query().
		Where("is_enabled", true).
		Order("\"order\" ASC").
		Get(&categories); err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	tree := &categoryTree{
		parentMap:   make(map[string][]models.GameCategory),
		hasChildren: make(map[string]bool),
	}

	for _, cat := range categories {
		key := ""
		if cat.ParentID != nil {
			key = *cat.ParentID
			tree.hasChildren[*cat.ParentID] = true
		}
		tree.parentMap[key] = append(tree.parentMap[key], cat)

		if cat.Name == SyncCategoryName && cat.ParentID == nil {
			tree.syncID = cat.ID
		}
	}

	return tree, nil
}

// ListCategories returns all enabled categories except the sync subtree.
func ListCategories() ([]CategoryData, error) {
	tree, err := loadCategoryTree()
	if err != nil {
		return nil, err
	}

	var result []CategoryData
	var walk func(parentID string, depth int)
	walk = func(parentID string, depth int) {
		for _, cat := range tree.parentMap[parentID] {
			if cat.ID == tree.syncID {
				continue
			}
			result = append(result, CategoryData{
				ID:     cat.ID,
				Name:   cat.Name,
				Depth:  depth,
				IsLeaf: !tree.hasChildren[cat.ID],
			})
			walk(cat.ID, depth+1)
		}
	}
	walk("", 0)

	return result, nil
}

// ListSyncCategories returns the sync subtree with depths adjusted to start at 0.
func ListSyncCategories() ([]CategoryData, error) {
	tree, err := loadCategoryTree()
	if err != nil {
		return nil, err
	}

	if tree.syncID == "" {
		return []CategoryData{}, nil
	}

	var result []CategoryData
	var walk func(parentID string, depth int)
	walk = func(parentID string, depth int) {
		for _, cat := range tree.parentMap[parentID] {
			result = append(result, CategoryData{
				ID:     cat.ID,
				Name:   cat.Name,
				Depth:  depth,
				IsLeaf: !tree.hasChildren[cat.ID],
			})
			walk(cat.ID, depth+1)
		}
	}
	walk(tree.syncID, 0)

	return result, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 3: Run go vet**

Run: `cd dx-api && go vet ./...`
Expected: no issues

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/game_category_service.go
git commit -m "refactor(api): split category service into main and sync"
```

---

### Task 2: Backend — Controller, Route, Menu Item

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_category_controller.go`
- Modify: `dx-api/routes/api.go`
- Modify: `dx-api/app/consts/hall_menu.go`

- [ ] **Step 1: Add SyncCategories method to controller**

In `dx-api/app/http/controllers/api/game_category_controller.go`, add after the existing `Categories` method (after line 27):

```go

// SyncCategories returns the sync subtree categories.
func (c *GameCategoryController) SyncCategories(ctx contractshttp.Context) contractshttp.Response {
	categories, err := services.ListSyncCategories()
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list sync categories")
	}

	return helpers.Success(ctx, categories)
}
```

- [ ] **Step 2: Add route**

In `dx-api/routes/api.go`, find the category route (line 54):

```go
		gameCategoryController := apicontrollers.NewGameCategoryController()
		router.Get("/game-categories", gameCategoryController.Categories)
```

Replace with:

```go
		gameCategoryController := apicontrollers.NewGameCategoryController()
		router.Get("/game-categories", gameCategoryController.Categories)
		router.Get("/game-categories/sync", gameCategoryController.SyncCategories)
```

- [ ] **Step 3: Add hall menu item**

In `dx-api/app/consts/hall_menu.go`, find the 学习课程 entry (line 21):

```go
			{Icon: "Gamepad2", Label: "学习课程", Subtitle: "选择一个游戏模式，边玩边学英语！", Href: "/hall/games"},
```

Add after it:

```go
			{Icon: "BookMarked", Label: "同步练习", Subtitle: "同步参考练习，巩固学习内容", Href: "/hall/sync"},
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 5: Run go vet**

Run: `cd dx-api && go vet ./...`
Expected: no issues

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/http/controllers/api/game_category_controller.go dx-api/routes/api.go dx-api/app/consts/hall_menu.go
git commit -m "feat(api): add sync categories endpoint and menu item"
```

---

### Task 3: Frontend — Icon Registry

**Files:**
- Modify: `dx-web/src/features/web/hall/helpers/icon-registry.ts`

- [ ] **Step 1: Add BookMarked to icon registry**

In `dx-web/src/features/web/hall/helpers/icon-registry.ts`, add `BookMarked` to the import and the registry object.

Replace:

```ts
import type { LucideIcon } from "lucide-react"
import {
  LayoutDashboard,
  Gamepad2,
  Users,
  Star,
  Bell,
  Sparkles,
  BookOpen,
  RotateCcw,
  CheckCircle2,
  Trophy,
  Medal,
  MessageCircle,
} from "lucide-react"

export const iconRegistry: Record<string, LucideIcon> = {
  LayoutDashboard,
  Gamepad2,
  Users,
  Star,
  Bell,
  Sparkles,
  BookOpen,
  RotateCcw,
  CheckCircle2,
  Trophy,
  Medal,
  MessageCircle,
}
```

With:

```ts
import type { LucideIcon } from "lucide-react"
import {
  LayoutDashboard,
  Gamepad2,
  Users,
  Star,
  Bell,
  Sparkles,
  BookOpen,
  BookMarked,
  RotateCcw,
  CheckCircle2,
  Trophy,
  Medal,
  MessageCircle,
} from "lucide-react"

export const iconRegistry: Record<string, LucideIcon> = {
  LayoutDashboard,
  Gamepad2,
  Users,
  Star,
  Bell,
  Sparkles,
  BookOpen,
  BookMarked,
  RotateCcw,
  CheckCircle2,
  Trophy,
  Medal,
  MessageCircle,
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-web/src/features/web/hall/helpers/icon-registry.ts
git commit -m "feat(web): add BookMarked to icon registry"
```

---

### Task 4: Frontend — Sync Feature and Route Page

**Files:**
- Create: `dx-web/src/features/web/sync/components/sync-page-content.tsx`
- Create: `dx-web/src/app/(web)/hall/(main)/sync/page.tsx`

- [ ] **Step 1: Create sync-page-content.tsx**

Create `dx-web/src/features/web/sync/components/sync-page-content.tsx`:

```tsx
"use client"

import { useMemo, useState } from "react"
import { Gamepad2 } from "lucide-react"
import { PageSpinner } from "@/components/in/page-spinner"
import { FilterSection } from "@/features/web/games/components/filter-section"
import { GameCard } from "@/features/web/games/components/game-card"
import { useInfinitePublicGames } from "@/features/web/games/hooks/use-infinite-public-games"

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean }
type PressOption = { id: string; name: string }
type Filters = { categoryIds?: string[]; pressId?: string; mode?: string }

type SyncPageContentProps = {
  categories: CategoryOption[]
  presses: PressOption[]
}

export function SyncPageContent({ categories, presses }: SyncPageContentProps) {
  const allCategoryIds = useMemo(() => categories.map((c) => c.id), [categories])
  const [filters, setFilters] = useState<Filters>({})

  // Always scope to sync categories — when no category selected, use all sync IDs
  const apiFilters = useMemo<Filters>(
    () => ({
      ...filters,
      categoryIds: filters.categoryIds ?? allCategoryIds,
    }),
    [filters, allCategoryIds],
  )

  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfinitePublicGames(apiFilters)

  return (
    <>
      <FilterSection
        categories={categories}
        presses={presses}
        filters={filters}
        onFiltersChange={setFilters}
      />

      {isLoading && <PageSpinner size="lg" />}

      {!isLoading && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
          {games.map((game) => (
            <GameCard key={game.id} game={game} />
          ))}
        </div>
      )}

      {isValidating && !isLoading && <PageSpinner size="sm" />}

      {!isLoading && !isValidating && games.length === 0 && (
        <div className="flex flex-col items-center gap-2 py-12 text-center">
          <Gamepad2 className="h-10 w-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">暂无游戏</p>
        </div>
      )}

      {hasMore && <div ref={sentinelRef} className="h-1" />}
    </>
  )
}
```

Key difference from `GamesPageContent`: the `apiFilters` memo ensures the API always receives sync category IDs, even when FilterSection shows "全部分类" (which sends `categoryIds: undefined`). This scopes the game listing to sync categories only.

- [ ] **Step 2: Create sync/page.tsx**

Create `dx-web/src/app/(web)/hall/(main)/sync/page.tsx`:

```tsx
"use client"

import useSWR from "swr"
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
import { SyncPageContent } from "@/features/web/sync/components/sync-page-content"
import { PageSpinner } from "@/components/in/page-spinner"

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean }
type PressOption = { id: string; name: string }

export default function HallSyncPage() {
  const menu = useHallMenuItem("/hall/sync")

  const { data: categories, isLoading: catLoading } =
    useSWR<CategoryOption[]>("/api/game-categories/sync")
  const { data: presses, isLoading: pressLoading } =
    useSWR<PressOption[]>("/api/game-presses")

  const isLoading = catLoading || pressLoading

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <GreetingTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
      {isLoading ? (
        <PageSpinner size="lg" />
      ) : (
        <SyncPageContent
          categories={categories ?? []}
          presses={presses ?? []}
        />
      )}
    </div>
  )
}
```

Same pattern as `games/page.tsx` — fetches categories (from `/api/game-categories/sync` instead of `/api/game-categories`) and presses, shows spinner until ready, then renders content.

- [ ] **Step 3: Run lint on new files**

Run: `cd dx-web && npx eslint src/features/web/sync/components/sync-page-content.tsx "src/app/(web)/hall/(main)/sync/page.tsx"`
Expected: no errors

- [ ] **Step 4: Run full build**

Run: `cd dx-web && npm run build`
Expected: build succeeds with no errors

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/sync/components/sync-page-content.tsx dx-web/src/app/\(web\)/hall/\(main\)/sync/page.tsx
git commit -m "feat(web): add sync practice page with scoped category filters"
```

---

### Task 5: Verification

- [ ] **Step 1: Verify backend endpoints**

Start the API server:

```bash
cd dx-api && go run .
```

Test sync categories endpoint (no auth needed):

```bash
curl -s http://localhost:3001/api/game-categories/sync | jq '.data | length'
```

Expected: a number > 0 (the sync subtree children)

Test main categories endpoint excludes sync:

```bash
curl -s http://localhost:3001/api/game-categories | jq '[.data[] | select(.name == "同步练习")] | length'
```

Expected: `0` (同步练习 no longer appears)

Test menu endpoint includes sync item:

```bash
curl -s http://localhost:3001/api/hall/menus -H "Authorization: Bearer <token>" | jq '[.data[].items[] | select(.label == "同步练习")] | length'
```

Expected: `1`

- [ ] **Step 2: Verify frontend pages**

Start the dev server:

```bash
cd dx-web && npm run dev
```

Open http://localhost:3000/hall in browser and verify:
- Sidebar shows "同步练习" with BookMarked icon, positioned after 学习课程 and before 我的课程
- Click 同步练习 — navigates to `/hall/sync`
- Sync page shows GreetingTopBar with "同步练习" title and "同步参考练习，巩固学习内容" subtitle
- Category filter pills show children of 同步练习 (not 同步练习 itself)
- Press filter pills show all publishers
- Game mode filter pills show all modes
- Games grid shows only games belonging to the sync category subtree
- Clicking a specific child category narrows the games
- Clicking "全部分类" shows all sync games again
- Infinite scroll works

- [ ] **Step 3: Verify no regressions on games page**

Navigate to `/hall/games` and verify:
- 同步练习 category no longer appears in category filter pills
- All other categories still appear and work
- Game listing still works with filters
- Infinite scroll still works
- No blank/broken state

- [ ] **Step 4: Verify other pages unaffected**

- Game detail pages (`/hall/games/{id}`) still work for both sync and non-sync games
- My games (`/hall/games/mine`) still shows played games including sync games
- Favorites still works
- Search still returns sync games
- Mobile sidebar trigger still works
