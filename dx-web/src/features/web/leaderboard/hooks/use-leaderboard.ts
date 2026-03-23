"use client";

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { fetchLeaderboardAction } from "../actions/leaderboard.action";
import type {
  LeaderboardType,
  LeaderboardPeriod,
  LeaderboardResult,
} from "../types/leaderboard.types";

interface UseLeaderboardParams {
  initialData: LeaderboardResult;
}

/** Manage leaderboard tab state and data fetching */
export function useLeaderboard({ initialData }: UseLeaderboardParams) {
  const [type, setType] = useState<LeaderboardType>("exp");
  const [period, setPeriod] = useState<LeaderboardPeriod>("all");
  const [data, setData] = useState<LeaderboardResult>(initialData);
  const [isLoading, setIsLoading] = useState(false);

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
