# Sync Practice Page Design

## Summary

Add a dedicated "同步练习" page (`/hall/sync`) to the hall sidebar. Extract the 同步练习 category subtree from the existing games page into this new page, with its own category filters, press filters, and game listing. The existing games page (`/hall/games`) will no longer show 同步练习 or its children.

## Menu Item

| Field | Value |
|-------|-------|
| Icon | `BookMarked` |
| Label | 同步练习 |
| Subtitle | 同步参考练习，巩固学习内容 |
| Href | `/hall/sync` |
| Position | After 学习课程, before 我的课程 in section 1 |

## Backend

### 1. Category service split

**File:** `dx-api/app/services/api/game_category_service.go`

Add constant:

```go
const SyncCategoryName = "同步练习"
```

Modify `ListCategories()`:
- Same tree-walking logic, but skips the node where `name == SyncCategoryName` and all its descendants
- No rename needed — the only public API caller is `GameCategoryController.Categories`

Add `ListSyncCategories()`:
- Finds the 同步练习 node in the full category tree
- Walks only its subtree
- Adjusts depth: children of 同步练习 become depth 0, grandchildren become depth 1, etc.
- Returns `[]CategoryData` (same shape as existing)
- Returns empty slice if 同步练习 category not found

Both functions share the same internal pattern: load all enabled categories, build parent-children map, walk selectively. Refactor the shared loading and map-building into an internal helper to avoid duplication.

### 2. Controller method

**File:** `dx-api/app/http/controllers/api/game_category_controller.go`

Add `SyncCategories` method:

```go
func (c *GameCategoryController) SyncCategories(ctx contractshttp.Context) contractshttp.Response {
    categories, err := services.ListSyncCategories()
    if err != nil {
        return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list sync categories")
    }
    return helpers.Success(ctx, categories)
}
```

No changes to existing `Categories` method — it still calls `services.ListCategories()`, which now excludes the sync subtree.

### 3. Route

**File:** `dx-api/routes/api.go`

Add alongside existing `/game-categories`:

```go
router.Get("/game-categories/sync", gameCategoryController.SyncCategories)
```

Both routes are public (no auth required).

### 4. Hall menu item

**File:** `dx-api/app/consts/hall_menu.go`

Insert after the 学习课程 entry in section 1:

```go
{Icon: "BookMarked", Label: "同步练习", Subtitle: "同步参考练习，巩固学习内容", Href: "/hall/sync"},
```

### API Response Shape

`GET /api/game-categories/sync`:

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    { "id": "child-1", "name": "三年级上册", "depth": 0, "isLeaf": false },
    { "id": "grandchild-1", "name": "Unit 1", "depth": 1, "isLeaf": true },
    { "id": "grandchild-2", "name": "Unit 2", "depth": 1, "isLeaf": true },
    { "id": "child-2", "name": "四年级上册", "depth": 0, "isLeaf": true }
  ]
}
```

`GET /api/game-categories` now returns the same structure as before, minus the 同步练习 subtree.

## Frontend

### 1. Icon registry

**File:** `dx-web/src/features/web/hall/helpers/icon-registry.ts`

Add `BookMarked` import and registry entry.

### 2. New feature directory

```
dx-web/src/features/web/sync/
├── components/
│   └── sync-page-content.tsx
└── hooks/
    └── use-sync-categories.ts
```

### 3. Hook: `use-sync-categories.ts`

```ts
export function useSyncCategories() {
  return useSWR<CategoryOption[]>("/api/game-categories/sync")
}
```

Returns `{ data, isLoading }` — same category shape as the existing hook.

### 4. Component: `sync-page-content.tsx`

Same structure as `GamesPageContent`:
- Composes `FilterSection` + `GameCard` grid via `useInfinitePublicGames`
- Receives `categories` prop (sync subtree). Computes `allCategoryIds = categories.map(c => c.id)` once
- Initial filter state: `{ categoryIds: allCategoryIds }` — shows all sync games by default
- "全部分类" click resets to `allCategoryIds`; child click narrows to `[childId]` (or `[childId, ...grandchildren]`)
- Reuses: `FilterSection`, `GameCard`, `useInfinitePublicGames`, `PageSpinner`
- Note: games categorized directly under the root 同步练习 (not a child) won't appear. This is acceptable — in the textbook-aligned structure, games are assigned to leaf categories (units), not the root.

### 5. New page: `sync/page.tsx`

**File:** `dx-web/src/app/(web)/hall/(main)/sync/page.tsx`

```
"use client"
- useSyncCategories() for category pills
- useSWR("/api/game-presses") for press pills
- useHallMenuItem("/hall/sync") for title/subtitle
- GreetingTopBar with menu title
- SyncPageContent with categories + presses
```

Same pattern as `games/page.tsx`.

### 6. Existing games page

**No code changes** to `games/page.tsx` — it already calls `useSWR("/api/game-categories")` which will now return categories minus the 同步练习 subtree via the updated backend.

## Files Changed

**New files:**
- `dx-web/src/features/web/sync/components/sync-page-content.tsx`
- `dx-web/src/features/web/sync/hooks/use-sync-categories.ts`
- `dx-web/src/app/(web)/hall/(main)/sync/page.tsx`

**Modified files:**
- `dx-api/app/consts/hall_menu.go` — add menu item
- `dx-api/app/services/api/game_category_service.go` — split into main/sync, add constant
- `dx-api/app/http/controllers/api/game_category_controller.go` — add SyncCategories method
- `dx-api/routes/api.go` — add sync categories route
- `dx-web/src/features/web/hall/helpers/icon-registry.ts` — add BookMarked

## Not Changed

- `FilterSection`, `GameCard`, `useInfinitePublicGames`, `game-card.ts` helper — reused as-is
- Game detail page (`/hall/games/[id]`) — sync games still link here
- `games/page.tsx` — no code changes, backend handles exclusion
- `api-client.ts` — the SWR hooks use URL strings directly, no need for explicit API client methods
- My games page, favorites, search — sync games still appear there as expected
