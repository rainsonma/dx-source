"use client";

import { useState, useEffect, useRef } from "react";
import {
  ArrowLeft,
  Settings,
  Pause,
  RotateCcw,
  Flag,
  Maximize,
  Minimize,
  Timer,
  Clock,
  ChevronsDown,
  ChevronsUp,
  Trophy,
  Flame,
  SkipForward,
  SquareM,
  Plus,
} from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useGameStore } from "@/features/web/play/hooks/use-game-store";
import { StatRow } from "@/features/web/play/components/stat-row";
import { GAME_DEGREES } from "@/consts/game-degree";

const actionButtons = [
  { icon: Settings, label: "设置", action: "settings" },
  { icon: Pause, label: "暂停", action: "pause" },
  { icon: RotateCcw, label: "重置", action: "reset" },
  { icon: Flag, label: "反馈", action: "report" },
  { icon: Maximize, label: "全屏", action: "fullscreen" },
] as const;

interface GameTopBarProps {
  player: { nickname: string; avatarUrl: string | null };
  levelName: string;
  elapsedTime: string;
  onExit: () => void;
  onReset: () => void;
  onSettings: () => void;
  onPause: () => void;
  onReport: () => void;
  onFullscreen: () => void;
  isFullscreen: boolean;
  levelTimeLimit?: number | null;
  onLevelTimeUp?: () => void;
}

