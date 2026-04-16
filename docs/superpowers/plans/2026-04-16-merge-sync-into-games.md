# Merge 同步练习 Into /hall/games — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the 同步练习 category and its games back into the unified `/hall/games` page, remove the separate `/hall/sync` route and sidebar menu item, and conditionally show the press filter when a sync subcategory is selected.

**Architecture:** Backend removes the sync/non-sync category split so all categories and games are returned together. Frontend detects the sync root by name and conditionally shows the press filter. The `/hall/sync` page and its components are deleted.

**Tech Stack:** Go/Goravel (backend), Next.js/React/SWR (frontend), TypeScript

---

## File Map

| Action | File | Responsibility |
|--------|------|---------------|
| Modify | `dx-api/app/services/api/game_category_service.go` | Remove sync-specific functions and constants |
| Modify | `dx-api/app/services/api/game_service.go` | Remove sync exclusion from default game listing |
| Modify | `dx-api/app/http/controllers/api/game_category_controller.go` | Remove SyncCategories method |
| Modify | `dx-api/routes/api.go` | Remove `/game-categories/sync` route |
| Modify | `dx-api/app/consts/hall_menu.go` | Remove 同步练习 menu item |
| Modify | `dx-web/src/app/(web)/hall/(main)/games/page.tsx` | Add presses fetch, pass to content |
| Modify | `dx-web/src/features/web/games/components/games-page-content.tsx` | Add sync detection, conditional press filter |
| Delete | `dx-web/src/app/(web)/hall/(main)/sync/page.tsx` | Sync page route |
| Delete | `dx-web/src/features/web/sync/components/sync-page-content.tsx` | Sync page content component |

---

### Task 1: Remove sync-specific backend logic

**Files:**
- Modify: `dx-api/app/services/api/game_category_service.go`

- [ ] **Step 1: Simplify `loadCategoryTree()` — remove syncID tracking**

Replace the entire file with this content. The `categoryTree` struct drops the `syncID` field, `loadCategoryTree()` no longer scans for the sync category name, and the three sync-specific functions (`ListSyncCategories`, `SyncCategoryIDs`, `SyncCategoryName` constant) are removed. `ListCategories()` no longer skips any category.

```go
package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

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
	}

	return tree, nil
}

// ListCategories returns all enabled categories in hierarchical order.
func ListCategories() ([]CategoryData, error) {
	tree, err := loadCategoryTree()
	if err != nil {
		return nil, err
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
	walk("", 0)

	return result, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: Compilation errors in `game_service.go` (references `SyncCategoryIDs`) and `game_category_controller.go` (references `ListSyncCategories`). These are fixed in the next steps.

---

### Task 2: Remove sync exclusion from game listing

**Files:**
- Modify: `dx-api/app/services/api/game_service.go:66-95`

- [ ] **Step 1: Simplify the category filter in `ListPublishedGames()`**

Replace the category filter block (lines 76-89) with a simple conditional:

```go
	if len(categoryIDs) > 0 {
		query = query.Where("game_category_id IN ?", categoryIDs)
	}
```

This removes the entire `else` branch that called `SyncCategoryIDs()` and excluded sync games. When no category filter is provided, all published games are returned.

- [ ] **Step 2: Remove unused import if needed**

After removing the `SyncCategoryIDs()` call, check if the `"sort"` and `"time"` imports are still used (they are — by `GetPlayedGames()`). No import changes needed.

- [ ] **Step 3: Verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: Still fails on `game_category_controller.go`. Fixed in next task.

---

### Task 3: Remove sync controller method and route

**Files:**
- Modify: `dx-api/app/http/controllers/api/game_category_controller.go`
- Modify: `dx-api/routes/api.go:55`

- [ ] **Step 1: Remove `SyncCategories()` from the controller**

Replace the entire controller file:

```go
package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type GameCategoryController struct{}

func NewGameCategoryController() *GameCategoryController {
	return &GameCategoryController{}
}

