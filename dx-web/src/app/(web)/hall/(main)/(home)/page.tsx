"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar";
import { AdCardsRow } from "@/features/web/hall/components/ad-cards-row";
import { StatsRow } from "@/features/web/hall/components/stats-row";
import { GameProgressCard } from "@/features/web/hall/components/game-progress-card";
import { DailyChallengeCard } from "@/features/web/hall/components/daily-challenge-card";
import { TodayStarsCard } from "@/features/web/hall/components/today-stars-card";
import { LearningHeatmap } from "@/features/web/hall/components/learning-heatmap";
import { NotificationBanner } from "@/features/web/hall/components/notification-banner";

type DashboardData = {
  profile: {
    username: string;
    nickname: string | null;
    exp: number;
    currentPlayStreak: number;
  };
  masterStats: { total: number; thisWeek: number };
  reviewStats: { pending: number };
  sessions: {
    gameId: string;
    gameName: string;
    gameMode: string;
    completedLevels: number;
    totalLevels: number;
    score: number;
    exp: number;
    lastPlayedAt: Date;
  }[];
  todayAnswers: number;
  greeting: {
    title: string;
    subtitle: string;
  };
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
        title={
          data?.greeting
            ? `${data.greeting.title}，${displayName}`
            : `早上好，${displayName}`
        }
        subtitle={
          data?.greeting?.subtitle ?? "继续你的学习之旅，今天也要加油！"
        }
      />
      <AdCardsRow />
      <NotificationBanner />

      {/* Main content row */}
      <div className="flex flex-col gap-5 lg:grid lg:grid-cols-3">
        {/* Right column - daily check-in (shows first on mobile) */}
        <div className="order-first lg:order-last lg:h-full">
          <DailyChallengeCard />
        </div>

        {/* Left column - game progress */}
        <div className="order-2 lg:order-first lg:h-full">
          <GameProgressCard sessions={data?.sessions ?? []} />
        </div>

        {/* Center column - today's stars */}
        <div className="order-3 lg:order-2 lg:h-full">
          <TodayStarsCard />
        </div>
      </div>

      <StatsRow
        exp={data?.profile.exp ?? 0}
        currentPlayStreak={data?.profile.currentPlayStreak ?? 0}
        masteredTotal={data?.masterStats.total ?? 0}
        masteredThisWeek={data?.masterStats.thisWeek ?? 0}
        reviewPending={data?.reviewStats.pending ?? 0}
      />

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
