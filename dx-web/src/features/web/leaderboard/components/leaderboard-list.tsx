import { Zap, Clock } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { formatLeaderboardValue } from "../helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "../types/leaderboard.types";

interface LeaderboardListProps {
  entries: LeaderboardEntry[];
  type: LeaderboardType;
}

/** Rank list for positions #4 and beyond */
export function LeaderboardList({ entries, type }: LeaderboardListProps) {
  const Icon = type === "exp" ? Zap : Clock;

  if (entries.length === 0) return null;

  return (
    <>
      {entries.map((entry) => {
        const displayName = entry.nickname ?? entry.username;
        const fallbackChar = displayName.charAt(0).toUpperCase();
        const avatarBg = getAvatarColor(entry.id);

        return (
          <div
            key={entry.id}
            className="flex items-center gap-4 border-b border-border px-4 py-3 last:border-b-0 md:px-6"
          >
            <span className="w-6 text-sm font-bold text-muted-foreground">
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
            <span className="flex-1 text-[13px] font-semibold text-foreground">
              {displayName}
            </span>
            <div className="flex items-center gap-1">
              <Icon className="h-3.5 w-3.5 text-amber-500" />
              <span className="text-sm font-semibold text-foreground">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
          </div>
        );
      })}
    </>
  );
}
