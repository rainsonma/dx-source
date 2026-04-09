"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { fetchLeaderboardAction } from "@/features/web/leaderboard/actions/leaderboard.action";
import type { LeaderboardResult } from "@/features/web/leaderboard/types/leaderboard.types";

const MAX_ENTRIES = 50;

/** Fetch today's leaderboard — playTime, day period, sliced to 50 */
export function useTodayStars() {
  const [data, setData] = useState<LeaderboardResult>({ entries: [], myRank: null });
  const [isLoading, setIsLoading] = useState(true);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await fetchLeaderboardAction("playtime", "day");
      if ("error" in result) {
        toast.error(result.error);
        return;
      }
      setData({
        entries: result.entries.slice(0, MAX_ENTRIES),
        myRank: result.myRank,
      });
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return { data, isLoading };
}
