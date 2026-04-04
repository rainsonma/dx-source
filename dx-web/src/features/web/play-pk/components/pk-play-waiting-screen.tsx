"use client";

import { Loader2 } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import type { PkPlayerActionEvent } from "../types/pk-play";

interface PkPlayWaitingScreenProps {
  opponentId: string | null;
  opponentName: string;
  lastOpponentAction: PkPlayerActionEvent | null;
}

export function PkPlayWaitingScreen({
  opponentId,
  opponentName,
  lastOpponentAction,
}: PkPlayWaitingScreenProps) {
  const avatarBg = opponentId ? getAvatarColor(opponentId) : "#6b7280";

  const actionText =
    lastOpponentAction?.action === "combo"
      ? `连击 x${lastOpponentAction.combo_streak}`
      : lastOpponentAction?.action === "score"
        ? "得分 +1"
        : lastOpponentAction?.action === "skip"
          ? "跳过了一题"
          : null;

  return (
    <div className="flex h-screen flex-col items-center justify-center px-4 py-12">
      <div className="flex w-full max-w-sm flex-col items-center gap-5 rounded-2xl border border-border bg-card p-6">
        {/* Opponent avatar */}
        <Avatar className="h-14 w-14" style={{ backgroundColor: avatarBg }}>
          <AvatarFallback
            className="text-lg font-bold text-white"
            style={{ backgroundColor: avatarBg }}
          >
            {opponentName[0]?.toUpperCase()}
          </AvatarFallback>
        </Avatar>
        <span className="text-sm font-semibold text-foreground">
          {opponentName}
        </span>

        <div className="h-px w-full bg-border" />

        {/* Spinner + message */}
        <div className="flex flex-col items-center gap-3 py-4">
          <Loader2 className="h-8 w-8 animate-spin text-teal-500" />
          <p className="text-center text-sm font-medium text-muted-foreground">
            等待对手完成...
          </p>
        </div>

        {/* Last opponent action */}
        {actionText && (
          <>
            <div className="h-px w-full bg-border" />
            <p className="text-xs text-muted-foreground">
              对手刚刚: <span className="font-semibold text-foreground">{actionText}</span>
            </p>
          </>
        )}
      </div>
    </div>
  );
}
