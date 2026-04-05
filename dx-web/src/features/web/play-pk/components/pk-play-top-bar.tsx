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
  SkipForward,
  Check,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage, AvatarBadge } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { usePkPlayStore } from "../hooks/use-pk-play-store";
import { GroupStatRow } from "@/features/web/play-core/components/group-stat-row";
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
  const comboStreak = usePkPlayStore((s) => s.combo.current);
  const currentIndex = usePkPlayStore((s) => s.currentIndex);
  const totalItems = usePkPlayStore((s) => s.contentItems?.length ?? 0);
  const opponentId = usePkPlayStore((s) => s.opponentId);
  const opponentCompleted = usePkPlayStore((s) => s.opponentCompleted);

  const opponentAvatarBg = opponentId ? getAvatarColor(opponentId) : "#6b7280";

  const [skipFlash, setSkipFlash] = useState({ key: 0, name: null as string | null });
  const [scoreFlash, setScoreFlash] = useState({ key: 0, name: null as string | null });
  const [comboFlash, setComboFlash] = useState({ key: 0, name: null as string | null, text: null as string | null });

  /* eslint-disable react-hooks/set-state-in-effect -- SSE event handlers require setState in effects */
  useEffect(() => {
    if (!lastOpponentAction) return;
    switch (lastOpponentAction.action) {
      case "skip":
        setSkipFlash((prev) => ({ key: prev.key + 1, name: lastOpponentAction.user_name }));
        break;
      case "score":
        setScoreFlash((prev) => ({ key: prev.key + 1, name: lastOpponentAction.user_name }));
        break;
      case "combo":
        setComboFlash((prev) => ({
          key: prev.key + 1,
          name: lastOpponentAction.user_name,
          text: `${lastOpponentAction.user_name} ×${lastOpponentAction.combo_streak}`,
        }));
        break;
    }
  }, [lastOpponentAction]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const progressPercent =
    totalItems > 0 ? Math.round((currentIndex / totalItems) * 100) : 0;

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

        {/* Center: PK label */}
        <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-xs font-medium text-muted-foreground">
          PK
        </span>

        {/* Right: action buttons */}
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
      </div>

      {/* Floating player panel — same as group play */}
      <div className="absolute right-1 top-full z-20 mt-1 w-56 rounded-xl border border-border bg-card shadow-sm md:right-1.5 md:w-64">
        {/* Avatar row: avatar + score + combo */}
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
            <span className="text-xs font-bold text-orange-500">
              连击 × {comboStreak}
            </span>
          )}
        </div>

        {/* Progress bar */}
        <div className="px-3 pb-2 pt-1.5">
          <div className="h-1.5 w-full rounded-sm bg-border">
            <div
              className="h-1.5 rounded-sm bg-teal-600 transition-all duration-300"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>

        {/* Member roster: human + robot */}
        <div className="border-t border-border px-3 py-2">
          <div className="flex flex-wrap gap-1.5">
            {/* Human player */}
            <Avatar
              size="sm"
              className="overflow-visible"
              style={{ backgroundColor: getAvatarColor(playerId) }}
            >
              {player.avatarUrl && (
                <AvatarImage src={player.avatarUrl} alt={player.nickname} />
              )}
              <AvatarFallback
                className="text-white text-[10px] font-bold"
                style={{ backgroundColor: getAvatarColor(playerId) }}
              >
                {player.nickname[0]?.toUpperCase()}
              </AvatarFallback>
            </Avatar>
            {/* Opponent (robot) */}
            <Avatar
              size="sm"
              className="overflow-visible"
              style={{ backgroundColor: opponentAvatarBg }}
            >
              <AvatarFallback
                className="text-white text-[10px] font-bold"
                style={{ backgroundColor: opponentAvatarBg }}
              >
                {opponentName[0]?.toUpperCase()}
              </AvatarFallback>
              {opponentCompleted && (
                <AvatarBadge className="bg-green-500 ring-0">
                  <Check className="h-2 w-2 text-white" />
                </AvatarBadge>
              )}
            </Avatar>
          </div>
        </div>

        {/* Stats: opponent action flashes */}
        <div className="border-t border-border px-3 py-2 space-y-1.5">
          <GroupStatRow
            icon={SkipForward}
            iconClass="text-muted-foreground"
            label="跳过"
            displayText={skipFlash.name}
            flashKey={skipFlash.key}
            flashColorClass="bg-pink-400"
          />
          <GroupStatRow
            icon={Trophy}
            iconClass="text-teal-600"
            label="得分"
            displayText={scoreFlash.name}
            flashKey={scoreFlash.key}
            flashColorClass="bg-teal-400"
          />
          <GroupStatRow
            icon={Flame}
            iconClass="text-orange-500"
            label="连击"
            displayText={comboFlash.text}
            flashKey={comboFlash.key}
            flashColorClass="bg-orange-400"
          />
        </div>
      </div>
    </div>
  );
}
