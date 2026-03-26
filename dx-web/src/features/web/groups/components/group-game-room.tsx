"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import useSWR from "swr";
import { toast } from "sonner";
import { Gamepad2, Users, User, Loader2, ArrowLeft, Play, Square } from "lucide-react";
import Link from "next/link";
import { swrMutate } from "@/lib/swr";
import type { GroupDetail } from "../types/group";
import { groupApi } from "../actions/group.action";
import { useGroupEvents } from "../hooks/use-group-events";
import { StartGameDialog } from "./start-game-dialog";

interface GroupGameRoomProps {
  groupId: string;
}

export function GroupGameRoom({ groupId }: GroupGameRoomProps) {
  const router = useRouter();
  const { data: group, isLoading } = useSWR<GroupDetail>(`/api/groups/${groupId}`);

  const [startGameOpen, setStartGameOpen] = useState(false);
  const [forceEnding, setForceEnding] = useState(false);

  const isOwner = group?.is_owner ?? false;

  // SSE: when owner starts, navigate to game play
  useGroupEvents(group ? groupId : null, {
    onGameStart: (event) => {
      router.push(
        `/hall/play/${event.game_id}?groupId=${event.game_group_id}&degree=${event.degree}${event.pattern ? `&pattern=${event.pattern}` : ""}&answerTimeLimit=${event.answer_time_limit}`
      );
    },
  });

  async function handleForceEnd() {
    setForceEnding(true);
    const res = await groupApi.forceEnd(groupId);
    setForceEnding(false);
    if (res.code !== 0) { toast.error(res.message); return; }
    toast.success("游戏已结束");
    swrMutate(`/api/groups/${groupId}`);
  }

  if (isLoading || !group) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
      </div>
    );
  }

  if (!group.current_game_id) {
    return (
      <div className="flex flex-col items-center gap-4">
        <p className="text-sm text-muted-foreground">暂未设置课程游戏</p>
        <Link
          href={`/hall/groups/${groupId}`}
          className="text-sm text-teal-600 hover:underline"
        >
          返回群详情
        </Link>
      </div>
    );
  }

  return (
    <div className="flex w-full max-w-sm flex-col items-center gap-6">
      {/* Back link */}
      <Link
        href={`/hall/groups/${groupId}`}
        className="flex items-center gap-1 self-start text-xs text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        返回群详情
      </Link>

      {/* Group info card */}
      <div className="flex w-full flex-col items-center gap-4 rounded-2xl border border-border bg-card p-6">
        <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-teal-100">
          <Gamepad2 className="h-7 w-7 text-teal-600" />
        </div>

        <div className="flex flex-col items-center gap-1">
          <h2 className="text-lg font-bold text-foreground">{group.name}</h2>
          <span className="text-xs text-muted-foreground">
            {group.member_count} 名成员
          </span>
        </div>

        <div className="h-px w-full bg-border" />

        {/* Game info */}
        <div className="flex w-full items-center gap-3 rounded-xl bg-muted px-4 py-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-teal-100">
            <Gamepad2 className="h-4 w-4 text-teal-600" />
          </div>
          <div className="flex flex-1 flex-col gap-0.5 overflow-hidden">
            <span className="truncate text-sm font-semibold text-foreground">
              {group.current_game_name || "未知游戏"}
            </span>
          </div>
          {group.game_mode === "team" ? (
            <span className="flex items-center gap-1 rounded-full bg-blue-500/10 px-2 py-0.5 text-[10px] font-medium text-blue-600">
              <Users className="h-3 w-3" />
              小组
            </span>
          ) : (
            <span className="flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-600">
              <User className="h-3 w-3" />
              单人
            </span>
          )}
        </div>

        <div className="h-px w-full bg-border" />

        {/* Waiting state */}
        <div className="flex flex-col items-center gap-3 py-4">
          <Loader2 className="h-8 w-8 animate-spin text-teal-500" />
          <p className="text-base font-medium text-muted-foreground">
            准备开始...
          </p>
          <p className="text-xs text-muted-foreground">
            {isOwner ? "选择难度和模式后开始游戏" : "等待群主开始游戏"}
          </p>
        </div>

        {/* Owner controls */}
        {isOwner && (
          group.is_playing ? (
            <button
              type="button"
              onClick={handleForceEnd}
              disabled={forceEnding}
              className="flex w-full items-center justify-center gap-1.5 rounded-[10px] bg-red-500 py-2.5 text-sm font-medium text-white hover:bg-red-600 disabled:opacity-50"
            >
              {forceEnding ? <Loader2 className="h-4 w-4 animate-spin" /> : <Square className="h-4 w-4" />}
              游戏中，强制结束
            </button>
          ) : (
            <button
              type="button"
              onClick={() => setStartGameOpen(true)}
              className="flex w-full items-center justify-center gap-1.5 rounded-[10px] bg-teal-600 py-2.5 text-sm font-medium text-white hover:bg-teal-700"
            >
              <Play className="h-4 w-4" />
              开始游戏
            </button>
          )
        )}
      </div>

      {/* Start game dialog */}
      {isOwner && group.current_game_id && (
        <StartGameDialog
          groupId={groupId}
          open={startGameOpen}
          onOpenChange={setStartGameOpen}
          onStarted={() => swrMutate(`/api/groups/${groupId}`)}
        />
      )}
    </div>
  );
}
