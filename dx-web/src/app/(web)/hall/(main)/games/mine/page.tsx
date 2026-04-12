"use client";

import { useEffect, useState } from "react";
import { Gamepad2 } from "lucide-react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
import { PlayedGameCard } from "@/features/web/hall/components/played-game-card";

type PlayedGame = {
  id: string;
  name: string;
  description: string | null;
  mode: string;
  cover: { url: string } | null;
  category: { name: string } | null;
  user: { username: string } | null;
  highestScore: number;
  totalPlayTime: number;
};

type ApiPlayedItem = {
  id: string;
  name: string;
  description?: string | null;
  mode: string;
  coverUrl?: string | null;
  categoryName?: string | null;
  author?: string | null;
  highestScore?: number;
  totalPlayTime?: number;
};

export default function MyGamesPage() {
  const menu = useHallMenuItem("/hall/games/mine");
  const [games, setGames] = useState<PlayedGame[]>([]);

  useEffect(() => {
    async function load() {
      const res = await apiClient.get<ApiPlayedItem[]>("/api/games/played");
      const rawGames = res.code === 0 ? res.data ?? [] : [];
      setGames(
        rawGames.map((g) => ({
          id: g.id,
          name: g.name,
          description: g.description ?? null,
          mode: g.mode,
          cover: g.coverUrl ? { url: g.coverUrl } : null,
          category: g.categoryName ? { name: g.categoryName } : null,
          user: g.author ? { username: g.author } : null,
          highestScore: g.highestScore ?? 0,
          totalPlayTime: g.totalPlayTime ?? 0,
        }))
      );
    }

    load();
  }, []);

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索游戏..."
      />

      {/* Filter row */}
      <div className="flex items-center justify-between">
        <span className="rounded-full bg-teal-600 px-5 py-2 text-[13px] font-semibold text-white">
          全部 ({games.length})
        </span>
        <span className="text-[13px] text-muted-foreground">
          共 {games.length} 个游戏
        </span>
      </div>

      {games.length > 0 ? (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
          {games.map((game) => (
            <PlayedGameCard key={game.id} game={game} />
          ))}
        </div>
      ) : (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 text-muted-foreground">
          <Gamepad2 className="h-12 w-12 stroke-1" />
          <p className="text-sm">还没有玩过游戏，去发现课程游戏吧</p>
        </div>
      )}
    </div>
  );
}
