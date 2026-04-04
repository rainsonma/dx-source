"use client";

import { useEffect, useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import {
  X,
  Zap,
  Flame,
  Trophy,
  ChevronRight,
  Play,
  RotateCcw,
  ArrowRight,
  Headphones,
  Mic,
  Eye,
  PenLine,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { GAME_DEGREES, type GameDegree } from "@/consts/game-degree";
import { GAME_MODES } from "@/consts/game-mode";
import {
  GAME_PATTERNS,
  DEFAULT_GAME_PATTERN,
  type GamePattern,
} from "@/consts/game-pattern";
import {
  checkActiveSessionAction,
  checkActiveLevelSessionAction,
  restartLevelSessionAction,
} from "@/features/web/play-single/actions/session.action";

interface ModeOption {
  value: GameDegree;
  title: string;
  desc: string;
  icon: React.ComponentType<{ className?: string }>;
  iconColor: string;
  iconBg: string;
}

const modeOptions: ModeOption[] = [
  {
    value: GAME_DEGREES.BEGINNER,
    title: "初级",
    desc: "基础拼写，逐步提高，不断进步",
    icon: Zap,
    iconColor: "text-emerald-500",
    iconBg: "bg-emerald-500/[0.08]",
  },
  {
    value: GAME_DEGREES.INTERMEDIATE,
    title: "中级",
    desc: "标准难度，中等输入，正常提升",
    icon: Flame,
    iconColor: "text-amber-500",
    iconBg: "bg-amber-500/[0.08]",
  },
  {
    value: GAME_DEGREES.ADVANCED,
    title: "高级",
    desc: "最高难度，长难输入，极限挑战",
    icon: Trophy,
    iconColor: "text-red-500",
    iconBg: "bg-red-500/[0.08]",
  },
];

const patternOptions: { value: GamePattern; label: string; icon: LucideIcon }[] = [
  { value: GAME_PATTERNS.LISTEN, label: "听", icon: Headphones },
  { value: GAME_PATTERNS.SPEAK, label: "说", icon: Mic },
  { value: GAME_PATTERNS.READ, label: "读", icon: Eye },
  { value: GAME_PATTERNS.WRITE, label: "写", icon: PenLine },
];

const difficultyOptions: { value: string; label: string; icon: LucideIcon }[] = [
  { value: "easy", label: "简单", icon: Zap },
  { value: "normal", label: "标准", icon: Flame },
  { value: "hard", label: "困难", icon: Trophy },
];

interface GameModeCardProps {
  gameId: string;
  gameName: string;
  gameMode: string;
  mode?: "single" | "pk";
  levelId?: string;
  levelLabel?: string;
  initialDegree?: string;
  initialPattern?: string | null;
  open: boolean;
  onClose: () => void;
}

export function GameModeCard({
  gameId,
  gameName,
  gameMode,
  mode = "single",
  levelId,
  levelLabel,
  initialDegree,
  initialPattern,
  open,
  onClose,
}: GameModeCardProps) {
  const router = useRouter();
  const [selectedDegree, setSelectedDegree] = useState<GameDegree>(
    (initialDegree as GameDegree) ?? GAME_DEGREES.BEGINNER
  );
  const isWordSentence = gameMode === GAME_MODES.WORD_SENTENCE;
  const [selectedPattern, setSelectedPattern] = useState<GamePattern>(
    (initialPattern as GamePattern) ?? DEFAULT_GAME_PATTERN
  );
  const isPk = mode === "pk";
  const [selectedDifficulty, setSelectedDifficulty] = useState("normal");
  const [activeSession, setActiveSession] = useState<{
    id: string;
    degree: string;
    pattern: string | null;
    currentLevelId: string;
  } | null>(null);
  const [hasActiveLevelSession, setHasActiveLevelSession] = useState(false);
  const [isPending, startTransition] = useTransition();

  // Re-check for active session when degree/pattern selection changes (single mode only)
  useEffect(() => {
    if (!open || isPk) return;
    let cancelled = false;
    const patternValue = isWordSentence ? selectedPattern : null;
    checkActiveSessionAction(gameId, selectedDegree, patternValue).then((result) => {
      if (cancelled) return;
      setActiveSession(result.data ?? null);
    });
    return () => { cancelled = true; };
  }, [open, isPk, gameId, selectedDegree, selectedPattern, isWordSentence]);

  // When a specific level is selected, check for active level session (single mode only)
  useEffect(() => {
    if (!open || isPk || !activeSession || !levelId) {
      setHasActiveLevelSession(false);
      return;
    }
    let cancelled = false;
    const patternValue = isWordSentence ? selectedPattern : null;
    checkActiveLevelSessionAction(gameId, selectedDegree, patternValue, levelId).then((result) => {
      if (cancelled) return;
      setHasActiveLevelSession(result.data !== null);
    });
    return () => { cancelled = true; };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- activeSession.id is sufficient, full object would cause infinite loop
  }, [open, isPk, activeSession?.id, levelId, gameId, selectedDegree, selectedPattern, isWordSentence]);

  if (!open) return null;

  const subtitle = isPk
    ? `${levelLabel ?? gameName} · PK 对战`
    : (levelLabel ?? gameName);

  function handlePkStart() {
    startTransition(() => {
      const params = new URLSearchParams({ degree: selectedDegree, difficulty: selectedDifficulty });
      if (isWordSentence) params.set("pattern", selectedPattern);
      if (levelId) params.set("level", levelId);
      router.push(`/hall/play-pk/${gameId}?${params}`);
    });
  }

  function handleStart() {
    startTransition(async () => {
      const params = new URLSearchParams({ degree: selectedDegree });
      if (isWordSentence) params.set("pattern", selectedPattern);
      if (levelId) params.set("level", levelId);
      router.push(`/hall/play-single/${gameId}?${params}`);
    });
  }

  function handleResume() {
    if (!activeSession) return;
    const params = new URLSearchParams({ degree: activeSession.degree });
    if (activeSession.pattern) params.set("pattern", activeSession.pattern);
    params.set("level", levelId ?? activeSession.currentLevelId);
    router.push(`/hall/play-single/${gameId}?${params}`);
  }

  function handleRestart() {
    if (!activeSession) return;
    startTransition(async () => {
      const targetLevelId = levelId ?? activeSession.currentLevelId;
      await restartLevelSessionAction(activeSession.id, targetLevelId);
      setActiveSession(null);
      const params = new URLSearchParams({ degree: selectedDegree });
      if (isWordSentence) params.set("pattern", selectedPattern);
      params.set("level", targetLevelId);
      router.push(`/hall/play-single/${gameId}?${params}`);
    });
  }

  const showResumeButtons = levelId
    ? hasActiveLevelSession
    : activeSession !== null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/[0.38] px-4">
      <div className="flex w-full max-w-[720px] flex-col overflow-hidden rounded-[20px] bg-card shadow-[0_12px_40px_rgba(15,23,42,0.19)] md:min-h-[560px] md:flex-row md:max-h-[90vh]">
        {/* Left: Cover image */}
        <div className="flex h-40 items-center justify-center bg-gradient-to-br from-teal-100 via-sky-100 to-purple-100 md:h-auto md:w-[280px] md:shrink-0 md:rounded-l-[20px]">
          <span className="text-6xl">🎮</span>
        </div>

        {/* Right: Content */}
        <div className="flex flex-1 flex-col justify-between p-6 md:p-8 md:px-9">
          {/* Header */}
          <div className="flex items-start justify-between">
            <div className="flex flex-col gap-1.5">
              <h2 className="text-xl font-bold text-foreground md:text-[22px]">
                选择游戏模式
              </h2>
              <p className="text-sm text-muted-foreground">{subtitle}</p>
            </div>
            <button
              type="button"
              aria-label="关闭"
              className="mt-[-2px] flex h-9 w-9 items-center justify-center rounded-[10px] bg-muted"
              onClick={onClose}
            >
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </button>
          </div>

          {/* Mode options */}
          <div className="flex flex-col gap-3 py-6">
            {modeOptions.map((mode) => {
              const isSelected = selectedDegree === mode.value;
              return (
                <button
                  key={mode.value}
                  type="button"
                  onClick={() => setSelectedDegree(mode.value)}
                  className={`flex items-center gap-4 rounded-[14px] border-2 px-4 py-3.5 md:px-5 md:py-[18px] ${
                    isSelected
                      ? "border-teal-600/30 bg-teal-50"
                      : "border-border bg-card"
                  }`}
                >
                  <div
                    className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-xl ${mode.iconBg}`}
                  >
                    <mode.icon
                      className={`h-[22px] w-[22px] ${mode.iconColor}`}
                    />
                  </div>
                  <div className="flex flex-1 flex-col gap-1 text-left">
                    <span className="text-base font-bold text-foreground">
                      {mode.title}
                    </span>
                    <span className="text-[13px] text-muted-foreground">
                      {mode.desc}
                    </span>
                  </div>
                  <ChevronRight
                    className={`h-[18px] w-[18px] shrink-0 ${isSelected ? "text-teal-600" : "text-muted-foreground"}`}
                  />
                </button>
              );
            })}
          </div>

          {/* Pattern options (Word-Sentence only) */}
          {isWordSentence && (
            <>
              <p className="mb-1 text-xs font-medium text-muted-foreground">
                游戏方式
              </p>
              <div className="flex w-full overflow-hidden border border-border">
                {patternOptions.map(({ value, label, icon: Icon }) => {
                  const isWrite = value === GAME_PATTERNS.WRITE;
                  const isPatternSelected = selectedPattern === value;
                  return (
                    <button
                      key={value}
                      type="button"
                      disabled={!isWrite}
                      onClick={() => isWrite && setSelectedPattern(value as GamePattern)}
                      className={`flex flex-1 items-center justify-center gap-1.5 border-r border-border py-2.5 text-sm font-medium transition-colors last:border-r-0 ${
                        isPatternSelected
                          ? "bg-teal-600 text-white"
                          : isWrite
                            ? "bg-card text-muted-foreground hover:bg-accent"
                            : "bg-muted text-muted-foreground cursor-not-allowed"
                      }`}
                    >
                      <Icon className="h-4 w-4" />
                      {label}
                    </button>
                  );
                })}
              </div>
              {!isPk && <div className="h-px bg-border my-5" />}
            </>
          )}

          {/* Difficulty options (PK mode only) */}
          {isPk && (
            <>
              <p className="mt-4 mb-1 text-xs font-medium text-muted-foreground">
                对手强度
              </p>
              <div className="flex w-full overflow-hidden border border-border">
                {difficultyOptions.map(({ value, label, icon: Icon }) => {
                  const isDiffSelected = selectedDifficulty === value;
                  return (
                    <button
                      key={value}
                      type="button"
                      onClick={() => setSelectedDifficulty(value)}
                      className={`flex flex-1 items-center justify-center gap-1.5 border-r border-border py-2.5 text-sm font-medium transition-colors last:border-r-0 ${
                        isDiffSelected
                          ? "bg-teal-600 text-white"
                          : "bg-card text-muted-foreground hover:bg-accent"
                      }`}
                    >
                      <Icon className="h-4 w-4" />
                      {label}
                    </button>
                  );
                })}
              </div>
              <div className="h-px bg-border my-5" />
            </>
          )}

          {/* Action buttons */}
          <div className="flex h-[48px] items-center gap-3">
            {isPk ? (
              <>
                <button
                  type="button"
                  onClick={onClose}
                  className="flex flex-1 items-center justify-center gap-2 rounded-xl border-[1.5px] border-border bg-card py-3"
                >
                  <X className="h-[18px] w-[18px] text-muted-foreground" />
                  <span className="text-[15px] font-semibold text-muted-foreground">
                    取消游戏
                  </span>
                </button>
                <button
                  type="button"
                  onClick={handlePkStart}
                  disabled={isPending}
                  className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-teal-600 py-3 disabled:opacity-50"
                >
                  <Play className="h-[18px] w-[18px] text-white" />
                  <span className="text-[15px] font-semibold text-white">
                    开始 PK
                  </span>
                </button>
              </>
            ) : showResumeButtons ? (
              <>
                <button
                  type="button"
                  onClick={handleRestart}
                  disabled={isPending}
                  className="flex flex-1 items-center justify-center gap-2 rounded-xl border-[1.5px] border-border bg-card py-3 disabled:opacity-50"
                >
                  <RotateCcw className="h-[18px] w-[18px] text-muted-foreground" />
                  <span className="text-[15px] font-semibold text-muted-foreground">
                    重新开始
                  </span>
                </button>
                <button
                  type="button"
                  onClick={handleResume}
                  className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-teal-600 py-3"
                >
                  <ArrowRight className="h-[18px] w-[18px] text-white" />
                  <span className="text-[15px] font-semibold text-white">
                    继续游戏
                  </span>
                </button>
              </>
            ) : (
              <>
                <button
                  type="button"
                  onClick={onClose}
                  className="flex flex-1 items-center justify-center gap-2 rounded-xl border-[1.5px] border-border bg-card py-3"
                >
                  <X className="h-[18px] w-[18px] text-muted-foreground" />
                  <span className="text-[15px] font-semibold text-muted-foreground">
                    取消游戏
                  </span>
                </button>
                <button
                  type="button"
                  onClick={handleStart}
                  disabled={isPending}
                  className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-teal-600 py-3 disabled:opacity-50"
                >
                  <Play className="h-[18px] w-[18px] text-white" />
                  <span className="text-[15px] font-semibold text-white">
                    开始游戏
                  </span>
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
