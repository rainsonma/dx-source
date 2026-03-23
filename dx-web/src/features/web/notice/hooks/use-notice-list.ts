"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { toast } from "sonner";
import {
  type NoticeItem,
  fetchNoticesAction,
  createNoticeAction,
  updateNoticeAction,
  deleteNoticeAction,
} from "@/features/web/notice/actions/notice.action";

interface UseNoticeListParams {
  initialItems: NoticeItem[];
  initialCursor: string | null;
}

/** Manages infinite scroll and publish for notice list */
export function useNoticeList({ initialItems, initialCursor }: UseNoticeListParams) {
  const [items, setItems] = useState<NoticeItem[]>(initialItems);
  const [cursor, setCursor] = useState(initialCursor);
  const [hasMore, setHasMore] = useState(initialCursor !== null);
  const [isLoading, setIsLoading] = useState(false);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  /** Sync state when server re-renders with new initial data */
  useEffect(() => {
    setItems(initialItems);
    setCursor(initialCursor);
    setHasMore(initialCursor !== null);
  }, [initialItems, initialCursor]);

  /** Load next page of notices */
  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore || !cursor) return;
    setIsLoading(true);
    try {
      const result = await fetchNoticesAction(cursor);
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

  /** Publish a new notice and prepend to list */
  const publishNotice = useCallback(
    async (input: { title: string; content?: string; icon?: string }) => {
      const result = await createNoticeAction(input);
      if ("error" in result) {
        toast.error(result.error);
        return false;
      }
      setItems((prev) => [result.data, ...prev]);
      toast.success("通知已发布");
      return true;
    },
    []
  );

  /** Update an existing notice in the list */
  const editNotice = useCallback(
    async (input: { id: string; title: string; content?: string; icon?: string }) => {
      const result = await updateNoticeAction(input);
      if ("error" in result) {
        toast.error(result.error);
        return false;
      }
      setItems((prev) =>
        prev.map((item) => (item.id === input.id ? result.data : item))
      );
      toast.success("通知已更新");
      return true;
    },
    []
  );

  /** Remove a notice from the list (soft delete) */
  const removeNotice = useCallback(
    async (id: string) => {
      const result = await deleteNoticeAction(id);
      if ("error" in result) {
        toast.error(result.error);
        return false;
      }
      setItems((prev) => prev.filter((item) => item.id !== id));
      toast.success("通知已删除");
      return true;
    },
    []
  );

  return {
    items,
    isLoading,
    hasMore,
    sentinelRef,
    publishNotice,
    editNotice,
    removeNotice,
  };
}
