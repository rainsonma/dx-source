import { Zap, Clock } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { formatLeaderboardValue } from "@/features/web/leaderboard/helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "@/features/web/leaderboard/types/leaderboard.types";

interface TodayStarsListProps {
  entries: LeaderboardEntry[];
  type: LeaderboardType;
}

/** Scrollable list for rank 4-50 */
export function TodayStarsList({ entries, type }: TodayStarsListProps) {
  const Icon = type === "exp" ? Zap : Clock;

  if (entries.length === 0) return null;

  return (
    <div className="max-h-[280px] overflow-y-auto">
      {entries.map((entry) => {
        const displayName = entry.nickname ?? entry.username;
        const fallbackChar = displayName.charAt(0).toUpperCase();
        const avatarBg = getAvatarColor(entry.id);

        return (
          <div
            key={entry.id}
            className="flex items-center gap-3 border-b border-border px-4 py-2.5 last:border-b-0"
          >
            <span className="w-5 text-xs font-bold text-muted-foreground">
              {entry.rank}
            </span>
            <Avatar size="sm">
              {entry.avatarUrl && (
                <AvatarImage src={entry.avatarUrl} alt={displayName} />
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
              <span className="text-xs font-semibold text-foreground">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
