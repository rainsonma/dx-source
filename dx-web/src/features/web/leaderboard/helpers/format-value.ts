import { formatPlayTime } from "@/lib/format";
import type { LeaderboardType } from "../types/leaderboard.types";

/** Format a leaderboard value based on type (EXP as number, play time as duration) */
export function formatLeaderboardValue(
  value: number,
  type: LeaderboardType
): string {
  if (type === "exp") return value.toLocaleString("zh-CN");
  return formatPlayTime(value);
}
