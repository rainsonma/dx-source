# Merge 同步练习 Back Into /hall/games

**Date:** 2026-04-16
**Status:** Approved

## Summary

Move the 同步练习 category and all its subcategories/games back into the unified `/hall/games` page. Remove the separate `/hall/sync` route and sidebar menu item entirely. The press/publisher filter (版本) appears conditionally when a sync subcategory is selected.

## Motivation

The 同步练习 category was previously split into its own page at `/hall/sync`. After reconsideration, it should live alongside all other categories on `/hall/games` for a simpler, unified browsing experience.

## Backend Changes

### game_category_service.go

- **`ListCategories()`** — Remove the `if cat.ID == tree.syncID { continue }` skip in the walk function. Returns all enabled categories including 同步练习 and its children (一年级 through 高中/中职) at their natural tree depths (0 and 1).
- **Delete `ListSyncCategories()`** — No longer needed.
- **Delete `SyncCategoryIDs()`** — No longer needed.
- **Delete `SyncCategoryName` constant** — No longer referenced.
- **Delete `categoryTree.syncID` field** — No longer needed; `loadCategoryTree()` no longer tracks it.

### game_service.go — ListPublishedGames()

Remove the sync exclusion block (the `else` branch that calls `SyncCategoryIDs()` and adds `NOT IN` filter). When no `categoryIDs` filter is provided, the query returns all published, active, non-private games regardless of category.

### game_category_controller.go

Delete the `SyncCategories()` controller method.

### routes/api.go

Remove the route: `router.Get("/game-categories/sync", gameCategoryController.SyncCategories)`.

### consts/hall_menu.go

Remove the 同步练习 menu item from `HallMenuSections()`:
```go
{Icon: "BookMarked", Label: "同步练习", Subtitle: "同步参考练习，巩固学习内容", Href: "/hall/sync"},
```

## Frontend Changes

### Games page — /hall/games/page.tsx

Add `useSWR<PressOption[]>("/api/game-presses")` alongside the existing categories fetch. Pass `presses` to `GamesPageContent`. Combine loading states.

### GamesPageContent — games-page-content.tsx

- Accept new `presses: PressOption[]` prop.
- Detect the sync root category: find the category where `name === "同步练习"` and `depth === 0`.
- Compute `showPresses` dynamically: `true` when the currently selected `categoryIds` filter contains the sync root ID or any of its depth-1 children.
- Pass `presses` and computed `showPresses` to `FilterSection`.

### FilterSection — filter-section.tsx

No changes. Already supports `presses`, `showPresses` props with conditional rendering.

## Cleanup — Delete Entirely

- `dx-web/src/app/(web)/hall/(main)/sync/` directory (route page)
- `dx-web/src/features/web/sync/` directory (SyncPageContent component)

## Constraints

- No breaking changes to existing game detail pages, play sessions, or any other functionality.
- The press filter only appears when the user has navigated into the sync category subtree.
- All existing category/game data remains unchanged in the database.
- No database migrations needed.
