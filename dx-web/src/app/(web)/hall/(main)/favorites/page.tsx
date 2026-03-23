"use client";

import { useEffect, useState } from "react";
import { Heart } from "lucide-react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { FavoriteCard } from "@/features/web/hall/components/favorite-card";

export default function FavoritesPage() {
  const [favorites, setFavorites] = useState<any[]>([]);

  useEffect(() => {
    async function load() {
      const res = await apiClient.get<any[]>("/api/favorites");
      const rawFavorites = res.code === 0 ? res.data ?? [] : [];
      setFavorites(
        rawFavorites.map((g: any) => ({
          id: g.id,
          name: g.name,
          description: g.description ?? null,
          mode: g.mode,
          cover: g.coverUrl ? { url: g.coverUrl } : null,
          category: g.categoryName ? { name: g.categoryName } : null,
          user: g.author ? { username: g.author } : null,
        }))
      );
    }

    load();
  }, []);

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="我的收藏"
        subtitle="收藏你喜欢的课程游戏和学习内容"
        searchPlaceholder="搜索收藏..."
      />

      {/* Filter row */}
      <div className="flex items-center justify-between">
        <span className="rounded-full bg-teal-600 px-5 py-2 text-[13px] font-semibold text-white">
          全部 ({favorites.length})
        </span>
        <span className="text-[13px] text-muted-foreground">
          共 {favorites.length} 个收藏
        </span>
      </div>

      <div className="h-px w-full bg-border" />

      {favorites.length > 0 ? (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
          {favorites.map((game) => (
            <FavoriteCard key={game.id} game={game} />
          ))}
        </div>
      ) : (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 text-muted-foreground">
          <Heart className="h-12 w-12 stroke-1" />
          <p className="text-sm">还没有收藏，去发现喜欢的课程游戏吧</p>
        </div>
      )}
    </div>
  );
}
