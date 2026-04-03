"use client";

import { useEffect, useState } from "react";
import { use } from "react";
import { notFound, useRouter } from "next/navigation";
import { useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { isVipActive } from "@/lib/vip";
import type { UserGrade } from "@/consts/user-grade";
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
    grade?: string;
    vip_due_at?: string | null;
  };

  const router = useRouter();
  const [game, setGame] = useState<GameData | null>(null);
  const [player, setPlayer] = useState<{ nickname: string; avatarUrl: string | null }>({
    nickname: "我",
    avatarUrl: null,
  });
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

      // VIP guard: redirect if non-VIP tries to access non-first level
      const profileData = profileRes.code === 0 ? profileRes.data : null;
      const userIsVip = profileData
        ? isVipActive((profileData.grade ?? "free") as UserGrade, profileData.vip_due_at ?? null)
        : false;

      if (!userIsVip && level) {
        const firstLevel = g.levels?.[0];
        if (firstLevel && level !== firstLevel.id) {
          toast.error("升级会员解锁全部关卡");
          router.replace(`/hall/games/${id}`);
          return;
        }
      }

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
          nickname: profileRes.data.nickname || profileRes.data.username || "我",
          avatarUrl: profileRes.data.avatarUrl ?? null,
        });
      }

      setLoaded(true);
    }

    load();
  }, [id, level, router]);

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
