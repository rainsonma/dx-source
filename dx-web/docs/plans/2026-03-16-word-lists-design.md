# Word Lists (生词本 / 复习本 / 已掌握) — Design

Wire real data to the three word list pages with infinite scroll pagination and batch/single delete.

## Pages

| Page | Route | Feature Module |
|------|-------|----------------|
| 生词本 | `/hall/unknown` | `user-unknown` |
| 复习本 | `/hall/review` | `user-review` |
| 已掌握 | `/hall/mastered` | `user-master` |

## Shared Components → `src/components/in/`

Consolidate identical copies from the 3 features into shared components:

### word-table.tsx

Generic `<WordTable<T>>` with:
- Working header checkbox (select-all) + per-row checkboxes
- Batch delete toolbar: appears when `selectedIds.size > 0`, shows count + "删除" button
- Per-row trash icon for single delete
- Optional action column slot (for review's "去复习" button)
- Callbacks: `onDelete(id)`, `onDeleteBatch(ids)`, `onSelectChange(selectedIds)`

### stat-card.tsx

Moved as-is from feature modules.

### badge.tsx

Moved as-is from feature modules.

### delete-confirm-dialog.tsx

AlertDialog (shadcn/ui) for delete confirmation:
- Title: "确认删除"
- Single: "确定要删除这个词汇吗？"
- Batch: "确定要删除选中的 N 个词汇吗？"
- Actions: 取消 / 确认删除 (destructive variant)

## Data Layer — Models

### Queries (new files)

**`src/models/user-unknown/user-unknown.query.ts`**
- `getUserUnknowns(userId, cursor?, limit=20)` — cursor pagination, joins ContentItem (content, translation, contentType) + Game (name), ordered by `createdAt desc`
- `getUserUnknownStats(userId)` — `{ total, today, lastThreeDays }`

**`src/models/user-review/user-review.query.ts`**
- `getUserReviews(userId, cursor?, limit=20)` — joins ContentItem + Game (name, slug), includes lastReviewAt, nextReviewAt, reviewCount, ordered by `nextReviewAt asc` (most urgent first)
- `getUserReviewStats(userId)` — `{ pending, overdue, reviewedToday }`

**`src/models/user-master/user-master.query.ts`**
- `getUserMasters(userId, cursor?, limit=20)` — joins ContentItem + Game (name), ordered by `masteredAt desc`
- `getUserMasterStats(userId)` — `{ total, thisWeek, thisMonth }`

### Mutations (added to existing files)

Each existing `*.mutation.ts` gets:
- `deleteUserUnknown(id)` / `deleteUserUnknowns(ids: string[])`
- `deleteUserReview(id)` / `deleteUserReviews(ids: string[])`
- `deleteUserMaster(id)` / `deleteUserMasters(ids: string[])`

## Server Actions

Each feature gets `actions/*.action.ts` with `"use server"`:

**Pattern:** get session via `auth()`, validate userId, call model, return `{ success, data }` | `{ error }`.

| Feature | Actions |
|---------|---------|
| user-unknown | `fetchUnknownsAction(cursor?)`, `deleteUnknownAction(id)`, `deleteUnknownsAction(ids)` |
| user-review | `fetchReviewsAction(cursor?)`, `deleteReviewAction(id)`, `deleteReviewsAction(ids)` |
| user-master | `fetchMastersAction(cursor?)`, `deleteMasterAction(id)`, `deleteMastersAction(ids)` |

## Client Hooks

One per feature, same structure:

**`use-unknown-list.ts`** / **`use-review-list.ts`** / **`use-master-list.ts`**

State: `items`, `cursor`, `hasMore`, `isLoading`, `selectedIds: Set<string>`

IntersectionObserver with `sentinelRef` + `rootMargin: "200px"`.

Returns: `items`, `isLoading`, `hasMore`, `sentinelRef`, `selectedIds`, `toggleSelect(id)`, `toggleSelectAll()`, `deleteOne(id)`, `deleteSelected()`, `stats`.

Optimistic updates on delete (remove from list, decrement stats). Toast feedback via sonner. Revert on failure.

## Content Components

All rewritten as `"use client"` components receiving `initialItems`, `initialCursor`, `initialStats` as props.

### unknown-content.tsx

- 3 StatCards: 全部生词, 今日添加, 最近三天
- WordTable columns: content + translation | 来源 (game name) | 添加时间

### review-content.tsx

- 3 StatCards: 待复习, 已逾期, 今日已复习
- WordTable columns: content + translation | 上次复习 / 下次复习 (red if overdue) | 紧急度 badge | 去复习 (Link → `/hall/game/{slug}`)

### master-content.tsx

- 3 StatCards: 已掌握总数, 本周掌握, 本月掌握
- WordTable columns: content + translation | 来源 (game name) | 掌握时间

## Page Components (Server Components)

All 3 pages stay thin:

```typescript
async function Page() {
  const session = await auth()
  const userId = session?.user?.id
  const [{ items, nextCursor }, stats] = await Promise.all([
    getItems(userId),
    getStats(userId),
  ])
  return <Content initialItems={items} initialCursor={nextCursor} initialStats={stats} />
}
```

## Delete Confirmation

All deletes (single + batch) trigger AlertDialog before executing.

## Naming

- All new code uses "master" (not "mastered") in function names, variables, file names
- Route URL stays `/hall/mastered` (no URL change)

## Files Created/Modified

### New files
- `src/components/in/word-table.tsx`
- `src/components/in/stat-card.tsx`
- `src/components/in/badge.tsx`
- `src/components/in/delete-confirm-dialog.tsx`
- `src/models/user-unknown/user-unknown.query.ts`
- `src/models/user-review/user-review.query.ts`
- `src/models/user-master/user-master.query.ts`
- `src/features/web/user-unknown/actions/unknown.action.ts`
- `src/features/web/user-review/actions/review.action.ts`
- `src/features/web/user-master/actions/master.action.ts`
- `src/features/web/user-unknown/hooks/use-unknown-list.ts`
- `src/features/web/user-review/hooks/use-review-list.ts`
- `src/features/web/user-master/hooks/use-master-list.ts`

### Modified files
- `src/models/user-unknown/user-unknown.mutation.ts` — add delete functions
- `src/models/user-review/user-review.mutation.ts` — add delete functions
- `src/models/user-master/user-master.mutation.ts` — add delete functions
- `src/features/web/user-unknown/components/unknown-content.tsx` — rewrite with real data
- `src/features/web/user-review/components/review-content.tsx` — rewrite with real data
- `src/features/web/user-mastered/components/mastered-content.tsx` → rename to `master-content.tsx`, rewrite
- `src/app/(web)/hall/(main)/unknown/page.tsx` — server-side data fetching
- `src/app/(web)/hall/(main)/review/page.tsx` — server-side data fetching
- `src/app/(web)/hall/(main)/mastered/page.tsx` — server-side data fetching, import rename

### Deleted files
- `src/features/web/user-unknown/components/word-table.tsx` (moved to shared)
- `src/features/web/user-unknown/components/stat-card.tsx` (moved to shared)
- `src/features/web/user-review/components/word-table.tsx` (moved to shared)
- `src/features/web/user-review/components/stat-card.tsx` (moved to shared)
- `src/features/web/user-mastered/components/word-table.tsx` (moved to shared)
- `src/features/web/user-mastered/components/stat-card.tsx` (moved to shared)
