import Link from "next/link";
import { ChevronRight } from "lucide-react";

import { getLevel, getExpForLevel } from "@/consts/user-level";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Learning stats block (level, EXP progress, streaks) with leaderboard link */
export function StatsBlock({ profile }: { profile: MeProfile }) {
  const level = getLevel(profile.exp);
  const currentLevelExp = getExpForLevel(level);
  const nextLevelExp = level < 100 ? getExpForLevel(level + 1) : currentLevelExp;
  const progress = nextLevelExp > currentLevelExp
    ? ((profile.exp - currentLevelExp) / (nextLevelExp - currentLevelExp)) * 100
    : 100;

  const lastPlayed = profile.lastPlayedAt
    ? new Date(profile.lastPlayedAt).toLocaleDateString("zh-CN")
    : "—";

  return (
    <div className="rounded-2xl border border-border bg-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-foreground">学习数据</h3>
        <Link
          href="/hall/leaderboard"
          className="flex items-center gap-1 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          查看排行
          <ChevronRight className="h-4 w-4" />
        </Link>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <div className="flex flex-col gap-2 md:col-span-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-muted-foreground">等级 Lv.{level}</span>
            <span className="text-xs text-muted-foreground">
              {profile.exp.toLocaleString()} / {nextLevelExp.toLocaleString()} EXP
            </span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-teal-500 transition-all"
              style={{ width: `${Math.min(progress, 100)}%` }}
            />
          </div>
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">当前连续</span>
          <span className="text-sm font-medium text-foreground">{profile.currentPlayStreak} 天</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">最高连续</span>
          <span className="text-sm font-medium text-foreground">{profile.maxPlayStreak} 天</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">上次学习</span>
          <span className="text-sm text-foreground">{lastPlayed}</span>
        </div>
      </div>
    </div>
  );
}
