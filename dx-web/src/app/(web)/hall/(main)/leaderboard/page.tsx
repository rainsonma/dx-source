"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { LeaderboardContent } from "@/features/web/leaderboard/components/leaderboard-content";
import type { LeaderboardResult } from "@/features/web/leaderboard/types/leaderboard.types";

export default function LeaderboardPage() {
  const [initialData, setInitialData] = useState<LeaderboardResult>({
    entries: [],
    myRank: null,
  });

  useEffect(() => {
    async function load() {
      const res = await apiClient.get<LeaderboardResult>(
        "/api/leaderboard?type=exp&period=all"
      );
      if (res.code === 0) setInitialData(res.data);
    }

    load();
  }, []);

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <PageTopBar
        title="排行榜"
        subtitle="查看学习排名，与好友一起进步"
        searchPlaceholder="搜索用户..."
      />
      <LeaderboardContent initialData={initialData} />
    </div>
  );
}
