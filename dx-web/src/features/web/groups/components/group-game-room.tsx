"use client";

import { useRouter } from "next/navigation";
import useSWR from "swr";
import { Gamepad2, Users, User, Loader2, ArrowLeft } from "lucide-react";
import Link from "next/link";
import type { GroupDetail } from "../types/group";
import { useGroupEvents } from "../hooks/use-group-events";

interface GroupGameRoomProps {
  groupId: string;
}

export function GroupGameRoom({ groupId }: GroupGameRoomProps) {
  const router = useRouter();
  const { data: group, isLoading } = useSWR<GroupDetail>(`/api/groups/${groupId}`);

  // SSE: when owner starts, navigate to game play
  useGroupEvents(group ? groupId : null, {
    onGameStart: (event) => {
      router.push(
        `/hall/play/${event.game_id}?groupId=${event.game_group_id}&degree=${event.degree}${event.pattern ? `&pattern=${event.pattern}` : ""}&answerTimeLimit=${event.answer_time_limit}`
      );
    },
  });

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
            等待群主开始游戏
          </p>
        </div>
      </div>
    </div>
  );
}
