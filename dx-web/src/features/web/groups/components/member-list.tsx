"use client";

import { X, LogOut } from "lucide-react";
import type { GroupMember } from "../types/group";

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
  groupId: string;
  isOwner: boolean;
  members: GroupMember[];
  onKick: (userId: string) => void;
  onLeave: () => void;
}

export function MemberList({ isOwner, members, onKick, onLeave }: MemberListProps) {
  return (
    <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
      <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
        <span className="text-sm font-semibold text-foreground">群成员（{members.length}）</span>
        {!isOwner && (
          <button
            type="button"
            onClick={onLeave}
            className="flex items-center gap-1 rounded-lg border border-red-200 px-2.5 py-1 text-[11px] font-semibold text-red-500 hover:bg-red-50"
          >
            <LogOut className="h-3 w-3" />
            退出群组
          </button>
        )}
      </div>
      <div className="flex-1 overflow-y-auto">
        {members.map((m, i) => {
          const color = m.is_owner
            ? { bg: "bg-teal-600", text: "text-white" }
            : getAvatarColor(m.user_name);

          return (
            <div key={m.id}>
              {m.is_owner ? (
                <div className="flex items-center gap-3 border-l-[3px] border-teal-600/20 bg-teal-50 px-5 py-3">
                  <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${color.bg}`}>
                    <span className={`text-sm font-semibold ${color.text}`}>{m.user_name[0]}</span>
                  </div>
                  <div className="flex flex-1 flex-col gap-0.5">
                    <span className="text-[13px] font-semibold text-foreground">{m.user_name}</span>
                    <span className="text-[11px] text-muted-foreground">创建者</span>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-3 px-5 py-2.5">
                  <div className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full ${color.bg}`}>
                    <span className={`text-xs font-semibold ${color.text}`}>{m.user_name[0]}</span>
                  </div>
                  <div className="flex flex-1 flex-col gap-0.5">
                    <span className="text-[13px] font-semibold text-foreground">{m.user_name}</span>
                  </div>
                  {isOwner && (
                    <button
                      type="button"
                      onClick={() => onKick(m.user_id)}
                      className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-red-50 hover:text-red-500"
                    >
                      <X className="h-3.5 w-3.5" />
                    </button>
                  )}
                </div>
              )}
              {i < members.length - 1 && <div className="h-px bg-border" />}
            </div>
          );
        })}
      </div>
    </div>
  );
}
