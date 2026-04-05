"use client";

import { useEffect, useState } from "react";
import {
  Keyboard,
  Swords,
  Headphones,
  Crosshair,
  Lightbulb,
  RotateCcw,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { GAME_MODES, type GameMode } from "@/consts/game-mode";
import { GAME_DEGREE_LABELS, type GameDegree } from "@/consts/game-degree";
import {
  startSessionAction,
  fetchSessionRestoreDataAction,
} from "@/features/web/play-single/actions/session.action";
import { fetchLevelContentAction } from "@/features/web/play-core/actions/content.action";
import { useGameStore, type ContentItem } from "@/features/web/play-core/hooks/use-game-store";

const MINIMUM_DISPLAY_MS = 1200;

interface LoadingConfig {
  title: string;
  tip: string;
  icon: LucideIcon;
  iconColor: string;
  iconBg: string;
  iconBorder: string;
}

const loadingData: Record<GameMode, LoadingConfig> = {
  [GAME_MODES.WORD_SENTENCE]: {
    title: "连词成句",
    tip: "提示：连续正确拼写可获得额外奖励",
    icon: Keyboard,
    iconColor: "text-purple-400",
    iconBg: "bg-purple-500/[0.13]",
    iconBorder: "border-purple-500/25",
  },
  [GAME_MODES.VOCAB_MATCH]: {
    title: "词汇配对",
    tip: "提示：连续正确匹配可获得连击奖励",
    icon: Swords,
    iconColor: "text-blue-300",
    iconBg: "bg-blue-500/[0.13]",
    iconBorder: "border-blue-500/25",
  },
  [GAME_MODES.LISTENING_CHALLENGE]: {
    title: "听力闯关",
    tip: "提示：可重复播放语音，但会扣减分数哦",
    icon: Headphones,
    iconColor: "text-yellow-300",
    iconBg: "bg-amber-500/[0.13]",
    iconBorder: "border-amber-500/25",
  },
  [GAME_MODES.VOCAB_ELIMINATION]: {
    title: "消消乐",
    tip: "提示：连续消除不同配对可获得连击奖励",
    icon: Keyboard,
    iconColor: "text-pink-300",
    iconBg: "bg-pink-500/[0.13]",
    iconBorder: "border-pink-500/25",
  },
  [GAME_MODES.VOCAB_BATTLE]: {
    title: "词汇对轰",
    tip: "提示：连续正确拼写可形成连击，炸弹威力更大",
    icon: Crosshair,
    iconColor: "text-red-300",
    iconBg: "bg-red-500/[0.13]",
    iconBorder: "border-red-500/25",
  },
};

interface GameLoadingScreenProps {
  gameId: string;
  gameName: string;
  gameMode: string;
  degree?: string;
  pattern?: string;
  levelId?: string;
  levelName?: string;
}

export function GameLoadingScreen({
  gameId,
  gameName,
  gameMode,
  degree,
  pattern,
  levelId,
  levelName,
}: GameLoadingScreenProps) {
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  const initSession = useGameStore((s) => s.initSession);

  const config = loadingData[gameMode as GameMode];
  const degreeLabel = degree
    ? GAME_DEGREE_LABELS[degree as GameDegree] ?? degree
    : "";
  const subtitle = [gameName, degreeLabel, levelName].filter(Boolean).join(" · ");

  useEffect(() => {
    let cancelled = false;

    loadGameData();

    async function loadGameData() {
      const timerPromise = new Promise<void>((resolve) =>
        setTimeout(resolve, MINIMUM_DISPLAY_MS)
      );

      try {
        setProgress(0);

        const resolvedLevelId = levelId ?? gameId;

        // Step 1: Start/resume session (now includes level)
        const sessionResult = await startSessionAction(gameId, resolvedLevelId, degree, pattern);
        if (cancelled) return;
        if (sessionResult.error || !sessionResult.data) {
          setError(sessionResult.error ?? "无法开始游戏");
          return;
        }
        setProgress(33);

        // Step 2: Restore session data if resuming
        const resumeItemId = sessionResult.data.currentContentItemId ?? null;

        let restored: {
          score: number;
          maxCombo: number;
          correctCount: number;
          wrongCount: number;
          skipCount: number;
          playTime: number;
        } | null = null;

        if (resumeItemId) {
          const restoreResult = await fetchSessionRestoreDataAction(
            sessionResult.data.id
          );
          if (!cancelled && restoreResult.data) {
            restored = {
              score: restoreResult.data.score,
              maxCombo: restoreResult.data.maxCombo,
              correctCount: restoreResult.data.correctCount,
              wrongCount: restoreResult.data.wrongCount,
              skipCount: restoreResult.data.skipCount,
              playTime: restoreResult.data.playTime,
            };
          }
        }
        setProgress(66);

        // Step 3: Fetch content for the resolved level
        const contentResult = await fetchLevelContentAction(gameId, resolvedLevelId, degree);
        if (cancelled) return;
        if (contentResult.error || !contentResult.data) {
          setError(contentResult.error ?? "加载内容失败");
          return;
        }
        setProgress(100);

        let startFromIndex = 0;
        if (resumeItemId) {
          const idx = contentResult.data.findIndex(
            (item) => item.id === resumeItemId
          );
          if (idx > 0) startFromIndex = idx;
        }

        await timerPromise;
        if (cancelled) return;

        initSession({
          sessionId: sessionResult.data.id,
          gameId,
          gameMode,
          degree: degree ?? "intermediate",
          pattern: pattern ?? null,
          levelId: resolvedLevelId,
          contentItems: contentResult.data as ContentItem[],
          startFromIndex,
          ...(restored && { restored }),
        });
      } catch {
        if (!cancelled) setError("加载失败，请重试");
      }
    }

    return () => {
      cancelled = true;
    };
  }, [gameId, gameMode, degree, pattern, levelId, initSession, retryCount]);

  function handleRetry() {
    setError(null);
    setProgress(0);
    setRetryCount((c) => c + 1);
  }

  if (!config) return null;

  const IconComponent = config.icon;

  return (
    <div className="flex h-screen w-full flex-col items-center justify-center gap-9 px-4 bg-[radial-gradient(ellipse_at_center,#1E1B4B,#0F0A2E)]">
      <div className="h-[3px] w-full max-w-xl rounded-full bg-gradient-to-r from-transparent via-purple-600 to-transparent" />

      <div
        className={`flex h-20 w-20 items-center justify-center rounded-[20px] border ${config.iconBorder} ${config.iconBg}`}
      >
        <IconComponent className={`h-10 w-10 ${config.iconColor}`} />
      </div>

      <h1 className="-tracking-wider text-3xl font-extrabold text-white md:text-4xl">
        {config.title}
      </h1>

      <p className="text-base font-medium text-slate-400">{subtitle}</p>

      {error ? (
        <div className="flex flex-col items-center gap-4">
          <p className="text-sm font-medium text-red-400">{error}</p>
          <button
            type="button"
            onClick={handleRetry}
            className="flex items-center gap-2 rounded-xl bg-white/10 px-5 py-2.5"
          >
            <RotateCcw className="h-4 w-4 text-white" />
            <span className="text-sm font-semibold text-white">重试</span>
          </button>
        </div>
      ) : (
        <>
          <div className="h-1.5 w-full max-w-xl overflow-hidden rounded-full bg-white/[0.08]">
            <div
              className="h-full rounded-full bg-gradient-to-r from-purple-600 to-teal-600 transition-all duration-300 ease-out"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-sm font-medium text-slate-500">
            {progress < 100 ? "正在加载..." : "准备就绪"}
          </p>
        </>
      )}

      <div className="flex items-center gap-2">
        <Lightbulb className="h-3.5 w-3.5 shrink-0 text-amber-500" />
        <span className="text-[13px] text-slate-600">{config.tip}</span>
      </div>
    </div>
  );
}
