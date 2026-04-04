"use client";

import { useState } from "react";
import { Crown, ArrowLeft, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { nextLevelAction, endPkAction } from "../actions/session.action";
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
  result: PkLevelCompleteEvent;
  pkId: string;
  gameId: string;
  levelName: string;
  nextLevelId: string | null;
  currentLevelId: string;
}

export function PkPlayResultPanel({
  result,
  pkId,
  gameId,
  levelName,
  nextLevelId,
  currentLevelId,
}: PkPlayResultPanelProps) {
  const router = useRouter();
  const [loadingNext, setLoadingNext] = useState(false);
  const [loadingEnd, setLoadingEnd] = useState(false);

  async function handleNextLevel() {
    setLoadingNext(true);
    try {
      const res = await nextLevelAction(pkId, currentLevelId);
      if (res.error) {
        toast.error(res.error);
        setLoadingNext(false);
      }
      // SSE will navigate both players away
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
              <span className="text-xs font-bold text-foreground">
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
              <span className="text-xs font-bold text-foreground">
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
        {nextLevelId ? (
          <div className="flex w-full gap-2">
            <Button variant="outline" asChild className="flex-1">
              <Link href={`/hall/games/${gameId}`}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                返回
              </Link>
            </Button>
            <Button
              className="flex-1 bg-teal-600 hover:bg-teal-700"
              onClick={handleNextLevel}
              disabled={loadingNext}
            >
              {loadingNext ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              下一关
            </Button>
          </div>
        ) : (
          <Button
            className="w-full bg-teal-600 hover:bg-teal-700"
            onClick={handleEnd}
            disabled={loadingEnd}
          >
            {loadingEnd ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            结束
          </Button>
        )}
      </div>
    </div>
  );
}
