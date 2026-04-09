"use client";

import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { LeaderboardContent } from "@/features/web/leaderboard/components/leaderboard-content";

export default function LeaderboardPage() {
  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <PageTopBar
        title="排行榜"
        subtitle="查看学习排名，与好友一起进步"
        searchPlaceholder="搜索用户..."
      />
      <LeaderboardContent />
    </div>
  );
}
