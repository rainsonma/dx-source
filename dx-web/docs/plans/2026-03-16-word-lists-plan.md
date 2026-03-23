# Word Lists (生词本 / 复习本 / 已掌握) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire real data to the three word list pages with infinite scroll and batch/single delete.

**Architecture:** Server components fetch initial page + stats, pass to client content components. Client hooks manage infinite scroll via IntersectionObserver and cursor-based pagination. Server actions handle loadMore and delete operations. Shared UI components for word table, stat cards, badges, and delete confirmation dialog.

**Tech Stack:** Next.js 16 App Router, Prisma v7, Server Actions, shadcn/ui AlertDialog, sonner toasts, IntersectionObserver

---

### Task 1: Shared Components — StatCard, Badge, DeleteConfirmDialog

**Files:**
- Create: `src/components/in/stat-card.tsx`
- Create: `src/components/in/badge.tsx`
- Create: `src/components/in/delete-confirm-dialog.tsx`

**Step 1: Create StatCard**

Move from `src/features/web/user-unknown/components/stat-card.tsx` to shared location:

```tsx
// src/components/in/stat-card.tsx
import type { LucideIcon } from "lucide-react";

interface StatCardProps {
  icon: LucideIcon;
  iconBg: string;
  iconColor: string;
  value: string;
  label: string;
}

/** Stat card displaying an icon, value, and label */
export function StatCard({ icon: Icon, iconBg, iconColor, value, label }: StatCardProps) {
  return (
    <div className="flex items-center gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3 md:px-5 md:py-4">
      <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${iconBg}`}>
        <Icon className={`h-5 w-5 ${iconColor}`} />
      </div>
      <div>
        <p className="text-lg font-bold text-slate-900 md:text-xl">{value}</p>
        <p className="text-xs text-slate-400">{label}</p>
      </div>
    </div>
  );
}
```

**Step 2: Create Badge**

```tsx
// src/components/in/badge.tsx

/** Colored badge for status/category labels */
export function Badge({ label, bg, text }: { label: string; bg: string; text: string }) {
  return (
    <span className={`rounded-md px-2.5 py-1 text-[11px] font-semibold ${bg} ${text}`}>
      {label}
    </span>
  );
}
```

**Step 3: Create DeleteConfirmDialog**

```tsx
// src/components/in/delete-confirm-dialog.tsx
"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface DeleteConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  count: number;
  onConfirm: () => void;
}

