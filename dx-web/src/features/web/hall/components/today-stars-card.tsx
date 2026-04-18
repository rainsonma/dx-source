"use client";

import Link from "next/link";
import { ArrowRight, Loader2, Trophy } from "lucide-react";
import { useTodayStars } from "@/features/web/hall/hooks/use-today-stars";
import { TodayStarsPodium } from "./today-stars-podium";
import { TodayStarsList } from "./today-stars-list";

/** 今日明星榜 — Today's star leaderboard for the hall dashboard */
export function TodayStarsCard() {
  const { data, isLoading } = useTodayStars();

  const podiumEntries = data.entries.slice(0, 3);
  const listEntries = data.entries.slice(3);

  return (
    <div className="flex h-full w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-5">
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

      {/* Body */}
      {isLoading ? (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 className="h-5 w-5 animate-spin text-teal-600" />
        </div>
      ) : data.entries.length === 0 ? (
        <div className="flex flex-1 items-center justify-center rounded-lg border border-border text-sm text-muted-foreground">
          暂无排名数据
        </div>
      ) : (
        <div className="flex min-h-0 flex-1 flex-col gap-4">
          {podiumEntries.length > 0 && (
            <TodayStarsPodium entries={podiumEntries} type="playtime" />
          )}
          {listEntries.length > 0 && (
            <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-border">
              <TodayStarsList entries={listEntries} type="playtime" />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
