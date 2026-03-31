"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import useSWR from "swr";
import { toast } from "sonner";
import { Gamepad2, Users, User, Loader2, CircleArrowLeft, Play, Square } from "lucide-react";
import Link from "next/link";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { swrMutate } from "@/lib/swr";
import type { GroupDetail, RoomMember } from "../types/group";
import { groupApi } from "../actions/group.action";
import { useGroupEvents } from "../hooks/use-group-events";
import { StartGameDialog } from "./start-game-dialog";

const avatarColors = [
  { bg: "bg-teal-100", text: "text-teal-700" },
  { bg: "bg-amber-100", text: "text-amber-700" },
  { bg: "bg-indigo-100", text: "text-indigo-700" },
  { bg: "bg-red-100", text: "text-red-700" },
  { bg: "bg-fuchsia-100", text: "text-fuchsia-700" },
  { bg: "bg-green-100", text: "text-green-700" },
  { bg: "bg-purple-100", text: "text-purple-700" },
];

function getAvatarColor(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = (hash * 31 + name.charCodeAt(i)) | 0;
  }
  return avatarColors[Math.abs(hash) % avatarColors.length];
}

interface GroupGameRoomProps {
  groupId: string;
}

export function GroupGameRoom({ groupId }: GroupGameRoomProps) {
  const router = useRouter();
  const { data: group, isLoading } = useSWR<GroupDetail>(`/api/groups/${groupId}`);

  const [startGameOpen, setStartGameOpen] = useState(false);
  const [forceEnding, setForceEnding] = useState(false);
  const [roomMembers, setRoomMembers] = useState<RoomMember[]>([]);

  const isOwner = group?.is_owner ?? false;
  const allMembersPresent = group ? roomMembers.length >= group.member_count : false;

  // Fetch initial room members on mount
  const fetchRoomMembers = useCallback(async () => {
    const res = await groupApi.roomMembers(groupId);
    if (res.code === 0 && res.data) {
      setRoomMembers(res.data);
    }
  }, [groupId]);

  useEffect(() => {
    if (!group) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- fetches initial data on mount, setState is inside async callback
    fetchRoomMembers();
  }, [group, fetchRoomMembers]);

  // SSE: listen for game start, member join/leave
  useGroupEvents(group ? groupId : null, {
    onGameStart: (event) => {
      try {
        sessionStorage.setItem(
          `group-participants:${event.game_group_id}`,
          JSON.stringify(event.participants)
        );
      } catch {
        // sessionStorage may be unavailable; play-group will still work without roster
      }
      router.push(
        `/hall/play-group/${event.game_id}?groupId=${event.game_group_id}&degree=${event.degree}${event.pattern ? `&pattern=${event.pattern}` : ""}&levelTimeLimit=${event.level_time_limit}&gameMode=${event.game_mode}${event.level_id ? `&level=${event.level_id}` : ""}`
      );
    },
    onRoomMemberJoined: () => {
      // Re-fetch the full list to get user names
      fetchRoomMembers();
    },
    onRoomMemberLeft: () => {
      fetchRoomMembers();
    },
    onDismissed: () => {
      toast.error("群组已被解散");
      router.push("/hall/groups");
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

  // Redirect non-owners away from room during gameplay
  if (group.is_playing && !isOwner) {
    router.replace(`/hall/groups/${groupId}`);
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
      {/* Group info card */}
      <div className="relative flex w-full flex-col items-center gap-4 rounded-2xl border border-border bg-card p-6">
        <Link
          href={`/hall/groups/${groupId}`}
          className="absolute left-4 top-4 flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <CircleArrowLeft className="h-5 w-5" />
        </Link>
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
            {group.start_game_level_name && (
              <span className="truncate text-[11px] text-muted-foreground">
                {group.start_game_level_name}
              </span>
            )}
          </div>
          {group.game_mode === "group_team" ? (
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

        {/* Room members */}
        <div className="flex w-full flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-medium text-muted-foreground">
              已进入教室（{roomMembers.length}/{group.member_count}）
            </span>
            {allMembersPresent && (
              <span className="text-[11px] font-medium text-teal-600">全员到齐</span>
            )}
          </div>
          <div className="flex flex-wrap gap-2">
            {roomMembers.map((m) => {
              const color = getAvatarColor(m.user_name);
              return (
                <div key={m.user_id} className="flex flex-col items-center gap-1">
                  <Avatar className={`h-9 w-9 ${color.bg}`}>
                    <AvatarFallback className={`${color.bg} ${color.text} text-xs font-semibold`}>
                      {m.user_name[0]}
                    </AvatarFallback>
                  </Avatar>
                  <span className="max-w-[48px] truncate text-[10px] text-muted-foreground">
                    {m.user_name}
                  </span>
                </div>
              );
            })}
          </div>
        </div>

        <div className="h-px w-full bg-border" />

        {/* Waiting state */}
        <div className="flex flex-col items-center gap-3 py-2">
          <Loader2 className="h-6 w-6 animate-spin text-teal-500" />
          <p className="text-sm font-medium text-muted-foreground">
            {allMembersPresent
              ? isOwner ? "全员到齐，可以开始" : "等待群主开始"
              : "等待成员进入教室..."}
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
              disabled={!allMembersPresent}
              className="flex w-full items-center justify-center gap-1.5 rounded-[10px] bg-teal-600 py-2.5 text-sm font-medium text-white hover:bg-teal-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Play className="h-4 w-4" />
              开始
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
