"use client";

import { useEffect, useState } from "react";
import { use } from "react";
import { notFound } from "next/navigation";
import { apiClient, sessionApi } from "@/lib/api-client";
import { isVipActive } from "@/lib/vip";
import type { UserGrade } from "@/consts/user-grade";
import { BreadcrumbTopBar } from "@/features/web/hall/components/breadcrumb-top-bar";
import { RulesCard } from "@/features/web/games/components/rules-card";
import { MyStatsCard } from "@/features/web/games/components/my-stats-card";
import { GameDetailContent } from "@/features/web/games/components/game-detail-content";

const GAME_RULES: Record<string, string[]> = {
  "word-sentence": [
    "根据提示拼写出单词、短语或句子",
    "难度逐级递增：单词→短语→句子",
    "连续正确拼写获得额外分数奖励",
    "拼写错误不能得分",
    "完成所有题目结束游戏",
  ],
  "vocab-battle": [
    "根据中文释义选择正确的英文单词",
    "答对题目加分，答错题目不得分",
    "连续答对触发连击加分",
    "难度随关卡递增",
    "完成所有题目后结算得分",
  ],
  "vocab-match": [
    "将英文单词与对应的中文释义配对",
    "点击两个匹配项即可消除",
    "正确匹配得分",
    "错误配对不得分",
    "连续答对触发连击加分",
  ],
  "vocab-elimination": [
    "从多个选项中找出正确的单词释义",
    "逐步淘汰错误选项",
    "每轮淘汰后难度提升",
    "连续正确获得额外奖励",
    "全部淘汰完成即通关",
  ],
};

const DEFAULT_RULES = [
  "按照提示完成每道题目",
  "答对得分，答错扣分",
  "连续正确触发连击奖励",
  "难度随关卡递增",
  "完成所有题目后结算得分",
];

type GameLevel = { id: string; name: string; order: number };

type GameDetail = {
  id: string;
  name: string;
  description: string | null;
  mode: string;
  coverUrl: string | null;
  levelCount: number;
  levels: GameLevel[];
};

type ApiGameData = {
  id: string;
  name: string;
  description?: string | null;
  mode: string;
  coverUrl?: string | null;
  levelCount?: number;
  levels?: { id: string; name: string; order: number }[];
};

type GameStats = {
  highestScore: number;
  totalSessions: number;
  totalScores: number;
  totalExp: number;
};

type ActiveSession = {
  gameLevelId: string;
  degree: string;
  pattern: string | null;
};

export default function GameDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const [game, setGame] = useState<GameDetail | null>(null);
  const [heroSession, setHeroSession] = useState<{
    degree: string;
    pattern: string | null;
    levelName: string;
  } | null>(null);
  const [myStats, setMyStats] = useState<GameStats | null>(null);
  const [isFavorited, setIsFavorited] = useState(false);
  const [vip, setVip] = useState(false);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    async function load() {
      const res = await apiClient.get<ApiGameData>(`/api/games/${id}`);

      if (res.code !== 0 || !res.data) {
        setLoaded(true);
        return;
      }

      const g = res.data;
      const mapped: GameDetail = {
        id: g.id,
        name: g.name,
        description: g.description ?? null,
        mode: g.mode,
        coverUrl: g.coverUrl ?? null,
        levelCount: g.levelCount ?? 0,
        levels: (g.levels ?? []).map((l) => ({
          id: l.id,
          name: l.name,
          order: l.order,
        })),
      };

      const [activeSessionRes, statsRes, favoritedRes, profileRes] = await Promise.all([
        sessionApi.checkAnyActive(mapped.id),
        apiClient.get<GameStats>(`/api/games/${mapped.id}/stats`),
        apiClient.get<{ favorited: boolean }>(`/api/games/${mapped.id}/favorited`),
        apiClient.get<{ grade: string; vip_due_at: string | null }>("/api/user/profile"),
      ]);

      const activeSession = activeSessionRes.code === 0
        ? (activeSessionRes.data as ActiveSession | null)
        : null;
      const stats = statsRes.code === 0 ? statsRes.data : null;
      const favorited = favoritedRes.code === 0 ? favoritedRes.data.favorited : false;
      const profile = profileRes.code === 0 ? profileRes.data : null;
      const isVip = profile
        ? isVipActive(profile.grade as UserGrade, profile.vip_due_at)
        : false;

      if (activeSession) {
        const level = mapped.levels.find((l) => l.id === activeSession.gameLevelId);
        if (level) {
          setHeroSession({
            degree: activeSession.degree,
            pattern: activeSession.pattern,
            levelName: level.name,
          });
        }
      }

      setGame(mapped);
      setMyStats(stats);
      setIsFavorited(favorited);
      setVip(isVip);
      setLoaded(true);
    }

    load();
  }, [id]);

  if (!loaded) return null;
  if (!game) {
    notFound();
    return null;
  }

  const rules = GAME_RULES[game.mode] ?? DEFAULT_RULES;

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <BreadcrumbTopBar
        backHref="/hall/games"
        items={[
          { label: "课程游戏", href: "/hall/games", maxChars: 10 },
          { label: game.name, maxChars: 5 },
        ]}
      />
      <GameDetailContent
        game={{
          id: game.id,
          name: game.name,
          description: game.description ?? "",
          mode: game.mode,
          coverUrl: game.coverUrl,
          levelCount: game.levelCount,
          playerCount: String(myStats?.totalSessions ?? 0),
          levels: game.levels,
          completedLevels: 0,
          isVip: vip,
        }}
        heroSession={heroSession}
        isFavorited={isFavorited}
        rules={<RulesCard rules={rules} />}
        stats={
          <MyStatsCard
            stats={[
              { label: "最高得分", value: String(myStats?.highestScore ?? 0) },
              { label: "已玩次数", value: String(myStats?.totalSessions ?? 0) },
              { label: "累计得分", value: String(myStats?.totalScores ?? 0) },
              { label: "总经验值", value: String(myStats?.totalExp ?? 0) },
            ]}
          />
        }
      />
    </div>
  );
}