// Categories returns all enabled categories in hierarchical order.
func (c *GameCategoryController) Categories(ctx contractshttp.Context) contractshttp.Response {
	categories, err := services.ListCategories()
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list categories")
	}

	return helpers.Success(ctx, categories)
}
```

- [ ] **Step 2: Remove the sync route from `api.go`**

In `dx-api/routes/api.go`, delete line 55:
```go
		router.Get("/game-categories/sync", gameCategoryController.SyncCategories)
```

- [ ] **Step 3: Verify full compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: BUILD SUCCESS — all sync references are now removed.

- [ ] **Step 4: Commit backend changes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/game_category_service.go \
        dx-api/app/services/api/game_service.go \
        dx-api/app/http/controllers/api/game_category_controller.go \
        dx-api/routes/api.go
git commit -m "refactor(api): remove sync category separation from game listing"
```

---

### Task 4: Remove 同步练习 from hall menu

**Files:**
- Modify: `dx-api/app/consts/hall_menu.go`

- [ ] **Step 1: Delete the sync menu item**

In `dx-api/app/consts/hall_menu.go`, remove this line from the first section's Items slice:
```go
			{Icon: "BookMarked", Label: "同步练习", Subtitle: "同步参考练习，巩固学习内容", Href: "/hall/sync"},
```

The first section should now be:
```go
		{Items: []HallMenuItem{
			{Icon: "LayoutDashboard", Label: "我的主页", Subtitle: "", Href: "/hall"},
			{Icon: "Gamepad2", Label: "学习课程", Subtitle: "选择一个游戏模式，边玩边学英语！", Href: "/hall/games"},
			{Icon: "Gamepad2", Label: "我的课程", Subtitle: "你玩过的所有学习课程", Href: "/hall/games/mine"},
			{Icon: "Users", Label: "学习群组", Subtitle: "浏览并加入学习群组，与小伙伴一起进步", Href: "/hall/groups"},
			{Icon: "Star", Label: "我的收藏", Subtitle: "收藏你喜欢的课程游戏和学习内容", Href: "/hall/favorites"},
			{Icon: "Trophy", Label: "排行榜单", Subtitle: "查看学习排名，与好友一起进步", Href: "/hall/leaderboard"},
		}},
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/consts/hall_menu.go
git commit -m "feat(api): remove sync practice from hall sidebar menu"
```

---

### Task 5: Update games page to fetch presses

**Files:**
- Modify: `dx-web/src/app/(web)/hall/(main)/games/page.tsx`

- [ ] **Step 1: Add press fetching and pass to content**

Replace the entire file:

```tsx
"use client"

import useSWR from "swr"
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
import { GamesPageContent } from "@/features/web/games/components/games-page-content"
import { PageSpinner } from "@/components/in/page-spinner"

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean }
type PressOption = { id: string; name: string }

export default function HallGamesPage() {
  const menu = useHallMenuItem("/hall/games")

  const { data: categories, isLoading: catLoading } =
    useSWR<CategoryOption[]>("/api/game-categories")
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
        <GamesPageContent
          categories={categories ?? []}
          presses={presses ?? []}
        />
      )}
    </div>
  )
}
```

- [ ] **Step 2: Verify no TypeScript errors (will fail until next task)**

The `presses` prop doesn't exist on `GamesPageContent` yet. This is fixed in Task 6.

---

### Task 6: Add conditional press filter to GamesPageContent

**Files:**
- Modify: `dx-web/src/features/web/games/components/games-page-content.tsx`

- [ ] **Step 1: Add presses prop, sync detection, and conditional showPresses**

Replace the entire file:

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

type GamesPageContentProps = {
  categories: CategoryOption[]
  presses: PressOption[]
}

const SYNC_CATEGORY_NAME = "同步练习"

