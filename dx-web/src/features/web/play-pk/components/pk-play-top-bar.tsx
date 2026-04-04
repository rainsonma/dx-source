"use client";

import { useState, useEffect } from "react";
import {
  ArrowLeft,
  Settings,
  Pause,
  Flag,
  Maximize,
  Minimize,
  Trophy,
  Flame,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { usePkPlayStore } from "../hooks/use-pk-play-store";
import type { PkPlayerActionEvent } from "../types/pk-play";

const actionButtons = [
  { icon: Settings, label: "设置", action: "settings" },
  { icon: Pause, label: "暂停", action: "pause" },
  { icon: Flag, label: "反馈", action: "report" },
  { icon: Maximize, label: "全屏", action: "fullscreen" },
] as const;

interface PkPlayTopBarProps {
  player: { nickname: string; avatarUrl: string | null };
  playerId: string;
  levelName: string;
  opponentName: string;
  lastOpponentAction: PkPlayerActionEvent | null;
  onExit: () => void;
  onPause: () => void;
  onSettings: () => void;
  onReport: () => void;
  onFullscreen: () => void;
  isFullscreen: boolean;
}

export function PkPlayTopBar({
  player,
  playerId,
  levelName,
  opponentName,
  lastOpponentAction,
  onExit,
  onPause,
  onSettings,
  onReport,
  onFullscreen,
  isFullscreen,
}: PkPlayTopBarProps) {
  const actionHandlers: Record<string, (() => void) | undefined> = {
    settings: onSettings,
    pause: onPause,
    report: onReport,
    fullscreen: onFullscreen,
  };

  const score = usePkPlayStore((s) => s.score);
  const comboStreak = usePkPlayStore((s) => s.combo.streak);
  const opponentId = usePkPlayStore((s) => s.opponentId);

  const playerAvatarBg = getAvatarColor(playerId);
  const opponentAvatarBg = opponentId ? getAvatarColor(opponentId) : "#6b7280";

  // Flash state for opponent actions
  const [opponentFlash, setOpponentFlash] = useState<{
    key: number;
    text: string | null;
    type: "score" | "skip" | "combo" | null;
  }>({ key: 0, text: null, type: null });

  /* eslint-disable react-hooks/set-state-in-effect -- SSE event handler requires setState in effect */
  useEffect(() => {
    if (!lastOpponentAction) return;
    switch (lastOpponentAction.action) {
      case "skip":
        setOpponentFlash((prev) => ({
          key: prev.key + 1,
          text: "跳过",
          type: "skip",
        }));
        break;
      case "score":
        setOpponentFlash((prev) => ({
          key: prev.key + 1,
          text: "得分",
          type: "score",
        }));
        break;
      case "combo":
        setOpponentFlash((prev) => ({
          key: prev.key + 1,
          text: `连击 x${lastOpponentAction.combo_streak}`,
          type: "combo",
        }));
        break;
    }
  }, [lastOpponentAction]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const flashColorClass =
    opponentFlash.type === "skip"
      ? "text-pink-400"
      : opponentFlash.type === "combo"
        ? "text-orange-500"
        : "text-teal-500";

  return (
    <div className="flex w-full flex-col bg-card border-b border-border">
      <div className="flex items-center justify-between px-4 py-2.5 md:px-6">
        {/* Left: player info */}
        <div className="flex items-center gap-2.5">
          <button
            type="button"
            aria-label="返回"
            onClick={onExit}
            className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-muted"
          >
            <ArrowLeft className="h-[18px] w-[18px] text-muted-foreground" />
          </button>
          <Avatar size="sm" style={{ backgroundColor: playerAvatarBg }}>
            {player.avatarUrl && (
              <AvatarImage src={player.avatarUrl} alt={player.nickname} />
            )}
            <AvatarFallback
              className="text-white text-xs font-bold"
              style={{ backgroundColor: playerAvatarBg }}
            >
              {player.nickname[0]?.toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <span className="text-xs font-semibold text-foreground leading-tight">
              {player.nickname}
            </span>
            <div className="flex items-center gap-1.5">
              <Trophy className="h-3 w-3 text-teal-600" />
              <span className="text-xs font-extrabold text-foreground">{score}</span>
              {comboStreak >= 3 && (
                <>
                  <Flame className="h-3 w-3 text-orange-500" />
                  <span className="text-[10px] font-bold text-orange-500">
                    x{comboStreak}
                  </span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Center: level name */}
        <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-sm font-semibold text-foreground">
          {levelName}
        </span>

        {/* Right: opponent info + action buttons */}
        <div className="flex items-center gap-2.5">
          {/* Action buttons */}
          <div className="flex items-center gap-1">
            {actionButtons.map((btn) => {
              const Icon =
                btn.action === "fullscreen" && isFullscreen ? Minimize : btn.icon;
              return (
                <button
                  key={btn.label}
                  type="button"
                  aria-label={
                    btn.action === "fullscreen" && isFullscreen
                      ? "退出全屏"
                      : btn.label
                  }
                  onClick={actionHandlers[btn.action]}
                  className="flex h-8 w-8 items-center justify-center rounded-lg"
                >
                  <Icon className="h-[18px] w-[18px] text-muted-foreground" />
                </button>
              );
            })}
          </div>

          {/* Opponent info */}
          <div className="flex items-center gap-2">
            <div className="flex flex-col items-end">
              <span className="text-xs font-semibold text-foreground leading-tight">
                {opponentName}
              </span>
              {opponentFlash.text && opponentFlash.key > 0 && (
                <span
                  key={opponentFlash.key}
                  className={`text-[10px] font-bold ${flashColorClass} animate-pulse`}
                >
                  {opponentFlash.text}
                </span>
              )}
            </div>
            <Avatar size="sm" style={{ backgroundColor: opponentAvatarBg }}>
              <AvatarFallback
                className="text-white text-xs font-bold"
                style={{ backgroundColor: opponentAvatarBg }}
              >
                {opponentName[0]?.toUpperCase()}
              </AvatarFallback>
            </Avatar>
          </div>
        </div>
      </div>
    </div>
  );
}
