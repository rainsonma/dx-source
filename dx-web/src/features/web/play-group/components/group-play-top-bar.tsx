"use client";

import { useState, useEffect, useRef } from "react";
import {
  ArrowLeft,
  Settings,
  RotateCcw,
  Flag,
  Maximize,
  Minimize,
  Clock,
  Trophy,
  Flame,
  SkipForward,
  Check,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage, AvatarBadge } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { useGroupPlayStore } from "../hooks/use-group-play-store";
import { StatRow } from "@/features/web/play-core/components/stat-row";

const actionButtons = [
  { icon: Settings, label: "设置", action: "settings" },
  { icon: RotateCcw, label: "重置", action: "reset" },
  { icon: Flag, label: "反馈", action: "report" },
  { icon: Maximize, label: "全屏", action: "fullscreen" },
] as const;

interface GroupPlayTopBarProps {
  player: { nickname: string; avatarUrl: string | null };
  playerId: string;
  levelName: string;
  levelTimeLimit: number;
  onExit: () => void;
  onReset: () => void;
  onSettings: () => void;
  onReport: () => void;
  onFullscreen: () => void;
  isFullscreen: boolean;
  onLevelTimeUp: () => void;
}

export function GroupPlayTopBar({
  player,
  playerId,
  levelName,
  levelTimeLimit,
  onExit,
  onReset,
  onSettings,
  onReport,
  onFullscreen,
  isFullscreen,
  onLevelTimeUp,
}: GroupPlayTopBarProps) {
  const actionHandlers: Record<string, (() => void) | undefined> = {
    settings: onSettings,
    reset: onReset,
    report: onReport,
    fullscreen: onFullscreen,
  };

  const score = useGroupPlayStore((s) => s.score);
  const comboStreak = useGroupPlayStore((s) => s.combo.streak);
  const skipCount = useGroupPlayStore((s) => s.skipCount);
  const currentIndex = useGroupPlayStore((s) => s.currentIndex);
  const totalItems = useGroupPlayStore((s) => s.contentItems?.length ?? 0);
  const participants = useGroupPlayStore((s) => s.participants);
  const completedPlayerIds = useGroupPlayStore((s) => s.completedPlayerIds);

  const progressPercent =
    totalItems > 0 ? Math.round((currentIndex / totalItems) * 100) : 0;

  const [countdown, setCountdown] = useState(levelTimeLimit * 60);
  const onLevelTimeUpRef = useRef(onLevelTimeUp);
  onLevelTimeUpRef.current = onLevelTimeUp;

  useEffect(() => {
    setCountdown(levelTimeLimit * 60);
    const interval = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          setTimeout(() => onLevelTimeUpRef.current(), 0);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [levelTimeLimit]);

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

        {/* Center: group countdown timer */}
        <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center gap-1.5">
          <Clock
            className={`h-4 w-4 ${countdownLow ? "text-red-500" : "text-teal-600"}`}
          />
          <span
            className={`text-xs font-medium ${countdownLow ? "text-red-500" : "text-muted-foreground"}`}
          >
            Group:
          </span>
          <span
            className={`text-base font-extrabold tracking-tight tabular-nums ${countdownLow ? "text-red-500" : "text-foreground"}`}
          >
            {countdownStr}
          </span>
        </div>

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

      {/* Player panel */}
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

        {/* Member roster */}
        {participants && (
          <div className="border-t border-border px-3 py-2 max-h-24 overflow-y-auto">
            {participants.mode === "group_solo" ? (
              <div className="flex flex-wrap gap-1.5">
                {participants.members.map((m) => {
                  const isCompleted = completedPlayerIds.includes(m.user_id);
                  const isMe = m.user_id === playerId;
                  const color = getAvatarColor(m.user_id);
                  return (
                    <Avatar
                      key={m.user_id}
                      size="sm"
                      className={isMe ? "ring-2 ring-teal-500" : ""}
                      style={{ backgroundColor: color }}
                    >
                      <AvatarFallback
                        className="text-white text-[10px] font-bold"
                        style={{ backgroundColor: color }}
                      >
                        {m.user_name[0]?.toUpperCase()}
                      </AvatarFallback>
                      {isCompleted && (
                        <AvatarBadge className="bg-green-500">
                          <Check className="h-2 w-2 text-white" />
                        </AvatarBadge>
                      )}
                    </Avatar>
                  );
                })}
              </div>
            ) : (
              <div className="space-y-1.5">
                {participants.teams.map((team) => (
                  <div key={team.subgroup_id}>
                    <p className="text-[10px] font-medium text-muted-foreground mb-1">
                      {team.subgroup_name}
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {team.members.map((m) => {
                        const isCompleted = completedPlayerIds.includes(m.user_id);
                        const isMe = m.user_id === playerId;
                        const color = getAvatarColor(m.user_id);
                        return (
                          <Avatar
                            key={m.user_id}
                            size="sm"
                            className={isMe ? "ring-2 ring-teal-500" : ""}
                            style={{ backgroundColor: color }}
                          >
                            <AvatarFallback
                              className="text-white text-[10px] font-bold"
                              style={{ backgroundColor: color }}
                            >
                              {m.user_name[0]?.toUpperCase()}
                            </AvatarFallback>
                            {isCompleted && (
                              <AvatarBadge className="bg-green-500">
                                <Check className="h-2 w-2 text-white" />
                              </AvatarBadge>
                            )}
                          </Avatar>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Stats */}
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
      </div>
    </div>
  );
}
