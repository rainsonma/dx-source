import { apiClient } from "@/lib/api-client";
import { leaderboardParamsSchema } from "../schemas/leaderboard.schema";
import type { LeaderboardResult } from "../types/leaderboard.types";

type FetchLeaderboardResult = LeaderboardResult | { error: string };

/** Fetch leaderboard data for the given type and period */
export async function fetchLeaderboardAction(
  type: string,
  period: string
): Promise<FetchLeaderboardResult> {
  const parsed = leaderboardParamsSchema.safeParse({ type, period });
  if (!parsed.success) return { error: "参数无效" };

  try {
    const res = await apiClient.get<LeaderboardResult>(
      `/api/leaderboard?type=${parsed.data.type}&period=${parsed.data.period}`
    );

    if (res.code !== 0) {
      return { error: res.message };
    }

    return res.data;
  } catch {
    return { error: "获取排行榜失败，请重试" };
  }
}
