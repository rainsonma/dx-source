"use client";

import { useEffect, useState } from "react";
import { Loader2 } from "lucide-react";
import { TabPill } from "@/components/in/tab-pill";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { useLeaderboard } from "../hooks/use-leaderboard";
import { LeaderboardMyRank } from "./leaderboard-my-rank";
import { LeaderboardPodium } from "./leaderboard-podium";
import { LeaderboardList } from "./leaderboard-list";
import type {
  LeaderboardType,
  LeaderboardPeriod,
} from "../types/leaderboard.types";

const TYPE_TABS: { label: string; value: LeaderboardType }[] = [
  { label: "时长", value: "playtime" },
  { label: "经验", value: "exp" },
];

const PERIOD_TABS: { label: string; value: LeaderboardPeriod }[] = [
  { label: "总榜", value: "all" },
  { label: "日榜", value: "day" },
  { label: "周榜", value: "week" },
  { label: "月榜", value: "month" },
];

/** Leaderboard content with type/period tab switching */
export function LeaderboardContent() {
  const { type, period, data, isLoading, handleTypeChange, handlePeriodChange } =
    useLeaderboard();

  const [user, setUser] = useState<{
    id: string; username: string; nickname: string | null; avatarUrl: string | null;
  } | null>(null);

  useEffect(() => {
    fetchUserProfile().then((profile) => {
      if (profile) setUser(profile);
    });
  }, []);

  const podiumEntries = data.entries.slice(0, 3);
  const listEntries = data.entries.slice(3);

  return (
    <>
      {/* Tab rows */}
      <div className="flex w-full flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2">
          {TYPE_TABS.map((tab) => (
            <TabPill
              key={tab.value}
              label={tab.label}
              active={type === tab.value}
              onClick={() => handleTypeChange(tab.value)}
            />
          ))}
        </div>
        <div className="flex items-center gap-2">
          {PERIOD_TABS.map((tab) => (
            <TabPill
              key={tab.value}
              label={tab.label}
              active={period === tab.value}
              onClick={() => handlePeriodChange(tab.value)}
              size="sm"
            />
          ))}
        </div>
      </div>

      {/* My rank */}
      {user && <LeaderboardMyRank entry={data.myRank} type={type} user={user} />}

      {/* Leaderboard content */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
        </div>
      ) : data.entries.length === 0 ? (
        <div className="flex items-center justify-center rounded-xl border border-border bg-card py-20 text-sm text-muted-foreground">
          暂无排名数据
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-border bg-card">
          {podiumEntries.length > 0 && (
            <>
              <LeaderboardPodium entries={podiumEntries} type={type} />
              <div className="h-px w-full bg-border" />
            </>
          )}
          <LeaderboardList entries={listEntries} type={type} />
        </div>
      )}
    </>
  );
}