/** Alert dialog for confirming single or batch delete */
export function DeleteConfirmDialog({ open, onOpenChange, count, onConfirm }: DeleteConfirmDialogProps) {
  const message =
    count === 1
      ? "确定要删除这个词汇吗？"
      : `确定要删除选中的 ${count} 个词汇吗？`;

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除</AlertDialogTitle>
          <AlertDialogDescription>{message}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={onConfirm}>
            确认删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

**Step 4: Verify build**

Run: `npm run build`

**Step 5: Commit**

```bash
git add src/components/in/stat-card.tsx src/components/in/badge.tsx src/components/in/delete-confirm-dialog.tsx
git commit -m "feat: add shared StatCard, Badge, and DeleteConfirmDialog components"
```

---

### Task 2: Shared Component — WordTable (enhanced)

**Files:**
- Create: `src/components/in/word-table.tsx`

**Step 1: Create enhanced WordTable**

```tsx
// src/components/in/word-table.tsx
"use client";

import { Trash2 } from "lucide-react";

export interface WordRow {
  id: string;
  content: string;
  translation: string | null;
}

export interface ColumnConfig<T extends WordRow> {
  key: string;
  label: string;
  width: string;
  render: (item: T) => React.ReactNode;
}

interface WordTableProps<T extends WordRow> {
  items: T[];
  columns: ColumnConfig<T>[];
  selectedIds: Set<string>;
  onSelectChange: (ids: Set<string>) => void;
  onDelete: (id: string) => void;
  onDeleteBatch: () => void;
}

/** Generic word table with checkboxes, single and batch delete */
export function WordTable<T extends WordRow>({
  items,
  columns,
  selectedIds,
  onSelectChange,
  onDelete,
  onDeleteBatch,
}: WordTableProps<T>) {
  const allSelected = items.length > 0 && selectedIds.size === items.length;

  /** Toggle all checkboxes */
  const handleToggleAll = () => {
    if (allSelected) {
      onSelectChange(new Set());
    } else {
      onSelectChange(new Set(items.map((i) => i.id)));
    }
  };

  /** Toggle a single checkbox */
  const handleToggle = (id: string) => {
    const next = new Set(selectedIds);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    onSelectChange(next);
  };

  return (
    <div className="overflow-x-auto">
      {/* Batch toolbar */}
      {selectedIds.size > 0 && (
        <div className="mb-2 flex items-center gap-3 rounded-lg bg-red-50 px-4 py-2">
          <span className="text-sm text-red-700">
            已选择 {selectedIds.size} 项
          </span>
          <button
            onClick={onDeleteBatch}
            className="rounded-md bg-red-500 px-3 py-1 text-xs font-medium text-white hover:bg-red-600"
          >
            删除
          </button>
        </div>
      )}

      <div className="min-w-[600px] overflow-hidden rounded-xl border border-slate-200 bg-white">
        {/* Header */}
        <div className="flex items-center gap-4 bg-slate-50 px-5 py-3.5">
          <input
            type="checkbox"
            checked={allSelected}
            onChange={handleToggleAll}
            className="h-4 w-4 shrink-0 rounded border-slate-300"
          />
          <span className="flex-1 text-xs font-semibold text-slate-500">
            {columns[0]?.label ?? "词汇"}
          </span>
          {columns.slice(1).map((col) => (
            <span
              key={col.key}
              className={`${col.width} text-xs font-semibold text-slate-500`}
            >
              {col.label}
            </span>
          ))}
          <div className="w-5" />
        </div>

        {/* Rows */}
        {items.map((item) => (
          <div
            key={item.id}
            className="flex items-center gap-4 border-t border-slate-100 px-5 py-3.5"
          >
            <input
              type="checkbox"
              checked={selectedIds.has(item.id)}
              onChange={() => handleToggle(item.id)}
              className="h-4 w-4 shrink-0 rounded border-slate-300"
            />
            <div className="flex flex-1 flex-col gap-0.5">
              <span className="text-sm font-semibold text-slate-900">
                {item.content}
              </span>
              <span className="text-xs text-slate-400">
                {item.translation ?? ""}
              </span>
            </div>
            {columns.slice(1).map((col) => (
              <div key={col.key} className={col.width}>
                {col.render(item)}
              </div>
            ))}
            <Trash2
              className="h-[18px] w-[18px] shrink-0 cursor-pointer text-slate-400 hover:text-slate-600"
              onClick={() => onDelete(item.id)}
            />
          </div>
        ))}

        {/* Empty state */}
        {items.length === 0 && (
          <div className="flex items-center justify-center py-12 text-sm text-slate-400">
            暂无数据
          </div>
        )}
      </div>
    </div>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/components/in/word-table.tsx
git commit -m "feat: add shared WordTable component with checkboxes and delete"
```

---

### Task 3: Model Queries — UserUnknown

**Files:**
- Create: `src/models/user-unknown/user-unknown.query.ts`

**Step 1: Create query file**

```ts
// src/models/user-unknown/user-unknown.query.ts
import "server-only";

import { db } from "@/lib/db";

const PAGE_SIZE = 20;

/** Paginated unknown words for a user, ordered newest first */
export async function getUserUnknowns(userId: string, cursor?: string, limit = PAGE_SIZE) {
  const items = await db.userUnknown.findMany({
    where: { userId },
    orderBy: { createdAt: "desc" },
    take: limit + 1,
    ...(cursor && { cursor: { id: cursor }, skip: 1 }),
    select: {
      id: true,
      createdAt: true,
      contentItem: {
        select: {
          id: true,
          content: true,
          translation: true,
          contentType: true,
        },
      },
      game: {
        select: { id: true, name: true },
      },
    },
  });

  const hasMore = items.length > limit;
  const list = hasMore ? items.slice(0, limit) : items;
  const nextCursor = hasMore ? list[list.length - 1].id : null;

  return { items: list, nextCursor };
}

/** Return type for a single unknown item */
export type UnknownItem = Awaited<ReturnType<typeof getUserUnknowns>>["items"][number];

/** Stats counts for the unknown page header */
export async function getUserUnknownStats(userId: string) {
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const threeDaysAgo = new Date(startOfToday);
  threeDaysAgo.setDate(threeDaysAgo.getDate() - 3);

  const [total, today, lastThreeDays] = await Promise.all([
    db.userUnknown.count({ where: { userId } }),
    db.userUnknown.count({ where: { userId, createdAt: { gte: startOfToday } } }),
    db.userUnknown.count({ where: { userId, createdAt: { gte: threeDaysAgo } } }),
  ]);

  return { total, today, lastThreeDays };
}

/** Return type for unknown stats */
export type UnknownStats = Awaited<ReturnType<typeof getUserUnknownStats>>;
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/models/user-unknown/user-unknown.query.ts
git commit -m "feat: add paginated query and stats for user unknowns"
```

---

### Task 4: Model Queries — UserReview

**Files:**
- Create: `src/models/user-review/user-review.query.ts`

**Step 1: Create query file**

```ts
// src/models/user-review/user-review.query.ts
import "server-only";

import { db } from "@/lib/db";

const PAGE_SIZE = 20;

/** Paginated review words for a user, ordered by most urgent first */
export async function getUserReviews(userId: string, cursor?: string, limit = PAGE_SIZE) {
  const items = await db.userReview.findMany({
    where: { userId },
    orderBy: { nextReviewAt: { sort: "asc", nulls: "last" } },
    take: limit + 1,
    ...(cursor && { cursor: { id: cursor }, skip: 1 }),
    select: {
      id: true,
      lastReviewAt: true,
      nextReviewAt: true,
      reviewCount: true,
      createdAt: true,
      contentItem: {
        select: {
          id: true,
          content: true,
          translation: true,
          contentType: true,
        },
      },
      game: {
        select: { id: true, name: true },
      },
    },
  });

  const hasMore = items.length > limit;
  const list = hasMore ? items.slice(0, limit) : items;
  const nextCursor = hasMore ? list[list.length - 1].id : null;

  return { items: list, nextCursor };
}

/** Return type for a single review item */
export type ReviewItem = Awaited<ReturnType<typeof getUserReviews>>["items"][number];

/** Stats counts for the review page header */
export async function getUserReviewStats(userId: string) {
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());

  const [pending, overdue, reviewedToday] = await Promise.all([
    db.userReview.count({ where: { userId, nextReviewAt: { lte: now } } }),
    db.userReview.count({ where: { userId, nextReviewAt: { lt: startOfToday } } }),
    db.userReview.count({ where: { userId, lastReviewAt: { gte: startOfToday } } }),
  ]);

  return { pending, overdue, reviewedToday };
}

/** Return type for review stats */
export type ReviewStats = Awaited<ReturnType<typeof getUserReviewStats>>;
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/models/user-review/user-review.query.ts
git commit -m "feat: add paginated query and stats for user reviews"
```

---

### Task 5: Model Queries — UserMaster

**Files:**
- Create: `src/models/user-master/user-master.query.ts`

**Step 1: Create query file**

```ts
// src/models/user-master/user-master.query.ts
import "server-only";

import { db } from "@/lib/db";

const PAGE_SIZE = 20;

/** Paginated mastered words for a user, ordered newest first */
export async function getUserMasters(userId: string, cursor?: string, limit = PAGE_SIZE) {
  const items = await db.userMaster.findMany({
    where: { userId },
    orderBy: { masteredAt: { sort: "desc", nulls: "last" } },
    take: limit + 1,
    ...(cursor && { cursor: { id: cursor }, skip: 1 }),
    select: {
      id: true,
      masteredAt: true,
      createdAt: true,
      contentItem: {
        select: {
          id: true,
          content: true,
          translation: true,
          contentType: true,
        },
      },
      game: {
        select: { id: true, name: true },
      },
    },
  });

  const hasMore = items.length > limit;
  const list = hasMore ? items.slice(0, limit) : items;
  const nextCursor = hasMore ? list[list.length - 1].id : null;

  return { items: list, nextCursor };
}

/** Return type for a single mastered item */
export type MasterItem = Awaited<ReturnType<typeof getUserMasters>>["items"][number];

/** Stats counts for the mastered page header */
export async function getUserMasterStats(userId: string) {
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const startOfWeek = new Date(startOfToday);
  startOfWeek.setDate(startOfWeek.getDate() - startOfWeek.getDay());
  const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);

  const [total, thisWeek, thisMonth] = await Promise.all([
    db.userMaster.count({ where: { userId } }),
    db.userMaster.count({ where: { userId, masteredAt: { gte: startOfWeek } } }),
    db.userMaster.count({ where: { userId, masteredAt: { gte: startOfMonth } } }),
  ]);

  return { total, thisWeek, thisMonth };
}

/** Return type for master stats */
export type MasterStats = Awaited<ReturnType<typeof getUserMasterStats>>;
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/models/user-master/user-master.query.ts
git commit -m "feat: add paginated query and stats for user masters"
```

---

### Task 6: Delete Mutations

**Files:**
- Modify: `src/models/user-unknown/user-unknown.mutation.ts`
- Modify: `src/models/user-review/user-review.mutation.ts`
- Modify: `src/models/user-master/user-master.mutation.ts`

**Step 1: Add delete functions to user-unknown.mutation.ts**

Append after existing `upsertUserUnknown` function:

```ts
/** Delete a single unknown entry by ID */
export async function deleteUserUnknown(id: string) {
  return db.userUnknown.delete({
    where: { id },
    select: { id: true },
  });
}

/** Delete multiple unknown entries by IDs */
export async function deleteUserUnknowns(ids: string[]) {
  return db.userUnknown.deleteMany({
    where: { id: { in: ids } },
  });
}
```

**Step 2: Add delete functions to user-review.mutation.ts**

Append after existing `createReviewIfNotExists` function:

```ts
/** Delete a single review entry by ID */
export async function deleteUserReview(id: string) {
  return db.userReview.delete({
    where: { id },
    select: { id: true },
  });
}

/** Delete multiple review entries by IDs */
export async function deleteUserReviews(ids: string[]) {
  return db.userReview.deleteMany({
    where: { id: { in: ids } },
  });
}
```

**Step 3: Add delete functions to user-master.mutation.ts**

Append after existing `upsertUserMaster` function:

```ts
/** Delete a single master entry by ID */
export async function deleteUserMaster(id: string) {
  return db.userMaster.delete({
    where: { id },
    select: { id: true },
  });
}

/** Delete multiple master entries by IDs */
export async function deleteUserMasters(ids: string[]) {
  return db.userMaster.deleteMany({
    where: { id: { in: ids } },
  });
}
```

**Step 4: Verify build**

Run: `npm run build`

**Step 5: Commit**

```bash
git add src/models/user-unknown/user-unknown.mutation.ts src/models/user-review/user-review.mutation.ts src/models/user-master/user-master.mutation.ts
git commit -m "feat: add single and batch delete mutations for unknown, review, master"
```

---

### Task 7: Server Actions — Unknown

**Files:**
- Create: `src/features/web/user-unknown/actions/unknown.action.ts`

**Step 1: Create server action**

```ts
// src/features/web/user-unknown/actions/unknown.action.ts
"use server";

import { auth } from "@/lib/auth";
import { getUserUnknowns } from "@/models/user-unknown/user-unknown.query";
import type { UnknownItem } from "@/models/user-unknown/user-unknown.query";
import { deleteUserUnknown, deleteUserUnknowns } from "@/models/user-unknown/user-unknown.mutation";

/** Fetch next page of unknown words */
export async function fetchUnknownsAction(cursor?: string): Promise<{
  items: UnknownItem[];
  nextCursor: string | null;
}> {
  const session = await auth();
  if (!session?.user?.id) return { items: [], nextCursor: null };
  return getUserUnknowns(session.user.id, cursor);
}

/** Delete a single unknown entry */
export async function deleteUnknownAction(id: string): Promise<{ success: true } | { error: string }> {
  try {
    const session = await auth();
    if (!session?.user?.id) return { error: "未登录" };
    await deleteUserUnknown(id);
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Delete multiple unknown entries */
export async function deleteUnknownsAction(ids: string[]): Promise<{ success: true; count: number } | { error: string }> {
  try {
    const session = await auth();
    if (!session?.user?.id) return { error: "未登录" };
    const result = await deleteUserUnknowns(ids);
    return { success: true, count: result.count };
  } catch {
    return { error: "删除失败" };
  }
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-unknown/actions/unknown.action.ts
git commit -m "feat: add server actions for unknown word list"
```

---

### Task 8: Server Actions — Review

**Files:**
- Create: `src/features/web/user-review/actions/review.action.ts`

**Step 1: Create server action**

```ts
// src/features/web/user-review/actions/review.action.ts
"use server";

import { auth } from "@/lib/auth";
import { getUserReviews } from "@/models/user-review/user-review.query";
import type { ReviewItem } from "@/models/user-review/user-review.query";
import { deleteUserReview, deleteUserReviews } from "@/models/user-review/user-review.mutation";

/** Fetch next page of review words */
export async function fetchReviewsAction(cursor?: string): Promise<{
  items: ReviewItem[];
  nextCursor: string | null;
}> {
  const session = await auth();
  if (!session?.user?.id) return { items: [], nextCursor: null };
  return getUserReviews(session.user.id, cursor);
}

/** Delete a single review entry */
export async function deleteReviewAction(id: string): Promise<{ success: true } | { error: string }> {
  try {
    const session = await auth();
    if (!session?.user?.id) return { error: "未登录" };
    await deleteUserReview(id);
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Delete multiple review entries */
export async function deleteReviewsAction(ids: string[]): Promise<{ success: true; count: number } | { error: string }> {
  try {
    const session = await auth();
    if (!session?.user?.id) return { error: "未登录" };
    const result = await deleteUserReviews(ids);
    return { success: true, count: result.count };
  } catch {
    return { error: "删除失败" };
  }
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-review/actions/review.action.ts
git commit -m "feat: add server actions for review word list"
```

---

### Task 9: Server Actions — Master

**Files:**
- Create: `src/features/web/user-master/actions/master.action.ts`

**Step 1: Create server action**

```ts
// src/features/web/user-master/actions/master.action.ts
"use server";

import { auth } from "@/lib/auth";
import { getUserMasters } from "@/models/user-master/user-master.query";
import type { MasterItem } from "@/models/user-master/user-master.query";
import { deleteUserMaster, deleteUserMasters } from "@/models/user-master/user-master.mutation";

/** Fetch next page of mastered words */
export async function fetchMastersAction(cursor?: string): Promise<{
  items: MasterItem[];
  nextCursor: string | null;
}> {
  const session = await auth();
  if (!session?.user?.id) return { items: [], nextCursor: null };
  return getUserMasters(session.user.id, cursor);
}

/** Delete a single master entry */
export async function deleteMasterAction(id: string): Promise<{ success: true } | { error: string }> {
  try {
    const session = await auth();
    if (!session?.user?.id) return { error: "未登录" };
    await deleteUserMaster(id);
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}

/** Delete multiple master entries */
export async function deleteMastersAction(ids: string[]): Promise<{ success: true; count: number } | { error: string }> {
  try {
    const session = await auth();
    if (!session?.user?.id) return { error: "未登录" };
    const result = await deleteUserMasters(ids);
    return { success: true, count: result.count };
  } catch {
    return { error: "删除失败" };
  }
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-master/actions/master.action.ts
git commit -m "feat: add server actions for master word list"
```

---

### Task 10: Client Hook — useUnknownList

**Files:**
- Create: `src/features/web/user-unknown/hooks/use-unknown-list.ts`

**Step 1: Create hook**

```ts
// src/features/web/user-unknown/hooks/use-unknown-list.ts
"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { toast } from "sonner";
import type { UnknownItem, UnknownStats } from "@/models/user-unknown/user-unknown.query";
import {
  fetchUnknownsAction,
  deleteUnknownAction,
  deleteUnknownsAction,
} from "@/features/web/user-unknown/actions/unknown.action";

interface UseUnknownListParams {
  initialItems: UnknownItem[];
  initialCursor: string | null;
  initialStats: UnknownStats;
}

/** Manages infinite scroll, selection, and delete for unknown word list */
export function useUnknownList({ initialItems, initialCursor, initialStats }: UseUnknownListParams) {
  const [items, setItems] = useState(initialItems);
  const [cursor, setCursor] = useState(initialCursor);
  const [hasMore, setHasMore] = useState(initialCursor !== null);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [stats, setStats] = useState(initialStats);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  /** Sync state when server re-renders with new initial data */
  useEffect(() => {
    setItems(initialItems);
    setCursor(initialCursor);
    setHasMore(initialCursor !== null);
  }, [initialItems, initialCursor]);

  useEffect(() => {
    setStats(initialStats);
  }, [initialStats]);

  /** Load next page of items */
  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore || !cursor) return;
    setIsLoading(true);
    try {
      const result = await fetchUnknownsAction(cursor);
      setItems((prev) => [...prev, ...result.items]);
      setCursor(result.nextCursor);
      setHasMore(result.nextCursor !== null);
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, hasMore, cursor]);

  /** IntersectionObserver for infinite scroll */
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoading) loadMore();
      },
      { rootMargin: "200px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasMore, isLoading, loadMore]);

  /** Delete a single item with optimistic update */
  const deleteOne = useCallback(async (id: string) => {
    const prev = items;
    const prevStats = stats;
    setItems((cur) => cur.filter((i) => i.id !== id));
    setSelectedIds((cur) => { const n = new Set(cur); n.delete(id); return n; });
    setStats((s) => ({ ...s, total: s.total - 1 }));

    const result = await deleteUnknownAction(id);
    if ("error" in result) {
      setItems(prev);
      setStats(prevStats);
      toast.error(result.error);
    } else {
      toast.success("已删除");
    }
  }, [items, stats]);

  /** Delete all selected items with optimistic update */
  const deleteSelected = useCallback(async () => {
    const ids = Array.from(selectedIds);
    if (ids.length === 0) return;

    const prev = items;
    const prevStats = stats;
    const deleteSet = new Set(ids);
    setItems((cur) => cur.filter((i) => !deleteSet.has(i.id)));
    setSelectedIds(new Set());
    setStats((s) => ({ ...s, total: s.total - ids.length }));

    const result = await deleteUnknownsAction(ids);
    if ("error" in result) {
      setItems(prev);
      setStats(prevStats);
      toast.error(result.error);
    } else {
      toast.success(`已删除 ${result.count} 项`);
    }
  }, [selectedIds, items, stats]);

  return {
    items,
    isLoading,
    hasMore,
    sentinelRef,
    selectedIds,
    setSelectedIds,
    stats,
    deleteOne,
    deleteSelected,
  };
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-unknown/hooks/use-unknown-list.ts
git commit -m "feat: add useUnknownList hook with infinite scroll and delete"
```

---

### Task 11: Client Hook — useReviewList

**Files:**
- Create: `src/features/web/user-review/hooks/use-review-list.ts`

**Step 1: Create hook**

Same structure as `useUnknownList` but uses `ReviewItem`, `ReviewStats`, and review action imports.

```ts
// src/features/web/user-review/hooks/use-review-list.ts
"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { toast } from "sonner";
import type { ReviewItem, ReviewStats } from "@/models/user-review/user-review.query";
import {
  fetchReviewsAction,
  deleteReviewAction,
  deleteReviewsAction,
} from "@/features/web/user-review/actions/review.action";

interface UseReviewListParams {
  initialItems: ReviewItem[];
  initialCursor: string | null;
  initialStats: ReviewStats;
}

/** Manages infinite scroll, selection, and delete for review word list */
export function useReviewList({ initialItems, initialCursor, initialStats }: UseReviewListParams) {
  const [items, setItems] = useState(initialItems);
  const [cursor, setCursor] = useState(initialCursor);
  const [hasMore, setHasMore] = useState(initialCursor !== null);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [stats, setStats] = useState(initialStats);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    setItems(initialItems);
    setCursor(initialCursor);
    setHasMore(initialCursor !== null);
  }, [initialItems, initialCursor]);

  useEffect(() => {
    setStats(initialStats);
  }, [initialStats]);

  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore || !cursor) return;
    setIsLoading(true);
    try {
      const result = await fetchReviewsAction(cursor);
      setItems((prev) => [...prev, ...result.items]);
      setCursor(result.nextCursor);
      setHasMore(result.nextCursor !== null);
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, hasMore, cursor]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoading) loadMore();
      },
      { rootMargin: "200px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasMore, isLoading, loadMore]);

  const deleteOne = useCallback(async (id: string) => {
    const prev = items;
    const prevStats = stats;
    setItems((cur) => cur.filter((i) => i.id !== id));
    setSelectedIds((cur) => { const n = new Set(cur); n.delete(id); return n; });
    setStats((s) => ({ ...s, pending: Math.max(0, s.pending - 1) }));

    const result = await deleteReviewAction(id);
    if ("error" in result) {
      setItems(prev);
      setStats(prevStats);
      toast.error(result.error);
    } else {
      toast.success("已删除");
    }
  }, [items, stats]);

  const deleteSelected = useCallback(async () => {
    const ids = Array.from(selectedIds);
    if (ids.length === 0) return;

    const prev = items;
    const prevStats = stats;
    const deleteSet = new Set(ids);
    setItems((cur) => cur.filter((i) => !deleteSet.has(i.id)));
    setSelectedIds(new Set());
    setStats((s) => ({ ...s, pending: Math.max(0, s.pending - ids.length) }));

    const result = await deleteReviewsAction(ids);
    if ("error" in result) {
      setItems(prev);
      setStats(prevStats);
      toast.error(result.error);
    } else {
      toast.success(`已删除 ${result.count} 项`);
    }
  }, [selectedIds, items, stats]);

  return {
    items,
    isLoading,
    hasMore,
    sentinelRef,
    selectedIds,
    setSelectedIds,
    stats,
    deleteOne,
    deleteSelected,
  };
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-review/hooks/use-review-list.ts
git commit -m "feat: add useReviewList hook with infinite scroll and delete"
```

---

### Task 12: Client Hook — useMasterList

**Files:**
- Create: `src/features/web/user-master/hooks/use-master-list.ts`

**Step 1: Create hook**

Same structure but uses `MasterItem`, `MasterStats`, and master action imports.

```ts
// src/features/web/user-master/hooks/use-master-list.ts
"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { toast } from "sonner";
import type { MasterItem, MasterStats } from "@/models/user-master/user-master.query";
import {
  fetchMastersAction,
  deleteMasterAction,
  deleteMastersAction,
} from "@/features/web/user-master/actions/master.action";

interface UseMasterListParams {
  initialItems: MasterItem[];
  initialCursor: string | null;
  initialStats: MasterStats;
}

/** Manages infinite scroll, selection, and delete for master word list */
export function useMasterList({ initialItems, initialCursor, initialStats }: UseMasterListParams) {
  const [items, setItems] = useState(initialItems);
  const [cursor, setCursor] = useState(initialCursor);
  const [hasMore, setHasMore] = useState(initialCursor !== null);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [stats, setStats] = useState(initialStats);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    setItems(initialItems);
    setCursor(initialCursor);
    setHasMore(initialCursor !== null);
  }, [initialItems, initialCursor]);

  useEffect(() => {
    setStats(initialStats);
  }, [initialStats]);

  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore || !cursor) return;
    setIsLoading(true);
    try {
      const result = await fetchMastersAction(cursor);
      setItems((prev) => [...prev, ...result.items]);
      setCursor(result.nextCursor);
      setHasMore(result.nextCursor !== null);
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, hasMore, cursor]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoading) loadMore();
      },
      { rootMargin: "200px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasMore, isLoading, loadMore]);

  const deleteOne = useCallback(async (id: string) => {
    const prev = items;
    const prevStats = stats;
    setItems((cur) => cur.filter((i) => i.id !== id));
    setSelectedIds((cur) => { const n = new Set(cur); n.delete(id); return n; });
    setStats((s) => ({ ...s, total: s.total - 1 }));

    const result = await deleteMasterAction(id);
    if ("error" in result) {
      setItems(prev);
      setStats(prevStats);
      toast.error(result.error);
    } else {
      toast.success("已删除");
    }
  }, [items, stats]);

  const deleteSelected = useCallback(async () => {
    const ids = Array.from(selectedIds);
    if (ids.length === 0) return;

    const prev = items;
    const prevStats = stats;
    const deleteSet = new Set(ids);
    setItems((cur) => cur.filter((i) => !deleteSet.has(i.id)));
    setSelectedIds(new Set());
    setStats((s) => ({ ...s, total: s.total - ids.length }));

    const result = await deleteMastersAction(ids);
    if ("error" in result) {
      setItems(prev);
      setStats(prevStats);
      toast.error(result.error);
    } else {
      toast.success(`已删除 ${result.count} 项`);
    }
  }, [selectedIds, items, stats]);

  return {
    items,
    isLoading,
    hasMore,
    sentinelRef,
    selectedIds,
    setSelectedIds,
    stats,
    deleteOne,
    deleteSelected,
  };
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-master/hooks/use-master-list.ts
git commit -m "feat: add useMasterList hook with infinite scroll and delete"
```

---

### Task 13: Content Component — UnknownContent

**Files:**
- Rewrite: `src/features/web/user-unknown/components/unknown-content.tsx`

**Step 1: Rewrite with real data**

```tsx
// src/features/web/user-unknown/components/unknown-content.tsx
"use client";

import { useState } from "react";
import { BookOpen, Clock, Layers } from "lucide-react";
import { Loader2 } from "lucide-react";
import { StatCard } from "@/components/in/stat-card";
import { WordTable } from "@/components/in/word-table";
import { DeleteConfirmDialog } from "@/components/in/delete-confirm-dialog";
import type { ColumnConfig } from "@/components/in/word-table";
import type { UnknownItem, UnknownStats } from "@/models/user-unknown/user-unknown.query";
import { useUnknownList } from "@/features/web/user-unknown/hooks/use-unknown-list";

interface UnknownContentProps {
  initialItems: UnknownItem[];
  initialCursor: string | null;
  initialStats: UnknownStats;
}

/** Format a date to YYYY-MM-DD */
function formatDate(date: Date | null): string {
  if (!date) return "-";
  return new Date(date).toLocaleDateString("zh-CN");
}

/** Flatten UnknownItem for WordTable compatibility */
type FlatItem = UnknownItem & { content: string; translation: string | null; gameName: string };

function flatten(item: UnknownItem): FlatItem {
  return {
    ...item,
    content: item.contentItem.content,
    translation: item.contentItem.translation,
    gameName: item.game.name,
  };
}

const columns: ColumnConfig<FlatItem>[] = [
  { key: "content", label: "生词", width: "flex-1", render: () => null },
  {
    key: "gameName",
    label: "来源",
    width: "w-[140px]",
    render: (item) => <span className="text-xs text-slate-500">{item.gameName}</span>,
  },
  {
    key: "createdAt",
    label: "添加时间",
    width: "w-[120px]",
    render: (item) => <span className="text-xs text-slate-400">{formatDate(item.createdAt)}</span>,
  },
];

export function UnknownContent({ initialItems, initialCursor, initialStats }: UnknownContentProps) {
  const {
    items, isLoading, hasMore, sentinelRef,
    selectedIds, setSelectedIds, stats, deleteOne, deleteSelected,
  } = useUnknownList({ initialItems, initialCursor, initialStats });

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [showBatchDelete, setShowBatchDelete] = useState(false);

  const flatItems = items.map(flatten);

  const statCards = [
    { icon: BookOpen, iconBg: "bg-red-100", iconColor: "text-red-500", value: String(stats.total), label: "全部生词" },
    { icon: Clock, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(stats.today), label: "今日添加" },
    { icon: Layers, iconBg: "bg-blue-100", iconColor: "text-blue-500", value: String(stats.lastThreeDays), label: "最近三天" },
  ];

  return (
    <>
      <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
        {statCards.map((stat) => (
          <StatCard key={stat.label} {...stat} />
        ))}
      </div>

      <WordTable
        items={flatItems}
        columns={columns}
        selectedIds={selectedIds}
        onSelectChange={setSelectedIds}
        onDelete={(id) => setDeleteTarget(id)}
        onDeleteBatch={() => setShowBatchDelete(true)}
      />

      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      )}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      {/* Single delete dialog */}
      <DeleteConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null); }}
        count={1}
        onConfirm={() => { if (deleteTarget) { deleteOne(deleteTarget); setDeleteTarget(null); } }}
      />

      {/* Batch delete dialog */}
      <DeleteConfirmDialog
        open={showBatchDelete}
        onOpenChange={setShowBatchDelete}
        count={selectedIds.size}
        onConfirm={() => { deleteSelected(); setShowBatchDelete(false); }}
      />
    </>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-unknown/components/unknown-content.tsx
git commit -m "feat: rewrite UnknownContent with real data and infinite scroll"
```

---

### Task 14: Content Component — ReviewContent

**Files:**
- Rewrite: `src/features/web/user-review/components/review-content.tsx`

**Step 1: Rewrite with real data**

```tsx
// src/features/web/user-review/components/review-content.tsx
"use client";

import { useState } from "react";
import { Clock, AlertTriangle, CheckCircle2, Loader2 } from "lucide-react";
import Link from "next/link";
import { StatCard } from "@/components/in/stat-card";
import { WordTable } from "@/components/in/word-table";
import { Badge } from "@/components/in/badge";
import { DeleteConfirmDialog } from "@/components/in/delete-confirm-dialog";
import type { ColumnConfig } from "@/components/in/word-table";
import type { ReviewItem, ReviewStats } from "@/models/user-review/user-review.query";
import { useReviewList } from "@/features/web/user-review/hooks/use-review-list";

interface ReviewContentProps {
  initialItems: ReviewItem[];
  initialCursor: string | null;
  initialStats: ReviewStats;
}

/** Format a date to YYYY-MM-DD */
function formatDate(date: Date | null): string {
  if (!date) return "-";
  return new Date(date).toLocaleDateString("zh-CN");
}

/** Derive urgency level from nextReviewAt */
function getUrgency(nextReviewAt: Date | null): { label: string; bg: string; text: string } {
  if (!nextReviewAt) return { label: "正常", bg: "bg-blue-100", text: "text-blue-700" };
  const now = new Date();
  const diff = new Date(nextReviewAt).getTime() - now.getTime();
  const days = diff / (1000 * 60 * 60 * 24);
  if (days < 0) return { label: "紧急", bg: "bg-red-100", text: "text-red-700" };
  if (days < 1) return { label: "较高", bg: "bg-red-100", text: "text-red-700" };
  if (days < 3) return { label: "中等", bg: "bg-amber-100", text: "text-amber-700" };
  return { label: "较低", bg: "bg-teal-100", text: "text-teal-700" };
}

type FlatItem = ReviewItem & {
  content: string;
  translation: string | null;
  gameId: string;
  gameName: string;
};

function flatten(item: ReviewItem): FlatItem {
  return {
    ...item,
    content: item.contentItem.content,
    translation: item.contentItem.translation,
    gameId: item.game.id,
    gameName: item.game.name,
  };
}

const columns: ColumnConfig<FlatItem>[] = [
  { key: "content", label: "词汇", width: "flex-1", render: () => null },
  {
    key: "lastReview",
    label: "上次复习",
    width: "w-[100px]",
    render: (item) => <span className="text-xs text-slate-400">{formatDate(item.lastReviewAt)}</span>,
  },
  {
    key: "nextReview",
    label: "下次复习",
    width: "w-[100px]",
    render: (item) => {
      const isOverdue = item.nextReviewAt && new Date(item.nextReviewAt) < new Date();
      return (
        <span className={`text-xs ${isOverdue ? "font-semibold text-red-500" : "text-slate-400"}`}>
          {formatDate(item.nextReviewAt)}
        </span>
      );
    },
  },
  {
    key: "urgency",
    label: "紧急度",
    width: "w-[80px]",
    render: (item) => {
      const u = getUrgency(item.nextReviewAt);
      return <Badge label={u.label} bg={u.bg} text={u.text} />;
    },
  },
  {
    key: "action",
    label: "",
    width: "w-[72px]",
    render: (item) => (
      <Link
        href={`/hall/games/${item.gameId}`}
        className="rounded-md bg-indigo-50 px-2.5 py-1 text-[11px] font-semibold text-indigo-600 hover:bg-indigo-100"
      >
        去复习
      </Link>
    ),
  },
];

export function ReviewContent({ initialItems, initialCursor, initialStats }: ReviewContentProps) {
  const {
    items, isLoading, hasMore, sentinelRef,
    selectedIds, setSelectedIds, stats, deleteOne, deleteSelected,
  } = useReviewList({ initialItems, initialCursor, initialStats });

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [showBatchDelete, setShowBatchDelete] = useState(false);

  const flatItems = items.map(flatten);

  const statCards = [
    { icon: Clock, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(stats.pending), label: "待复习" },
    { icon: AlertTriangle, iconBg: "bg-red-100", iconColor: "text-red-500", value: String(stats.overdue), label: "已逾期" },
    { icon: CheckCircle2, iconBg: "bg-teal-100", iconColor: "text-teal-600", value: String(stats.reviewedToday), label: "今日已复习" },
  ];

  return (
    <>
      <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
        {statCards.map((stat) => (
          <StatCard key={stat.label} {...stat} />
        ))}
      </div>

      <WordTable
        items={flatItems}
        columns={columns}
        selectedIds={selectedIds}
        onSelectChange={setSelectedIds}
        onDelete={(id) => setDeleteTarget(id)}
        onDeleteBatch={() => setShowBatchDelete(true)}
      />

      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      )}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      <DeleteConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null); }}
        count={1}
        onConfirm={() => { if (deleteTarget) { deleteOne(deleteTarget); setDeleteTarget(null); } }}
      />
      <DeleteConfirmDialog
        open={showBatchDelete}
        onOpenChange={setShowBatchDelete}
        count={selectedIds.size}
        onConfirm={() => { deleteSelected(); setShowBatchDelete(false); }}
      />
    </>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-review/components/review-content.tsx
git commit -m "feat: rewrite ReviewContent with real data, infinite scroll, and review link"
```

---

### Task 15: Content Component — MasterContent

**Files:**
- Delete: `src/features/web/user-mastered/components/mastered-content.tsx`
- Create: `src/features/web/user-master/components/master-content.tsx`

Note: The feature directory changes from `user-mastered` to `user-master`.

**Step 1: Create master-content.tsx**

```tsx
// src/features/web/user-master/components/master-content.tsx
"use client";

import { useState } from "react";
import { CheckCircle2, CalendarDays, CalendarRange, Loader2 } from "lucide-react";
import { StatCard } from "@/components/in/stat-card";
import { WordTable } from "@/components/in/word-table";
import { DeleteConfirmDialog } from "@/components/in/delete-confirm-dialog";
import type { ColumnConfig } from "@/components/in/word-table";
import type { MasterItem, MasterStats } from "@/models/user-master/user-master.query";
import { useMasterList } from "@/features/web/user-master/hooks/use-master-list";

interface MasterContentProps {
  initialItems: MasterItem[];
  initialCursor: string | null;
  initialStats: MasterStats;
}

/** Format a date to YYYY-MM-DD */
function formatDate(date: Date | null): string {
  if (!date) return "-";
  return new Date(date).toLocaleDateString("zh-CN");
}

type FlatItem = MasterItem & { content: string; translation: string | null; gameName: string };

function flatten(item: MasterItem): FlatItem {
  return {
    ...item,
    content: item.contentItem.content,
    translation: item.contentItem.translation,
    gameName: item.game.name,
  };
}

const columns: ColumnConfig<FlatItem>[] = [
  { key: "content", label: "内容", width: "flex-1", render: () => null },
  {
    key: "gameName",
    label: "来源",
    width: "w-[140px]",
    render: (item) => <span className="text-xs text-slate-500">{item.gameName}</span>,
  },
  {
    key: "masteredAt",
    label: "掌握时间",
    width: "w-[120px]",
    render: (item) => <span className="text-xs text-slate-400">{formatDate(item.masteredAt)}</span>,
  },
];

export function MasterContent({ initialItems, initialCursor, initialStats }: MasterContentProps) {
  const {
    items, isLoading, hasMore, sentinelRef,
    selectedIds, setSelectedIds, stats, deleteOne, deleteSelected,
  } = useMasterList({ initialItems, initialCursor, initialStats });

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [showBatchDelete, setShowBatchDelete] = useState(false);

  const flatItems = items.map(flatten);

  const statCards = [
    { icon: CheckCircle2, iconBg: "bg-teal-100", iconColor: "text-teal-600", value: String(stats.total), label: "已掌握总数" },
    { icon: CalendarDays, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(stats.thisWeek), label: "本周掌握" },
    { icon: CalendarRange, iconBg: "bg-blue-100", iconColor: "text-blue-500", value: String(stats.thisMonth), label: "本月掌握" },
  ];

  return (
    <>
      <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
        {statCards.map((stat) => (
          <StatCard key={stat.label} {...stat} />
        ))}
      </div>

      <WordTable
        items={flatItems}
        columns={columns}
        selectedIds={selectedIds}
        onSelectChange={setSelectedIds}
        onDelete={(id) => setDeleteTarget(id)}
        onDeleteBatch={() => setShowBatchDelete(true)}
      />

      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      )}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      <DeleteConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null); }}
        count={1}
        onConfirm={() => { if (deleteTarget) { deleteOne(deleteTarget); setDeleteTarget(null); } }}
      />
      <DeleteConfirmDialog
        open={showBatchDelete}
        onOpenChange={setShowBatchDelete}
        count={selectedIds.size}
        onConfirm={() => { deleteSelected(); setShowBatchDelete(false); }}
      />
    </>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```bash
git add src/features/web/user-master/components/master-content.tsx
git commit -m "feat: add MasterContent with real data and infinite scroll"
```

---

### Task 16: Page Components — Wire Server Data

**Files:**
- Modify: `src/app/(web)/hall/(main)/unknown/page.tsx`
- Modify: `src/app/(web)/hall/(main)/review/page.tsx`
- Modify: `src/app/(web)/hall/(main)/mastered/page.tsx`

**Step 1: Rewrite unknown/page.tsx**

```tsx
// src/app/(web)/hall/(main)/unknown/page.tsx
import { auth } from "@/lib/auth";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { UnknownContent } from "@/features/web/user-unknown/components/unknown-content";
import { getUserUnknowns, getUserUnknownStats } from "@/models/user-unknown/user-unknown.query";

export default async function UnknownPage() {
  const session = await auth();
  const userId = session?.user?.id;

  const [data, stats] = userId
    ? await Promise.all([getUserUnknowns(userId), getUserUnknownStats(userId)])
    : [{ items: [], nextCursor: null }, { total: 0, today: 0, lastThreeDays: 0 }];

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="生词本"
        subtitle="记录你遇到的新单词和生词"
        searchPlaceholder="搜索生词..."
      />
      <UnknownContent
        initialItems={data.items}
        initialCursor={data.nextCursor}
        initialStats={stats}
      />
    </div>
  );
}
```

**Step 2: Rewrite review/page.tsx**

```tsx
// src/app/(web)/hall/(main)/review/page.tsx
import { auth } from "@/lib/auth";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { ReviewContent } from "@/features/web/user-review/components/review-content";
import { getUserReviews, getUserReviewStats } from "@/models/user-review/user-review.query";

