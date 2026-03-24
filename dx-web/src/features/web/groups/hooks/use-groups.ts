"use client";

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { groupApi } from "../actions/group.action";
import type { Group } from "../types/group";

type Tab = "all" | "created" | "joined";

export function useGroups() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [tab, setTab] = useState<Tab>("all");
  const [nextCursor, setNextCursor] = useState("");
  const [hasMore, setHasMore] = useState(false);

  const fetchGroups = useCallback(async (selectedTab?: Tab, cursor?: string) => {
    setIsLoading(true);
    try {
      const res = await groupApi.list({ tab: selectedTab || tab, cursor });
      if (res.code !== 0) {
        toast.error(res.message);
        return;
      }
      if (cursor) {
        setGroups((prev) => [...prev, ...res.data.items]);
      } else {
        setGroups(res.data.items);
      }
      setNextCursor(res.data.nextCursor);
      setHasMore(res.data.hasMore);
    } catch {
      toast.error("加载失败");
    } finally {
      setIsLoading(false);
    }
  }, [tab]);

  const changeTab = useCallback((newTab: Tab) => {
    setTab(newTab);
    setGroups([]);
    setNextCursor("");
    fetchGroups(newTab);
  }, [fetchGroups]);

  const loadMore = useCallback(() => {
    if (hasMore && nextCursor) {
      fetchGroups(tab, nextCursor);
    }
  }, [hasMore, nextCursor, tab, fetchGroups]);

  return { groups, isLoading, tab, hasMore, fetchGroups, changeTab, loadMore };
}
