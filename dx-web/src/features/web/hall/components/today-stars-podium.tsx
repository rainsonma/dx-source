import { Crown, Zap, Clock } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { formatLeaderboardValue } from "@/features/web/leaderboard/helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "@/features/web/leaderboard/types/leaderboard.types";

const PODIUM_SLOTS = [
  { index: 1, color: "bg-slate-300", height: "h-16" },
  { index: 0, color: "bg-amber-400", height: "h-20" },
  { index: 2, color: "bg-amber-600", height: "h-12" },
];

interface TodayStarsPodiumProps {
  entries: LeaderboardEntry[];
  type: LeaderboardType;
}

/** Mini podium for top 3 users */
export function TodayStarsPodium({ entries, type }: TodayStarsPodiumProps) {
  const Icon = type === "exp" ? Zap : Clock;

  return (
    <div className="flex items-end justify-center gap-3 rounded-lg bg-gradient-to-b from-teal-50 to-white px-4 pb-0 pt-4">
      {PODIUM_SLOTS.map(({ index, color, height }) => {
        const entry = entries[index];
        if (!entry) return null;
        const displayName = entry.nickname ?? entry.username;
        const fallbackChar = displayName.charAt(0).toUpperCase();
        const avatarBg = getAvatarColor(entry.id);

        return (
          <div key={entry.id} className="flex flex-col items-center gap-1">
            {entry.rank === 1 && (
              <Crown className="h-4 w-4 text-amber-400" />
            )}
            <Avatar size="sm">
              {entry.avatarUrl && (
                <AvatarImage src={entry.avatarUrl} alt={displayName} />
              )}
              <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>
                {fallbackChar}
              </AvatarFallback>
            </Avatar>
            <span className="max-w-[72px] truncate text-xs font-semibold text-foreground">
              {displayName}
            </span>
            <div className="flex items-center gap-0.5">
              <Icon className="h-2.5 w-2.5 text-amber-500" />
              <span className="text-[10px] font-semibold text-muted-foreground">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
            <div
              className={`w-16 ${height} rounded-t-md ${color} flex items-center justify-center text-base font-extrabold text-white`}
            >
              {entry.rank}
            </div>
          </div>
        );
      })}
    </div>
  );
}
