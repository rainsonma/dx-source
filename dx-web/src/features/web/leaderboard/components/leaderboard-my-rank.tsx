import { Zap, Clock } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { formatLeaderboardValue } from "../helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "../types/leaderboard.types";

interface LeaderboardMyRankProps {
  entry: LeaderboardEntry | null;
  type: LeaderboardType;
}

/** Current user's rank bar — always visible at the top */
export function LeaderboardMyRank({ entry, type }: LeaderboardMyRankProps) {
  const displayName = entry ? (entry.nickname ?? entry.username) : "我";
  const rank = entry?.rank ?? null;
  const value = entry ? formatLeaderboardValue(entry.value, type) : "0";
  const Icon = type === "exp" ? Zap : Clock;
  const fallbackChar = displayName.charAt(0).toUpperCase();
  const avatarBg = entry ? getAvatarColor(entry.id) : "#94a3b8";

  return (
    <div className="flex w-full items-center gap-4 rounded-full border-[1.5px] border-teal-600 bg-teal-50 px-4 py-3.5 md:px-6">
      {rank !== null && (
        <span className="text-base font-bold text-teal-600">{rank}</span>
      )}
      <Avatar>
        {entry?.avatarUrl && (
          <AvatarImage src={entry.avatarUrl} alt={displayName} />
        )}
        <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>
          {fallbackChar}
        </AvatarFallback>
      </Avatar>
      <div className="flex flex-1 items-center gap-1.5">
        <span className="text-sm font-semibold text-foreground">
          {displayName}
        </span>
      </div>
      <span className="hidden rounded-xl bg-teal-600 px-3 py-1 text-[11px] font-semibold text-white sm:inline">
        {rank !== null ? "我的排名" : "未上榜"}
      </span>
      <div className="flex items-center gap-1">
        <Icon className="h-3.5 w-3.5 text-amber-500" />
        <span className="text-sm font-semibold text-foreground">{value}</span>
      </div>
    </div>
  );
}
