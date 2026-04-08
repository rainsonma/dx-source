"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, Loader2, Trophy } from "lucide-react";
import { TabPill } from "@/components/in/tab-pill";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { useTodayStars } from "@/features/web/hall/hooks/use-today-stars";
import type { LeaderboardType } from "@/features/web/leaderboard/types/leaderboard.types";
import { TodayStarsPodium } from "./today-stars-podium";
import { TodayStarsList } from "./today-stars-list";
import { TodayStarsMyRank } from "./today-stars-my-rank";

const TYPE_TABS: { label: string; value: LeaderboardType }[] = [
  { label: "经验", value: "exp" },
  { label: "时长", value: "playTime" },
];

/** 今日明星榜 — Today's star leaderboard for the hall dashboard */
export function TodayStarsCard() {
  const { type, data, isLoading, handleTypeChange } = useTodayStars();

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
    <div className="flex w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-amber-50">
            <Trophy className="h-5 w-5 text-amber-500" />
          </div>
          <h3 className="text-base font-bold text-foreground">今日明星榜</h3>
        </div>
        <Link
          href="/hall/leaderboard"
          className="flex items-center gap-1 text-[13px] font-semibold text-teal-600 hover:text-teal-700"
        >
          查看全部
          <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>

      {/* Type tabs */}
      <div className="flex items-center gap-2">
        {TYPE_TABS.map((tab) => (
          <TabPill
            key={tab.value}
            label={tab.label}
            active={type === tab.value}
            onClick={() => handleTypeChange(tab.value)}
            size="sm"
          />
        ))}
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-5 w-5 animate-spin text-teal-600" />
        </div>
      ) : data.entries.length === 0 ? (
        <div className="flex items-center justify-center rounded-lg border border-border py-12 text-sm text-muted-foreground">
          暂无排名数据
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-border">
          {podiumEntries.length > 0 && (
            <>
              <TodayStarsPodium entries={podiumEntries} type={type} />
              <div className="h-px w-full bg-border" />
            </>
          )}
          <TodayStarsList entries={listEntries} type={type} />
        </div>
      )}

      {/* My rank */}
      {user && <TodayStarsMyRank entry={data.myRank} type={type} user={user} />}
    </div>
  );
}
