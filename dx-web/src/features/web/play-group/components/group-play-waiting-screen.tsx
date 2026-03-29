"use client";

import { Loader2, ArrowLeft, Gamepad2, User, Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { getAvatarColor } from "@/lib/avatar";
import Link from "next/link";

interface GroupPlayWaitingScreenProps {
  groupId: string;
  player: { id: string; nickname: string; avatarUrl: string | null };
  gameName: string;
  gameMode: string | null;
  levelName: string;
}

export function GroupPlayWaitingScreen({
  groupId,
  player,
  gameName,
  gameMode,
  levelName,
}: GroupPlayWaitingScreenProps) {
  const avatarBg = getAvatarColor(player.id);

  return (
    <div className="flex h-screen flex-col items-center justify-center px-4 py-12">
      <div className="flex w-full max-w-sm flex-col items-center gap-4 rounded-2xl border border-border bg-card p-6">
        {/* Player avatar */}
        <Avatar className="h-14 w-14" style={{ backgroundColor: avatarBg }}>
          <AvatarFallback
            className="text-lg font-bold text-white"
            style={{ backgroundColor: avatarBg }}
          >
            {player.nickname[0]}
          </AvatarFallback>
        </Avatar>
        <span className="text-sm font-semibold text-foreground">
          {player.nickname}
        </span>

        <div className="h-px w-full bg-border" />

        {/* Game info row */}
        <div className="flex w-full items-center gap-3 rounded-xl bg-muted px-4 py-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-teal-100">
            <Gamepad2 className="h-4 w-4 text-teal-600" />
          </div>
          <div className="flex flex-1 flex-col gap-0.5 overflow-hidden">
            <span className="truncate text-sm font-semibold text-foreground">
              {gameName}
            </span>
            <span className="truncate text-[11px] text-muted-foreground">
              {levelName}
            </span>
          </div>
          {gameMode === "group_team" ? (
            <span className="flex items-center gap-1 rounded-full bg-blue-500/10 px-2 py-0.5 text-[10px] font-medium text-blue-600">
              <Users className="h-3 w-3" />
              小组
            </span>
          ) : gameMode === "group_solo" ? (
            <span className="flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-600">
              <User className="h-3 w-3" />
              单人
            </span>
          ) : null}
        </div>

        <div className="h-px w-full bg-border" />

        {/* Spinner + message */}
        <div className="flex flex-col items-center gap-3 py-4">
          <Loader2 className="h-8 w-8 animate-spin text-teal-500" />
          <p className="text-center text-sm font-medium text-muted-foreground">
            好厉害！请耐心等待其他选手完成...
          </p>
        </div>

        {/* Return button */}
        <Button asChild className="w-full bg-teal-600 hover:bg-teal-700">
          <Link href={`/hall/groups/${groupId}`}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            返回
          </Link>
        </Button>
      </div>
    </div>
  );
}
