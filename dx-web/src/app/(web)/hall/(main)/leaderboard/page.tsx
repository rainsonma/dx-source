"use client";

import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
import { LeaderboardContent } from "@/features/web/leaderboard/components/leaderboard-content";

export default function LeaderboardPage() {
  const menu = useHallMenuItem("/hall/leaderboard");

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索用户..."
      />
      <LeaderboardContent />
    </div>
  );
}
