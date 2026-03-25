"use client";

import { useState } from "react";
import { Check, X, LogOut, Users, Loader2 } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { GroupMember, Subgroup } from "../types/group";

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

interface MemberListProps {
  isOwner: boolean;
  members: GroupMember[];
  subgroups: Subgroup[];
  selectedUserIds: Set<string>;
  onToggleSelect: (userId: string) => void;
  onAssignToSubgroup: (subgroupId: string) => void;
  onKick: (userId: string) => void;
  onLeave: () => void;
}

export function MemberList({
  isOwner,
  members,
  subgroups,
  selectedUserIds,
  onToggleSelect,
  onAssignToSubgroup,
  onKick,
  onLeave,
}: MemberListProps) {
  const hasSelection = selectedUserIds.size > 0;
  const [leaveOpen, setLeaveOpen] = useState(false);
  const [leaving, setLeaving] = useState(false);

  async function handleLeave() {
    setLeaving(true);
    await onLeave();
    setLeaving(false);
    setLeaveOpen(false);
  }

  return (
    <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
      <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
        <span className="text-sm font-semibold text-foreground">群成员（{members.length}）</span>
        <div className="flex items-center gap-2">
          {isOwner && (
            <DropdownMenu modal={false} open={hasSelection ? undefined : false}>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  disabled={!hasSelection}
                  className="flex items-center gap-1 rounded-lg bg-teal-600 px-2.5 py-1 text-[11px] font-semibold text-white hover:bg-teal-700 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  <Users className="h-3 w-3" />
                  分配到小组{hasSelection ? `（${selectedUserIds.size}）` : ""}
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="max-h-60 overflow-y-auto">
                <DropdownMenuLabel className="text-xs text-muted-foreground">选择小组</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {subgroups.length === 0 ? (
                  <div className="px-2 py-3 text-center text-xs text-muted-foreground">暂无小组，请先创建</div>
                ) : (
                  subgroups.map((sg) => (
                    <DropdownMenuItem
                      key={sg.id}
                      onClick={() => onAssignToSubgroup(sg.id)}
                      className="flex items-center justify-between gap-4"
                    >
                      <span className="text-sm">{sg.name}</span>
                      <span className="text-xs text-muted-foreground">{sg.member_count} 人</span>
                    </DropdownMenuItem>
                  ))
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
          {!isOwner && (
            <button
              type="button"
              onClick={() => setLeaveOpen(true)}
              className="flex items-center gap-1 rounded-lg border border-red-200 px-2.5 py-1 text-[11px] font-semibold text-red-500 hover:bg-red-50"
            >
              <LogOut className="h-3 w-3" />
              退出群组
            </button>
          )}
        </div>
      </div>
      <div className="flex-1 overflow-y-auto">
        {members.map((m, i) => {
          const color = m.is_owner
            ? { bg: "bg-teal-600", text: "text-white" }
            : getAvatarColor(m.user_name);
          const isSelected = selectedUserIds.has(m.user_id);

          return (
            <div key={m.id}>
              <div
                className={`flex items-center gap-3 px-5 ${m.is_owner ? "border-l-[3px] border-teal-600/20 bg-teal-50 py-3" : "py-2.5"} ${isOwner ? "cursor-pointer hover:bg-muted/50" : ""}`}
                onClick={isOwner ? () => onToggleSelect(m.user_id) : undefined}
              >
                {isOwner && (
                  <div
                    className={`h-[18px] w-[18px] shrink-0 rounded ${
                      isSelected
                        ? "flex items-center justify-center bg-teal-600"
                        : "border-[1.5px] border-border bg-card"
                    }`}
                  >
                    {isSelected && <Check className="h-3 w-3 text-white" />}
                  </div>
                )}
                <div className={`flex ${m.is_owner ? "h-10 w-10" : "h-9 w-9"} shrink-0 items-center justify-center rounded-full ${color.bg}`}>
                  <span className={`${m.is_owner ? "text-sm" : "text-xs"} font-semibold ${color.text}`}>{m.user_name[0]}</span>
                </div>
                <div className="flex flex-1 flex-col gap-0.5">
                  <span className="text-[13px] font-semibold text-foreground">{m.user_name}</span>
                  {m.is_owner && <span className="text-[11px] text-muted-foreground">创建者</span>}
                </div>
                {isOwner && (
                  <button
                    type="button"
                    onClick={(e) => { e.stopPropagation(); onKick(m.user_id); }}
                    className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-red-50 hover:text-red-500"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
              {i < members.length - 1 && <div className="h-px bg-border" />}
            </div>
          );
        })}
      </div>

      {/* Leave confirmation */}
      <AlertDialog open={leaveOpen} onOpenChange={setLeaveOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认退出群组</AlertDialogTitle>
            <AlertDialogDescription>
              退出后将无法查看群组内容，需要重新申请加入。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleLeave} disabled={leaving}>
              {leaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认退出
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
