import { Zap, Clock } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { formatLeaderboardValue } from "@/features/web/leaderboard/helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "@/features/web/leaderboard/types/leaderboard.types";

interface TodayStarsMyRankProps {
  entry: LeaderboardEntry | null;
  type: LeaderboardType;
  user: { id: string; username: string; nickname: string | null; avatarUrl: string | null };
}

/** My rank bar */
export function TodayStarsMyRank({ entry, type, user }: TodayStarsMyRankProps) {
  const displayName = user.nickname ?? user.username;
  const rank = entry?.rank ?? null;
  const value = entry ? formatLeaderboardValue(entry.value, type) : "0";
  const Icon = type === "exp" ? Zap : Clock;
  const fallbackChar = displayName.charAt(0).toUpperCase();
  const avatarBg = getAvatarColor(user.id);

  return (
    <div className="flex items-center gap-3 rounded-full border-[1.5px] border-teal-600 bg-teal-50 px-3 py-2">
      {rank !== null && (
        <span className="text-sm font-bold text-teal-600">{rank}</span>
      )}
      <Avatar size="sm">
        {user.avatarUrl && (
          <AvatarImage src={user.avatarUrl} alt={displayName} />
        )}
        <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>
          {fallbackChar}
        </AvatarFallback>
      </Avatar>
      <span className="flex-1 truncate text-xs font-semibold text-foreground">
        {displayName}
      </span>
      <div className="flex items-center gap-0.5">
        <Icon className="h-3 w-3 text-amber-500" />
        <span className="text-xs font-semibold text-foreground">{value}</span>
      </div>
    </div>
  );
}
