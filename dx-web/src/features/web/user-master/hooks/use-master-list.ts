"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { toast } from "sonner";
import type { MasterItem, MasterStats } from "@/features/web/user-master/actions/master.action";
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
      const result = await fetchMastersAction(cursor);
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

    const result = await deleteMasterAction(id);
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
