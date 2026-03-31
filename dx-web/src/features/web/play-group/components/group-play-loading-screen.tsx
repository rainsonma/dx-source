"use client";

import { useEffect, useState } from "react";
import {
  Keyboard,
  Swords,
  Headphones,
  Lightbulb,
  Crosshair,
  RotateCcw,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { GAME_MODES, type GameMode } from "@/consts/game-mode";
import { GAME_DEGREE_LABELS, type GameDegree } from "@/consts/game-degree";
import {
  startSessionAction,
  startLevelAction,
  restoreSessionDataAction,
  fetchLevelContentAction,
} from "../actions/session.action";
import { useGroupPlayStore } from "../hooks/use-group-play-store";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import type { ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import type { Participants } from "../types/group-play";

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
  [GAME_MODES.LSRW]: {
    title: "听说读写",
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

interface GroupPlayLoadingScreenProps {
  gameId: string;
  gameName: string;
  gameMode: string;
  degree: string;
  pattern: string | null;
  levelId: string;
  levelName?: string;
  gameGroupId: string;
  levelTimeLimit: number;
}

export function GroupPlayLoadingScreen({
  gameId,
  gameName,
  gameMode,
  degree,
  pattern,
  levelId,
  levelName,
  gameGroupId,
  levelTimeLimit,
}: GroupPlayLoadingScreenProps) {
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  const initGroupSession = useGroupPlayStore((s) => s.initSession);
  const initGameSession = useGameStore((s) => s.initSession);

  const config = loadingData[gameMode as GameMode];
  const degreeLabel = GAME_DEGREE_LABELS[degree as GameDegree] ?? degree;
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

        // Read participants roster from sessionStorage (stored by game room on game start)
        let participants: Participants | null = null;
        try {
          const raw = sessionStorage.getItem(`group-participants:${gameGroupId}`);
          if (raw) {
            participants = JSON.parse(raw) as Participants;
          }
        } catch {
          // Proceed without participants if sessionStorage read fails
        }

        // Step 1: Start group-play session
        const sessionResult = await startSessionAction(
          gameId,
          degree,
          pattern,
          levelId,
          gameGroupId
        );
        if (cancelled) return;
        if (sessionResult.error || !sessionResult.data) {
          setError(sessionResult.error ?? "无法开始游戏");
          return;
        }
        setProgress(25);

        const resolvedLevelId = sessionResult.data.levelId ?? levelId;

        // Step 2: Start level session
        const levelResult = await startLevelAction(
          sessionResult.data.id,
          resolvedLevelId,
          degree,
          pattern
        );
        if (cancelled) return;
        if (levelResult.error || !levelResult.data) {
          setError(levelResult.error ?? "开始关卡失败");
          return;
        }
        setProgress(50);

        // Step 3: Restore data if resuming
        const levelSessionResumeItemId =
          levelResult.data.currentContentItemId ?? null;

        let restored: {
          score: number;
          maxCombo: number;
          correctCount: number;
          wrongCount: number;
          skipCount: number;
          playTime: number;
        } | null = null;

        if (levelSessionResumeItemId) {
          const restoreResult = await restoreSessionDataAction(
            sessionResult.data.id
          );
          if (!cancelled && restoreResult.data?.sessionLevel) {
            const sl = restoreResult.data.sessionLevel;
            restored = {
              score: sl.score,
              maxCombo: sl.maxCombo,
              correctCount: sl.correctCount,
              wrongCount: sl.wrongCount,
              skipCount: sl.skipCount,
              playTime: sl.playTime,
            };
          }
        }
        setProgress(75);

        // Step 4: Fetch content
        const contentResult = await fetchLevelContentAction(
          gameId,
          resolvedLevelId,
          degree
        );
        if (cancelled) return;
        if (contentResult.error || !contentResult.data) {
          setError(contentResult.error ?? "加载内容失败");
          return;
        }
        setProgress(100);

        let startFromIndex = 0;
        if (levelSessionResumeItemId) {
          const idx = contentResult.data.findIndex(
            (item) => item.id === levelSessionResumeItemId
          );
          if (idx > 0) startFromIndex = idx;
        }

        await timerPromise;
        if (cancelled) return;

        const sessionInit = {
          sessionId: sessionResult.data.id,
          levelSessionId: levelResult.data.id,
          gameId,
          gameMode,
          degree,
          pattern,
          levelId: resolvedLevelId,
          contentItems: contentResult.data as ContentItem[],
          startFromIndex,
          gameGroupId,
          levelTimeLimit,
          ...(restored && { restored }),
          ...(participants && { participants }),
        };

        // Init group store (shell state management)
        initGroupSession(sessionInit);
        // Init game store (required by shared game components like GameLsrw)
        initGameSession(sessionInit);
      } catch {
        if (!cancelled) setError("加载失败，请重试");
      }
    }

    return () => {
      cancelled = true;
    };
  }, [
    gameId,
    gameMode,
    degree,
    pattern,
    levelId,
    gameGroupId,
    levelTimeLimit,
    initGroupSession,
    initGameSession,
    retryCount,
  ]);

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