export default async function ReviewPage() {
  const session = await auth();
  const userId = session?.user?.id;

  const [data, stats] = userId
    ? await Promise.all([getUserReviews(userId), getUserReviewStats(userId)])
    : [{ items: [], nextCursor: null }, { pending: 0, overdue: 0, reviewedToday: 0 }];

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="复习本"
        subtitle="需要复习巩固的词汇和知识点"
        searchPlaceholder="搜索复习内容..."
      />
      <ReviewContent
        initialItems={data.items}
        initialCursor={data.nextCursor}
        initialStats={stats}
      />
    </div>
  );
}
```

**Step 3: Rewrite mastered/page.tsx**

```tsx
// src/app/(web)/hall/(main)/mastered/page.tsx
import { auth } from "@/lib/auth";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { MasterContent } from "@/features/web/user-master/components/master-content";
import { getUserMasters, getUserMasterStats } from "@/models/user-master/user-master.query";

export default async function MasterPage() {
  const session = await auth();
  const userId = session?.user?.id;

  const [data, stats] = userId
    ? await Promise.all([getUserMasters(userId), getUserMasterStats(userId)])
    : [{ items: [], nextCursor: null }, { total: 0, thisWeek: 0, thisMonth: 0 }];

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="已掌握"
        subtitle="你已经掌握的词汇和知识点"
        searchPlaceholder="搜索已掌握内容..."
      />
      <MasterContent
        initialItems={data.items}
        initialCursor={data.nextCursor}
        initialStats={stats}
      />
    </div>
  );
}
```

**Step 4: Verify build**

Run: `npm run build`

**Step 5: Commit**

```bash
git add src/app/(web)/hall/(main)/unknown/page.tsx src/app/(web)/hall/(main)/review/page.tsx src/app/(web)/hall/(main)/mastered/page.tsx
git commit -m "feat: wire real data to unknown, review, and master pages"
```

---

### Task 17: Cleanup — Remove Old Feature-Specific Copies

**Files:**
- Delete: `src/features/web/user-unknown/components/word-table.tsx`
- Delete: `src/features/web/user-unknown/components/stat-card.tsx`
- Delete: `src/features/web/user-review/components/word-table.tsx`
- Delete: `src/features/web/user-review/components/stat-card.tsx`
- Delete: `src/features/web/user-mastered/components/word-table.tsx`
- Delete: `src/features/web/user-mastered/components/stat-card.tsx`
- Delete: `src/features/web/user-mastered/components/mastered-content.tsx`

**Step 1: Delete old files**

```bash
rm src/features/web/user-unknown/components/word-table.tsx
rm src/features/web/user-unknown/components/stat-card.tsx
rm src/features/web/user-review/components/word-table.tsx
rm src/features/web/user-review/components/stat-card.tsx
rm src/features/web/user-mastered/components/word-table.tsx
rm src/features/web/user-mastered/components/stat-card.tsx
rm src/features/web/user-mastered/components/mastered-content.tsx
```

**Step 2: Verify no remaining imports to deleted files**

Run: `grep -r "user-unknown/components/word-table\|user-unknown/components/stat-card\|user-review/components/word-table\|user-review/components/stat-card\|user-mastered/components" src/ --include="*.tsx" --include="*.ts"`

Should return no matches.

**Step 3: Verify build**

Run: `npm run build`

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove old feature-specific WordTable and StatCard copies"
```

---

### Task 18: Final Verification

**Step 1: Full build**

Run: `npm run build`

**Step 2: Manual smoke test**

Run: `npm run dev`

Verify in browser:
- `/hall/unknown` — loads real data, infinite scroll works, single delete with AlertDialog, batch select + delete with AlertDialog, stat cards show real counts
- `/hall/review` — loads real data sorted by urgency, overdue dates in red, 去复习 links to `/hall/games/{id}`, delete works
- `/hall/mastered` — loads real data, all "mastered" references renamed to "master" in code, delete works
- All pages show "暂无数据" when list is empty
