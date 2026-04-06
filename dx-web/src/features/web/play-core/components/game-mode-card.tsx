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
  Gamepad2,
  Search,
  CheckCircle2,
  Loader2,
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import {
  checkActiveSessionAction,
  restartLevelSessionAction,
} from "@/features/web/play-single/actions/session.action";
import { verifyOpponentAction, invitePkAction } from "@/features/web/play-pk/actions/invite.action";

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
  levels?: { id: string; name: string }[];
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
  levels,
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
  const [selectedPkLevel, setSelectedPkLevel] = useState(levels?.[0]?.id ?? "");
  const [pkTab, setPkTab] = useState<"random" | "specified">("random");
  const [specifiedUsername, setSpecifiedUsername] = useState("");
  const [verifyResult, setVerifyResult] = useState<{
    userId: string;
    nickname: string;
    isOnline: boolean;
    isVip: boolean;
  } | null>(null);
  const [verifyError, setVerifyError] = useState<string | null>(null);
  const [isVerifying, setIsVerifying] = useState(false);
  const [activeSession, setActiveSession] = useState<{
    id: string;
    degree: string;
    pattern: string | null;
    gameLevelId: string;
  } | null>(null);
  const [isPending, startTransition] = useTransition();

  // Re-check for active session when degree/pattern/level selection changes (single mode only)
  useEffect(() => {
    if (!open || isPk || !levelId) return;
    let cancelled = false;
    const patternValue = isWordSentence ? selectedPattern : null;
    checkActiveSessionAction(levelId, selectedDegree, patternValue).then((result) => {
      if (cancelled) return;
      setActiveSession(result.data ?? null);
    });
    return () => { cancelled = true; };
  }, [open, isPk, levelId, selectedDegree, selectedPattern, isWordSentence]);

  if (!open) return null;

  const subtitle = isPk
    ? `${levelLabel ?? gameName} · PK 对战`
    : (levelLabel ?? gameName);

  function handleTabChange(value: string) {
    setPkTab(value as "random" | "specified");
    if (value === "random") {
      setSpecifiedUsername("");
      setVerifyResult(null);
      setVerifyError(null);
    } else {
      setSelectedDifficulty("normal");
    }
  }

  async function handleVerify() {
    if (!specifiedUsername.trim()) return;
    setIsVerifying(true);
    setVerifyResult(null);
    setVerifyError(null);
    const res = await verifyOpponentAction(specifiedUsername.trim());
    setIsVerifying(false);
    if (res.error) {
      setVerifyError(res.error);
      return;
    }
    if (res.data) {
      if (!res.data.is_online) {
        setVerifyError("对方不在线");
      } else if (!res.data.is_vip) {
        setVerifyError("对方会员已过期");
      } else {
        setVerifyResult({
          userId: res.data.user_id,
          nickname: res.data.nickname,
          isOnline: res.data.is_online,
          isVip: res.data.is_vip,
        });
      }
    }
  }

  function handlePkStart() {
    startTransition(async () => {
      if (pkTab === "specified") {
        if (!verifyResult) return;
        const pkLevel = selectedPkLevel || levels?.[0]?.id;
        const res = await invitePkAction({
          gameId,
          gameLevelId: pkLevel || "",
          degree: selectedDegree,
          pattern: isWordSentence ? selectedPattern : null,
          opponentId: verifyResult.userId,
        });
        if (res.error || !res.data) return;
        const params = new URLSearchParams({
          sessionId: res.data.session_id,
        });
        router.push(`/hall/pk-room/${res.data.pk_id}?${params}`);
      } else {
        const params = new URLSearchParams({ degree: selectedDegree, difficulty: selectedDifficulty });
        if (isWordSentence) params.set("pattern", selectedPattern);
        const pkLevel = selectedPkLevel || levels?.[0]?.id;
        if (pkLevel) params.set("level", pkLevel);
        router.push(`/hall/play-pk/${gameId}?${params}`);
      }
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
    params.set("level", levelId ?? activeSession.gameLevelId);
    router.push(`/hall/play-single/${gameId}?${params}`);
  }

  function handleRestart() {
    if (!activeSession) return;
    startTransition(async () => {
      const targetLevelId = levelId ?? activeSession.gameLevelId;
      await restartLevelSessionAction(activeSession.id, targetLevelId);
      setActiveSession(null);
      const params = new URLSearchParams({ degree: selectedDegree });
      if (isWordSentence) params.set("pattern", selectedPattern);
      params.set("level", targetLevelId);
      router.push(`/hall/play-single/${gameId}?${params}`);
    });
  }

  const showResumeButtons = activeSession !== null;

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

          {/* Tabs (PK mode only) */}
          {isPk && (
            <>
              <Tabs value={pkTab} onValueChange={handleTabChange} className="mt-4">
                <TabsList className="w-full">
                  <TabsTrigger value="random" className="flex-1">随机对手</TabsTrigger>
                  <TabsTrigger value="specified" className="flex-1">指定对手</TabsTrigger>
                </TabsList>

                <TabsContent value="random" className="flex flex-col gap-3 pt-3">
                  <div className="flex items-center gap-3">
                    <Flame className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <span className="shrink-0 text-[13px] font-medium text-foreground">对手强度</span>
                    <Select value={selectedDifficulty} onValueChange={setSelectedDifficulty}>
                      <SelectTrigger className="h-9 flex-1 text-sm">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {difficultyOptions.map(({ value, label }) => (
                          <SelectItem key={value} value={value}>{label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  {levels && levels.length > 0 && (
                    <div className="flex items-center gap-3">
                      <Gamepad2 className="h-4 w-4 shrink-0 text-muted-foreground" />
                      <span className="shrink-0 text-[13px] font-medium text-foreground">起始关卡</span>
                      <Select value={selectedPkLevel} onValueChange={setSelectedPkLevel}>
                        <SelectTrigger className="h-9 flex-1 text-sm">
                          <SelectValue placeholder="选择关卡" />
                        </SelectTrigger>
                        <SelectContent>
                          {levels.map((level) => (
                            <SelectItem key={level.id} value={level.id}>{level.name}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="specified" className="flex flex-col gap-3 pt-3">
                  <div className="flex items-center gap-2">
                    <div className="relative flex-1">
                      <Input
                        value={specifiedUsername}
                        onChange={(e) => {
                          setSpecifiedUsername(e.target.value);
                          setVerifyResult(null);
                          setVerifyError(null);
                        }}
                        placeholder="输入对手用户名"
                        className="h-9 pr-[calc(var(--tag-width,0px)+0.5rem)] text-sm"
                      />
                      {verifyResult && (
                        <span className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1 rounded bg-emerald-50 px-1.5 py-0.5 text-[11px] font-medium text-emerald-600">
                          <CheckCircle2 className="h-3 w-3" />
                          {verifyResult.nickname}
                        </span>
                      )}
                      {verifyError && (
                        <span className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1 rounded bg-red-50 px-1.5 py-0.5 text-[11px] font-medium text-red-500">
                          {verifyError}
                        </span>
                      )}
                    </div>
                    <button
                      type="button"
                      onClick={handleVerify}
                      disabled={isVerifying || !specifiedUsername.trim()}
                      className="flex h-9 shrink-0 items-center gap-1.5 rounded-md bg-teal-600 px-3 text-sm font-medium text-white disabled:opacity-50"
                    >
                      {isVerifying ? (
                        <Loader2 className="h-3.5 w-3.5 animate-spin" />
                      ) : (
                        <Search className="h-3.5 w-3.5" />
                      )}
                      搜索
                    </button>
                  </div>
                  {levels && levels.length > 0 && (
                    <div className="flex items-center gap-3">
                      <Gamepad2 className="h-4 w-4 shrink-0 text-muted-foreground" />
                      <span className="shrink-0 text-[13px] font-medium text-foreground">起始关卡</span>
                      <Select value={selectedPkLevel} onValueChange={setSelectedPkLevel}>
                        <SelectTrigger className="h-9 flex-1 text-sm">
                          <SelectValue placeholder="选择关卡" />
                        </SelectTrigger>
                        <SelectContent>
                          {levels.map((level) => (
                            <SelectItem key={level.id} value={level.id}>{level.name}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </TabsContent>
              </Tabs>

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
                  disabled={isPending || (pkTab === "specified" && !verifyResult)}
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
