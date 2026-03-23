import { Crown, Zap, Clock } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { formatLeaderboardValue } from "../helpers/format-value";
import type { LeaderboardEntry, LeaderboardType } from "../types/leaderboard.types";

const PODIUM_SLOTS = [
  { index: 1, color: "bg-slate-300", height: "h-24" },
  { index: 0, color: "bg-amber-400", height: "h-32" },
  { index: 2, color: "bg-amber-600", height: "h-20" },
];

interface LeaderboardPodiumProps {
  entries: LeaderboardEntry[];
  type: LeaderboardType;
}

/** Top-3 podium display with medal heights */
export function LeaderboardPodium({ entries, type }: LeaderboardPodiumProps) {
  const Icon = type === "exp" ? Zap : Clock;

  return (
    <div className="flex flex-col items-center gap-4 bg-gradient-to-b from-teal-50 to-white px-6 pb-0 pt-6 sm:flex-row sm:items-end sm:justify-center sm:gap-6 sm:px-10">
      {PODIUM_SLOTS.map(({ index, color, height }) => {
        const entry = entries[index];
        if (!entry) return null;
        const displayName = entry.nickname ?? entry.username;
        const fallbackChar = displayName.charAt(0).toUpperCase();
        const avatarBg = getAvatarColor(entry.id);

        return (
          <div key={entry.id} className="flex flex-col items-center gap-1.5">
            {entry.rank === 1 && (
              <Crown className="h-5 w-5 text-amber-400" />
            )}
            <Avatar size="lg">
              {entry.avatarUrl && (
                <AvatarImage src={entry.avatarUrl} alt={displayName} />
              )}
              <AvatarFallback style={{ backgroundColor: avatarBg, color: "#fff" }}>
                {fallbackChar}
              </AvatarFallback>
            </Avatar>
            <span className="text-[13px] font-semibold text-foreground">
              {displayName}
            </span>
            <div className="flex items-center gap-1">
              <Icon className="h-3 w-3 text-amber-500" />
              <span className="text-xs font-semibold text-muted-foreground">
                {formatLeaderboardValue(entry.value, type)}
              </span>
            </div>
            <div
              className={`hidden w-40 ${height} rounded-t-lg ${color} items-center justify-center text-2xl font-extrabold text-white sm:flex`}
            >
              {entry.rank}
            </div>
          </div>
        );
      })}
    </div>
  );
}
