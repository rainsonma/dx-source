"use client";

import { useEffect, useState } from "react";
import { use } from "react";
import { notFound } from "next/navigation";
import { useSearchParams } from "next/navigation";
import { apiClient } from "@/lib/api-client";
import { GamePlayShell } from "@/features/web/play-single/components/game-play-shell";

export default function GamePlayPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const searchParams = useSearchParams();
  const degree = searchParams.get("degree") ?? undefined;
  const level = searchParams.get("level") ?? undefined;
  const pattern = searchParams.get("pattern") ?? undefined;

  const [game, setGame] = useState<any>(null);
  const [player, setPlayer] = useState<{ nickname: string; avatarUrl: string | null }>({
    nickname: "我",
    avatarUrl: null,
  });
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    async function load() {
      const [gameRes, profileRes] = await Promise.all([
        apiClient.get<any>(`/api/games/${id}`),
        apiClient.get<any>("/api/user/profile"),
      ]);

      if (gameRes.code !== 0 || !gameRes.data) {
        setLoaded(true);
        return;
      }

      const g = gameRes.data;
      setGame({
        id: g.id,
        name: g.name,
        mode: g.mode,
        levels: ((g.levels as any[]) ?? []).map((l: any) => ({
          id: l.id,
          name: l.name,
          order: l.order,
        })),
      });

      if (profileRes.code === 0 && profileRes.data) {
        setPlayer({
          nickname: profileRes.data.nickname || profileRes.data.username || "我",
          avatarUrl: profileRes.data.avatarUrl ?? null,
        });
      }

      setLoaded(true);
    }

    load();
  }, [id]);

  if (!loaded) return null;
  if (!game) {
    notFound();
    return null;
  }

  return (
    <GamePlayShell
      game={game}
      player={player}
      degree={degree}
      pattern={pattern}
      levelId={level}
    />
  );
}
