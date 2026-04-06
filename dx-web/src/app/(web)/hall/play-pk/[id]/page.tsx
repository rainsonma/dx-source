"use client";

import { useEffect, useState } from "react";
import { use } from "react";
import { notFound } from "next/navigation";
import { useSearchParams } from "next/navigation";
import { apiClient } from "@/lib/api-client";
import { PkPlayShell } from "@/features/web/play-pk/components/pk-play-shell";

export default function PkPlayPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const searchParams = useSearchParams();

  const degree = searchParams.get("degree");
  const pattern = searchParams.get("pattern");
  const level = searchParams.get("level");
  const difficulty = searchParams.get("difficulty") ?? "normal";
  const pkId = searchParams.get("pkId");
  const sessionId = searchParams.get("sessionId");

  type GameData = {
    id: string;
    name: string;
    mode: string;
    levels: { id: string; name: string; order: number }[];
  };

  type ApiGameData = {
    id: string;
    name: string;
    mode: string;
    levels?: { id: string; name: string; order: number }[];
  };

  type ApiProfileData = {
    id?: string;
    nickname?: string | null;
    username?: string;
    avatarUrl?: string | null;
  };

  const [game, setGame] = useState<GameData | null>(null);
  const [player, setPlayer] = useState<{
    id: string;
    nickname: string;
    avatarUrl: string | null;
  }>({ id: "", nickname: "我", avatarUrl: null });
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    async function load() {
      const [gameRes, profileRes] = await Promise.all([
        apiClient.get<ApiGameData>(`/api/games/${id}`),
        apiClient.get<ApiProfileData>("/api/user/profile"),
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
        levels: (g.levels ?? []).map((l) => ({
          id: l.id,
          name: l.name,
          order: l.order,
        })),
      });

      if (profileRes.code === 0 && profileRes.data) {
        setPlayer({
          id: profileRes.data.id ?? "",
          nickname:
            profileRes.data.nickname || profileRes.data.username || "我",
          avatarUrl: profileRes.data.avatarUrl ?? null,
        });
      }

      setLoaded(true);
    }

    load();
  }, [id]);

  if (!loaded) return null;
  if (!game || !degree) {
    notFound();
    return null;
  }

  const targetLevelId = level ?? game.levels[0]?.id ?? "";

  return (
    <PkPlayShell
      game={game}
      player={player}
      degree={degree}
      pattern={pattern}
      levelId={targetLevelId}
      difficulty={difficulty}
      existingPkId={pkId}
      existingSessionId={sessionId}
    />
  );
}
