"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar";
import { AdCardsRow } from "@/features/web/hall/components/ad-cards-row";
import { StatsRow } from "@/features/web/hall/components/stats-row";
import { GameProgressCard } from "@/features/web/hall/components/game-progress-card";
import { DailyChallengeCard } from "@/features/web/hall/components/daily-challenge-card";
import { LearningHeatmap } from "@/features/web/hall/components/learning-heatmap";

type DashboardData = {
  profile: {
    username: string;
    nickname: string | null;
    exp: number;
    currentPlayStreak: number;
  };
  masterStats: { total: number; thisWeek: number };
  reviewStats: { pending: number };
  sessions: any[];
  todayAnswers: number;
};

type HeatmapData = {
  year: number;
  days: { date: string; count: number }[];
  accountYear: number;
};

export default function HallDashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [heatmap, setHeatmap] = useState<HeatmapData | null>(null);

  useEffect(() => {
    async function load() {
      const [dashRes, heatmapRes] = await Promise.all([
        apiClient.get<DashboardData>("/api/hall/dashboard"),
        apiClient.get<HeatmapData>(`/api/hall/heatmap?year=${new Date().getFullYear()}`),
      ]);

      if (dashRes.code === 0) setData(dashRes.data);
      if (heatmapRes.code === 0) setHeatmap(heatmapRes.data);
    }

    load();
  }, []);

  const displayName = data?.profile.nickname ?? data?.profile.username ?? "同学";

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <GreetingTopBar
        title={`早上好，${displayName} 👋`}
        subtitle="继续你的学习之旅，今天也要加油！"
      />
      <AdCardsRow />
      <StatsRow
        exp={data?.profile.exp ?? 0}
        currentPlayStreak={data?.profile.currentPlayStreak ?? 0}
        masteredTotal={data?.masterStats.total ?? 0}
        masteredThisWeek={data?.masterStats.thisWeek ?? 0}
        reviewPending={data?.reviewStats.pending ?? 0}
      />

      {/* Main content row */}
      <div className="flex flex-1 flex-col gap-5 lg:flex-row">
        {/* Left column - game progress */}
        <div className="flex flex-1 flex-col gap-5">
          <GameProgressCard sessions={data?.sessions ?? []} />
        </div>

        {/* Right column - daily challenge */}
        <div className="flex w-full flex-col gap-5 lg:w-80 lg:shrink-0">
          <DailyChallengeCard />
        </div>
      </div>

      {/* Learning heatmap */}
      {heatmap && (
        <LearningHeatmap
          initialYear={heatmap.year}
          initialDays={heatmap.days}
          accountYear={heatmap.accountYear}
        />
      )}
    </div>
  );
}
