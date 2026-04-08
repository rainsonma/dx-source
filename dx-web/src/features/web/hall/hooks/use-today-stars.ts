"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { fetchLeaderboardAction } from "@/features/web/leaderboard/actions/leaderboard.action";
import type {
  LeaderboardType,
  LeaderboardResult,
} from "@/features/web/leaderboard/types/leaderboard.types";

const MAX_ENTRIES = 50;

/** Fetch today's leaderboard (day period only), type-switchable, sliced to 50 */
export function useTodayStars() {
  const [type, setType] = useState<LeaderboardType>("exp");
  const [data, setData] = useState<LeaderboardResult>({ entries: [], myRank: null });
  const [isLoading, setIsLoading] = useState(true);

  const fetchData = useCallback(async (t: LeaderboardType) => {
    setIsLoading(true);
    try {
      const result = await fetchLeaderboardAction(t, "day");
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
    fetchData("exp");
  }, [fetchData]);

  const handleTypeChange = useCallback(
    (newType: LeaderboardType) => {
      if (newType === type) return;
      setType(newType);
      fetchData(newType);
    },
    [type, fetchData]
  );

  return { type, data, isLoading, handleTypeChange };
}
