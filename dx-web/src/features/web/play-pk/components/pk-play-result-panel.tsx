"use client";

import { useState } from "react";
import { Crown, ArrowLeft, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { nextPkLevelAction, endPkAction } from "../actions/session.action";
import type { PkLevelCompleteEvent } from "../types/pk-play";

/** Teal palette for podium columns — 1st (winner) and 2nd (loser) */
const PODIUM_STYLES = [
  {
    colWidth: "w-[96px]",
    avatarSize: "h-[48px] w-[48px]",
    nameColor: "text-[#2dd4bf]",
    nameSize: "text-sm",
    barHeight: "h-[90px]",
    barGradient: "from-[#0d9488] to-[#0f766e]",
    rankSize: "text-2xl",
    rankColor: "text-[#2dd4bf]",
  },
  {
    colWidth: "w-[84px]",
    avatarSize: "h-[38px] w-[38px]",
    nameColor: "text-[#5eead4]",
    nameSize: "text-xs",
    barHeight: "h-[60px]",
    barGradient: "from-[#115e59] to-[#134e4a]",
    rankSize: "text-xl",
    rankColor: "text-[#5eead4]",
  },
];

interface PkPlayResultPanelProps {
  result: PkLevelCompleteEvent | null;
  pkId: string;
  gameId: string;
  levelName: string;
  nextLevelId: string | null;
}

export function PkPlayResultPanel({
  result,
  pkId,
  gameId,
  levelName,
  nextLevelId,
}: PkPlayResultPanelProps) {
  const router = useRouter();
  const [loadingNext, setLoadingNext] = useState(false);
  const [loadingEnd, setLoadingEnd] = useState(false);

  async function handleNextLevel() {
    setLoadingNext(true);
    try {
      const res = await nextPkLevelAction(pkId);
      if (res.error) {
        toast.error(res.error);
        setLoadingNext(false);
        return;
      }
      if (res.data) {
        // For robot PK, navigate directly. For specified PK, the backend
        // broadcasts pk_next_level which the shell's onNextLevel handler
        // picks up — but we also navigate here so the clicking player
        // doesn't wait for SSE. Duplicate navigation is harmless.
        const store = await import("../hooks/use-pk-play-store").then(m => m.usePkPlayStore.getState());
        const degree = store.degree ?? "";
        const pattern = store.pattern;
        const difficulty = store.difficulty ?? "";
        const params = new URLSearchParams({ degree, level: res.data.game_level_id });
        if (pattern) params.set("pattern", pattern);
        if (difficulty) params.set("difficulty", difficulty);
        // For specified PK (session_id set), include pkId+sessionId so play page uses specified flow
        if (res.data.session_id && res.data.pk_id) {
          params.set("pkId", res.data.pk_id);
          params.set("sessionId", res.data.session_id);
        }
        router.push(`/hall/play-pk/${gameId}?${params}`);
      }
    } catch {
      toast.error("进入下一关失败");
      setLoadingNext(false);
    }
  }

  async function handleEnd() {
    setLoadingEnd(true);
    try {
      await endPkAction(pkId);
      router.push(`/hall/games/${gameId}`);
    } catch {
      toast.error("结束PK失败");
      setLoadingEnd(false);
    }
  }

  // While waiting for SSE result
  if (!result) {
    return (
      <div className="flex h-screen flex-col items-center justify-center px-4 py-12">
        <div className="flex w-full max-w-sm flex-col items-center gap-5 rounded-2xl border border-border bg-card p-6">
          <Loader2 className="h-8 w-8 animate-spin text-teal-500" />
          <p className="text-sm font-medium text-muted-foreground">
            等待结果...
          </p>
        </div>
      </div>
    );
  }

  // Sort participants: winner first, then loser
  const sorted = [...result.participants].sort((a, b) => b.score - a.score);
  const winner = sorted[0];
  const loser = sorted[1];

  return (
    <div className="flex h-screen flex-col items-center justify-center px-4 py-12">
      <div className="flex w-full max-w-sm flex-col items-center gap-4 rounded-2xl border border-border bg-card p-6">
        <h2 className="text-lg font-bold text-foreground">PK 结果</h2>
        <p className="text-sm text-muted-foreground">{levelName}</p>

        {/* Podium: 2 players */}
        <div className="flex items-end justify-center gap-3">
          {/* Winner (1st) */}
          {winner && (
            <div className={`flex flex-col items-center ${PODIUM_STYLES[0].colWidth}`}>
              <Crown className="mb-0.5 h-[18px] w-[18px] text-amber-400" />
              <Avatar
                className={PODIUM_STYLES[0].avatarSize}
                style={{ backgroundColor: getAvatarColor(winner.user_id) }}
              >
                <AvatarFallback
                  className="text-white font-bold text-sm"
                  style={{ backgroundColor: getAvatarColor(winner.user_id) }}
                >
                  {winner.user_name[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span
                className={`mt-1 truncate text-center font-medium leading-tight ${PODIUM_STYLES[0].nameColor} ${PODIUM_STYLES[0].nameSize} max-w-full`}
              >
                {winner.user_name}
              </span>
              <span className="mt-2.5 text-xs font-bold text-foreground">
                {winner.score} 分
              </span>
              <div
                className={`mt-1.5 w-full rounded-t-md bg-gradient-to-b ${PODIUM_STYLES[0].barGradient} ${PODIUM_STYLES[0].barHeight} flex items-center justify-center`}
              >
                <span
                  className={`font-bold ${PODIUM_STYLES[0].rankSize} ${PODIUM_STYLES[0].rankColor}`}
                >
                  1
                </span>
              </div>
            </div>
          )}

          {/* Loser (2nd) */}
          {loser && (
            <div className={`flex flex-col items-center ${PODIUM_STYLES[1].colWidth}`}>
              <Avatar
                className={PODIUM_STYLES[1].avatarSize}
                style={{ backgroundColor: getAvatarColor(loser.user_id) }}
              >
                <AvatarFallback
                  className="text-white font-bold text-sm"
                  style={{ backgroundColor: getAvatarColor(loser.user_id) }}
                >
                  {loser.user_name[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span
                className={`mt-1 truncate text-center font-medium leading-tight ${PODIUM_STYLES[1].nameColor} ${PODIUM_STYLES[1].nameSize} max-w-full`}
              >
                {loser.user_name}
              </span>
              <span className="mt-2.5 text-xs font-bold text-foreground">
                {loser.score} 分
              </span>
              <div
                className={`mt-1.5 w-full rounded-t-md bg-gradient-to-b ${PODIUM_STYLES[1].barGradient} ${PODIUM_STYLES[1].barHeight} flex items-center justify-center`}
              >
                <span
                  className={`font-bold ${PODIUM_STYLES[1].rankSize} ${PODIUM_STYLES[1].rankColor}`}
                >
                  2
                </span>
              </div>
            </div>
          )}
        </div>

        {/* Action buttons */}
        <div className="h-px w-full bg-border" />
        <div className="flex w-full gap-2">
          <Button
            variant="outline"
            className="flex-1"
            onClick={handleEnd}
            disabled={loadingEnd}
          >
            {loadingEnd ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ArrowLeft className="mr-2 h-4 w-4" />}
            结束
          </Button>
          {nextLevelId && (
            <Button
              className="flex-1 bg-teal-600 hover:bg-teal-700"
              onClick={handleNextLevel}
              disabled={loadingNext}
            >
              {loadingNext ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              下一关
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