export function GameTopBar({ player, levelName, elapsedTime, onExit, onReset, onSettings, onPause, onReport, onFullscreen, isFullscreen, levelTimeLimit, onLevelTimeUp }: GameTopBarProps) {
  const [playersOpen, setPlayersOpen] = useState(false);
  const actionHandlers: Record<string, (() => void) | undefined> = {
    settings: onSettings,
    pause: onPause,
    reset: onReset,
    report: onReport,
    fullscreen: onFullscreen,
  };
  const score = useGameStore((s) => s.score);
  const comboStreak = useGameStore((s) => s.combo.streak);
  const skipCount = useGameStore((s) => s.skipCount);
  const currentIndex = useGameStore((s) => s.currentIndex);
  const totalItems = useGameStore((s) => s.contentItems?.length ?? 0);
  const degree = useGameStore((s) => s.degree);

  const progressPercent = totalItems > 0 ? Math.round((currentIndex / totalItems) * 100) : 0;

  // Level countdown for group games
  const isGroupGame = !!levelTimeLimit && levelTimeLimit > 0;
  const [countdown, setCountdown] = useState(isGroupGame ? levelTimeLimit * 60 : 0);
  const onLevelTimeUpRef = useRef(onLevelTimeUp);
  onLevelTimeUpRef.current = onLevelTimeUp;

  useEffect(() => {
    if (!isGroupGame) return;
    setCountdown(levelTimeLimit! * 60);
    const interval = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          setTimeout(() => onLevelTimeUpRef.current?.(), 0);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [isGroupGame, levelTimeLimit]);

  const countdownMins = Math.floor(countdown / 60);
  const countdownSecs = countdown % 60;
  const countdownStr = `${String(countdownMins).padStart(2, "0")}:${String(countdownSecs).padStart(2, "0")}`;
  const countdownLow = countdown <= 60;

  return (
    <div className="relative flex w-full flex-col bg-card border-b border-border">
      {/* Nav row */}
      <div className="flex items-center justify-between px-4 py-2.5 md:px-6">
        {/* Left: back + level name */}
        <div className="flex items-center gap-3.5">
          <button
            type="button"
            aria-label="返回"
            onClick={onExit}
            className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-muted"
          >
            <ArrowLeft className="h-[18px] w-[18px] text-muted-foreground" />
          </button>
          <span className="text-sm font-semibold text-foreground">
            {levelName}
          </span>
        </div>

        {/* Center: timer */}
        <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center gap-1.5">
          {isGroupGame ? (
            <>
              <Clock className={`h-4 w-4 ${countdownLow ? "text-red-500" : "text-teal-600"}`} />
              <span className={`text-xs font-medium ${countdownLow ? "text-red-500" : "text-muted-foreground"}`}>Group:</span>
              <span className={`text-base font-extrabold tracking-tight tabular-nums ${countdownLow ? "text-red-500" : "text-foreground"}`}>
                {countdownStr}
              </span>
            </>
          ) : (
            <>
              <Timer className="h-4 w-4 text-teal-600" />
              <span className="text-base font-extrabold tracking-tight text-foreground">
                {elapsedTime}
              </span>
            </>
          )}
        </div>

        {/* Right: action buttons */}
        <div className="flex items-center gap-1">
          {actionButtons.map((btn) => {
            const Icon = btn.action === "fullscreen" && isFullscreen ? Minimize : btn.icon;
            return (
              <button
                key={btn.label}
                type="button"
                aria-label={btn.action === "fullscreen" && isFullscreen ? "退出全屏" : btn.label}
                onClick={actionHandlers[btn.action]}
                className="flex h-8 w-8 items-center justify-center rounded-lg"
              >
                <Icon className="h-[18px] w-[18px] text-muted-foreground" />
              </button>
            );
          })}
        </div>
      </div>

      {/* Player panel — floating collapsible */}
      <Collapsible
        open={playersOpen}
        onOpenChange={setPlayersOpen}
        className="absolute right-1 top-full z-20 mt-1 w-56 rounded-xl border border-border bg-card shadow-sm md:right-1.5 md:w-64"
      >
        {/* Top row: avatar + score + combo + degree badge */}
        <div className="flex items-center gap-2.5 px-3 pt-2">
          <Avatar size="sm" className="bg-teal-600">
            {player.avatarUrl && (
              <AvatarImage src={player.avatarUrl} alt={player.nickname} />
            )}
            <AvatarFallback className="bg-teal-600 text-white text-xs font-bold">
              {player.nickname[0]?.toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <span className="text-sm font-extrabold text-foreground">{score}</span>
          {comboStreak >= 3 && (
            <span className="text-xs font-bold text-orange-500">连击 × {comboStreak}</span>
          )}
          {degree === GAME_DEGREES.PRACTICE ? (
            <span className="ml-auto inline-flex items-center gap-1 rounded-md bg-teal-50 px-2 py-0.5 text-[11px] font-semibold text-teal-700">
              <SquareM className="h-3 w-3" />
              练习
            </span>
          ) : (
            <span className="ml-auto inline-flex items-center gap-1 rounded-md bg-orange-50 px-2 py-0.5 text-[11px] font-semibold text-orange-600">
              <Plus className="h-3 w-3" />
              1v1
            </span>
          )}
        </div>

        {/* Progress bar + centered toggle */}
        <CollapsibleTrigger className="flex w-full flex-col items-center gap-1 px-3 pb-2 pt-1.5">
          <div className="h-1.5 w-full rounded-sm bg-border">
            <div
              className="h-1.5 rounded-sm bg-teal-600 transition-all duration-300"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
          {playersOpen ? (
            <ChevronsUp className="h-3.5 w-3.5 text-muted-foreground" />
          ) : (
            <ChevronsDown className="h-3.5 w-3.5 text-muted-foreground" />
          )}
        </CollapsibleTrigger>

        {/* Expanded: player 2 */}
        <CollapsibleContent>
          <div className="h-px w-full bg-border" />

          {/* Player 2 (solo mode) */}
          <div className="flex items-center gap-2.5 px-3 py-2.5">
            <Avatar size="sm" className="border border-border bg-muted">
              <AvatarFallback className="bg-muted text-muted-foreground text-xs font-bold">
                ?
              </AvatarFallback>
            </Avatar>
            <span className="text-xs text-muted-foreground">单人模式</span>
            <button
              type="button"
              className="ml-auto rounded-lg border border-border bg-card px-2.5 py-0.5"
            >
              <span className="text-[11px] font-medium text-muted-foreground">
                邀请
              </span>
            </button>
          </div>
        </CollapsibleContent>

        {/* Stats: always visible below player panel */}
        <div className="border-t border-border px-3 py-2 space-y-1.5">
          <StatRow
            icon={SkipForward}
            iconClass="text-muted-foreground"
            label="跳过"
            value={skipCount}
            valueClass="text-foreground"
            flashColorClass="bg-pink-400"
          />
          <StatRow
            icon={Trophy}
            iconClass="text-teal-600"
            label="得分"
            value={score}
            valueClass="text-foreground"
            flashColorClass="bg-teal-400"
          />
          <StatRow
            icon={Flame}
            iconClass="text-orange-500"
            label="连击"
            value={comboStreak >= 3 ? comboStreak : 0}
            valueClass="text-orange-500"
            flashColorClass="bg-orange-400"
          />
        </div>
      </Collapsible>
    </div>
  );
}