export function GamesPageContent({ categories, presses }: GamesPageContentProps) {
  const [filters, setFilters] = useState<Filters>({})
  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfinitePublicGames(filters)

  const syncRootId = useMemo(
    () => categories.find((c) => c.name === SYNC_CATEGORY_NAME && c.depth === 0)?.id,
    [categories],
  )

  const syncChildIds = useMemo(() => {
    if (!syncRootId) return new Set<string>()
    const idx = categories.findIndex((c) => c.id === syncRootId)
    const ids = new Set<string>([syncRootId])
    for (let i = idx + 1; i < categories.length; i++) {
      if (categories[i].depth === 0) break
      ids.add(categories[i].id)
    }
    return ids
  }, [categories, syncRootId])

  const showPresses = useMemo(() => {
    if (!syncRootId || !filters.categoryIds?.length) return false
    return filters.categoryIds.some((id) => syncChildIds.has(id))
  }, [syncRootId, filters.categoryIds, syncChildIds])

  function handleFiltersChange(next: Filters) {
    if (next.pressId && syncRootId) {
      const nextInSync = next.categoryIds?.some((id) => syncChildIds.has(id)) ?? false
      if (!nextInSync) {
        setFilters({ ...next, pressId: undefined })
        return
      }
    }
    setFilters(next)
  }

  return (
    <>
      <FilterSection
        categories={categories}
        presses={presses}
        filters={filters}
        onFiltersChange={handleFiltersChange}
        showPresses={showPresses}
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

Key additions:
- `presses` prop passed through to `FilterSection`
- `syncRootId`: finds the 同步练习 category at depth 0
- `syncChildIds`: collects sync root + all its children (depth > 0 until next depth-0 category)
- `showPresses`: true only when the selected category filter overlaps the sync subtree

- [ ] **Step 2: Verify lint**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx next lint`
Expected: No errors

- [ ] **Step 3: Commit frontend changes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/app/\(web\)/hall/\(main\)/games/page.tsx \
        dx-web/src/features/web/games/components/games-page-content.tsx
git commit -m "feat(web): add sync categories and conditional press filter to games page"
```

---

### Task 7: Delete sync page and components

**Files:**
- Delete: `dx-web/src/app/(web)/hall/(main)/sync/page.tsx`
- Delete: `dx-web/src/features/web/sync/components/sync-page-content.tsx`
- Delete: `dx-web/src/features/web/sync/` directory
- Delete: `dx-web/src/app/(web)/hall/(main)/sync/` directory

- [ ] **Step 1: Delete sync page route**

```bash
rm -rf /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web/src/app/\(web\)/hall/\(main\)/sync
```

- [ ] **Step 2: Delete sync feature directory**

```bash
rm -rf /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web/src/features/web/sync
```

- [ ] **Step 3: Verify no broken imports**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx next lint`
Expected: No errors (nothing else imports from the deleted files)

- [ ] **Step 4: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run build`
Expected: BUILD SUCCESS

- [ ] **Step 5: Commit cleanup**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add -A dx-web/src/app/\(web\)/hall/\(main\)/sync \
           dx-web/src/features/web/sync
git commit -m "chore(web): remove sync page and components"
```

---

### Task 8: Verify end-to-end

- [ ] **Step 1: Start backend**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run .`
Expected: Server starts on port 3001

- [ ] **Step 2: Verify `/api/game-categories` returns sync categories**

Run: `curl -s http://localhost:3001/api/game-categories | jq '.data[] | select(.name == "同步练习")'`
Expected: Returns the 同步练习 category object with depth 0

- [ ] **Step 3: Verify `/api/game-categories/sync` is gone**

Run: `curl -s -o /dev/null -w "%{http_code}" http://localhost:3001/api/game-categories/sync`
Expected: 404

- [ ] **Step 4: Verify `/api/hall/menus` has no sync entry**

This requires auth. Verify by checking the menu constant directly:

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source && grep -c "sync" dx-api/app/consts/hall_menu.go`
Expected: 0

- [ ] **Step 5: Start frontend and test in browser**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run dev`

Open `http://localhost:3000/hall/games` and verify:
1. All categories appear in the filter, including 同步练习
2. Clicking 同步练习 shows its subcategories (一年级, 二年级, etc.) and the 版本 (press) filter row appears
3. Clicking a non-sync category hides the press filter
4. Games from all categories (including sync) appear in the default unfiltered view
5. Sidebar no longer shows 同步练习 as a separate menu item
6. Navigating to `/hall/sync` shows a 404 page
