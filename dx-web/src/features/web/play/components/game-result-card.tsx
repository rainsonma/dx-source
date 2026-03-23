"use client";

import { useRouter } from "next/navigation";
import {
  Trophy,
  Home,
  List,
  RotateCw,
  ChevronRight,
  BookOpen,
  Loader2,
} from "lucide-react";
import { ProgressRing } from "@/features/web/play/components/progress-ring";
import { useGameStore } from "@/features/web/play/hooks/use-game-store";
import { useGameResult } from "@/features/web/play/hooks/use-game-result";
import { getScoreRating } from "@/consts/score-rating";
import { advanceSessionLevelAction } from "@/features/web/play/actions/session.action";
import { SCORING } from "@/consts/scoring";

interface StatItem {
  value: string;
  label: string;
}

function StatBlock({ value, label }: StatItem) {
  return (
    <div className="flex h-16 flex-1 flex-col items-center justify-center rounded-[10px] border border-border bg-muted">
      <span className="text-base font-bold text-foreground">{value}</span>
      <span className="text-xs text-muted-foreground">{label}</span>
    </div>
  );
}

interface GameResultCardProps {
  game: {
    id: string;
    levels: { id: string; name: string; order: number }[];
  };
  elapsedTime: string;
}

export function GameResultCard({ game, elapsedTime }: GameResultCardProps) {
  const correctCount = useGameStore((s) => s.correctCount);
  const wrongCount = useGameStore((s) => s.wrongCount);
  const skipCount = useGameStore((s) => s.skipCount);
  const score = useGameStore((s) => s.score);
  const combo = useGameStore((s) => s.combo);
  const contentItems = useGameStore((s) => s.contentItems);
  const sessionId = useGameStore((s) => s.sessionId);
  const levelId = useGameStore((s) => s.levelId);
  const degree = useGameStore((s) => s.degree);
  const pattern = useGameStore((s) => s.pattern);
  const exitGame = useGameStore((s) => s.exitGame);

  const router = useRouter();
  const { completing, isLastLevel } = useGameResult({ levels: game.levels });

  const totalItems = contentItems?.length ?? 0;
  const accuracy = totalItems > 0 ? correctCount / totalItems : 0;
  const accuracyPercent = Math.round(accuracy * 100);
  const completionPercent = totalItems > 0
    ? Math.round(((totalItems - skipCount) / totalItems) * 100)
    : 0;
  const scorePercent = totalItems > 0
    ? Math.round(Math.min((score / (totalItems * SCORING.CORRECT_ANSWER)) * 100, 100))
    : 0;
  const comboPercent = totalItems > 0
    ? Math.round(Math.min((combo.maxCombo / totalItems) * 100, 100))
    : 0;
  const meetsThreshold = accuracy >= SCORING.EXP_ACCURACY_THRESHOLD;

  const rating = getScoreRating(accuracy);

  // Find next level for navigation
  const sortedLevels = [...game.levels].sort((a, b) => a.order - b.order);
  const currentLevelIndex = sortedLevels.findIndex((l) => l.id === levelId);
  const nextLevel = currentLevelIndex >= 0 ? sortedLevels[currentLevelIndex + 1] : undefined;

  const expEarned = meetsThreshold ? SCORING.LEVEL_COMPLETE_EXP : 0;

  const statsRow1: StatItem[] = [
    { value: `${correctCount}/${totalItems}`, label: "正确题数" },
    { value: `${wrongCount}`, label: "错误" },
    { value: `${skipCount}`, label: "跳过" },
  ];

  const statsRow2: StatItem[] = [
    { value: `${combo.maxCombo}连击`, label: "最高连击" },
    { value: elapsedTime, label: "用时" },
    { value: `+${expEarned}`, label: "经验值" },
  ];

  /** Navigate to home page */
  function handleHome() {
    router.push("/");
  }

  /** Navigate to games list */
  function handleGamesList() {
    router.push("/hall/games");
  }

  /** Navigate to levels list for current game */
  function handleLevelsList() {
    router.push(`/hall/games/${game.id}`);
  }

  /** Replay the same level — navigate to trigger full loading flow */
  function handleReplay() {
    const params = new URLSearchParams();
    if (degree) params.set("degree", degree);
    if (levelId) params.set("level", levelId);
    if (pattern) params.set("pattern", pattern);
    exitGame();
    router.push(`/hall/play/${game.id}?${params.toString()}`);
  }

  /** Advance to the next level via URL navigation (do NOT exitGame here — the store is reset by GamePlayShell's useEffect when new props arrive via soft navigation; calling exitGame before router.push would trigger a spurious loading screen with the OLD level ID) */
  async function handleNextLevel() {
    if (!sessionId || !nextLevel) return;
    const { error } = await advanceSessionLevelAction(sessionId, nextLevel.id);
    if (error) return;

    const params = new URLSearchParams();
    if (degree) params.set("degree", degree);
    params.set("level", nextLevel.id);
    if (pattern) params.set("pattern", pattern);
    router.push(`/hall/play/${game.id}?${params.toString()}`);
  }

  return (
    <div className="flex h-screen w-full items-center justify-center bg-slate-900/50 px-4">
      <div className="flex w-full max-w-2xl flex-col items-center gap-6 rounded-[20px] bg-card px-6 py-9 md:px-8">
        {/* Header */}
        <div className="flex w-full flex-col items-center gap-2">
          <Trophy className="h-12 w-12 text-amber-500" />
          <h2 className="text-[22px] font-bold text-foreground">恭喜完成!</h2>
        </div>

        {/* Score */}
        <div className="flex w-full flex-col items-center gap-1">
          <span className="-tracking-wider text-4xl font-extrabold text-teal-600 md:text-5xl">
            {score}分
          </span>
          <span className={`rounded-full px-3 py-0.5 text-sm font-semibold ${rating.colorClass} ${rating.bgClass}`}>{rating.label}</span>
        </div>

        {!meetsThreshold && (
          <p className="text-sm text-amber-500">
            正确率需达到60%才能获得经验值
          </p>
        )}

        <div className="h-px w-full bg-border" />

        {/* Ring row — key percentage KPIs */}
        <div className="grid w-full grid-cols-4 gap-2">
          <ProgressRing percent={accuracyPercent} color="stroke-amber-500" label="正确率" />
          <ProgressRing percent={completionPercent} color="stroke-amber-500" label="完成率" />
          <ProgressRing percent={scorePercent} color="stroke-teal-500" label="得分率" />
          <ProgressRing percent={comboPercent} color="stroke-teal-500" label="连击率" />
        </div>

        <div className="h-px w-full bg-border" />

        {/* Stats blocks — raw data */}
        <div className="flex w-full flex-col gap-3">
          <div className="grid grid-cols-3 gap-3">
            {statsRow1.map((stat) => (
              <StatBlock key={stat.label} {...stat} />
            ))}
          </div>
          <div className="grid grid-cols-3 gap-3">
            {statsRow2.map((stat) => (
              <StatBlock key={stat.label} {...stat} />
            ))}
          </div>
        </div>

        <div className="h-px w-full bg-border" />

        {completing && (
          <div className="flex items-center gap-2 text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span className="text-sm">正在保存...</span>
          </div>
        )}

        {/* Action buttons */}
        <div className="flex w-full flex-wrap gap-3">
          <button
            type="button"
            onClick={handleHome}
            disabled={completing}
            className="flex h-11 flex-1 basis-[calc(50%-6px)] items-center justify-center gap-2 rounded-xl border border-border bg-muted disabled:opacity-50 md:basis-0"
          >
            <Home className="h-[18px] w-[18px] text-muted-foreground" />
            <span className="text-sm font-semibold text-muted-foreground">返回主页</span>
          </button>
          <button
            type="button"
            onClick={handleGamesList}
            disabled={completing}
            className="flex h-11 flex-1 basis-[calc(50%-6px)] items-center justify-center gap-1.5 rounded-xl border border-border bg-muted disabled:opacity-50 md:basis-0"
          >
            <List className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold text-muted-foreground">课程列表</span>
          </button>
          <button
            type="button"
            onClick={handleLevelsList}
            disabled={completing}
            className="flex h-11 flex-1 basis-[calc(50%-6px)] items-center justify-center gap-1.5 rounded-xl border border-border bg-muted disabled:opacity-50 md:basis-0"
          >
            <BookOpen className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold text-muted-foreground">关卡列表</span>
          </button>
          <button
            type="button"
            onClick={handleReplay}
            disabled={completing}
            className="flex h-11 flex-1 basis-[calc(50%-6px)] items-center justify-center gap-2 rounded-xl bg-teal-600 disabled:opacity-50 md:basis-0"
          >
            <RotateCw className="h-4 w-4 text-white" />
            <span className="text-sm font-semibold text-white">再来一局</span>
          </button>
          {!isLastLevel && (
            <button
              type="button"
              onClick={handleNextLevel}
              disabled={completing}
              className="flex h-11 flex-1 basis-full items-center justify-center gap-1.5 rounded-xl border border-border bg-muted disabled:opacity-50 md:basis-0"
            >
              <span className="text-sm font-semibold text-muted-foreground">下一关</span>
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
