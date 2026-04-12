"use client";

import { useState, useCallback, useEffect } from "react";
import { toast } from "sonner";
import { fetchLeaderboardAction } from "../actions/leaderboard.action";
import type {
  LeaderboardType,
  LeaderboardPeriod,
  LeaderboardResult,
} from "../types/leaderboard.types";

const EMPTY_DATA: LeaderboardResult = { entries: [], myRank: null };

/** Manage leaderboard tab state and data fetching */
export function useLeaderboard() {
  const [type, setType] = useState<LeaderboardType>("playtime");
  const [period, setPeriod] = useState<LeaderboardPeriod>("month");
  const [data, setData] = useState<LeaderboardResult>(EMPTY_DATA);
  const [isLoading, setIsLoading] = useState(true);

  /** Fetch leaderboard data for a type+period combination */
  const fetchData = useCallback(
    async (newType: LeaderboardType, newPeriod: LeaderboardPeriod) => {
      setIsLoading(true);
      try {
        const result = await fetchLeaderboardAction(newType, newPeriod);
        if ("error" in result) {
          toast.error(result.error);
          return;
        }
        setData(result);
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  // Fetch initial data on mount for the default tab
  useEffect(() => {
    fetchData(type, period);
    // eslint-disable-next-line react-hooks/exhaustive-deps -- run once on mount
  }, []);

  /** Switch the leaderboard type tab */
  const handleTypeChange = useCallback(
    (newType: LeaderboardType) => {
      if (newType === type) return;
      setType(newType);
      fetchData(newType, period);
    },
    [type, period, fetchData]
  );

  /** Switch the leaderboard period tab */
  const handlePeriodChange = useCallback(
    (newPeriod: LeaderboardPeriod) => {
      if (newPeriod === period) return;
      setPeriod(newPeriod);
      fetchData(type, newPeriod);
    },
    [type, period, fetchData]
  );

  return { type, period, data, isLoading, handleTypeChange, handlePeriodChange };
}
