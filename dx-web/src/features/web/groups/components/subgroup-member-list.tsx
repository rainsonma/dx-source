"use client";

import { X } from "lucide-react";
import type { SubgroupMember } from "../types/group";

const avatarColors = [
  { bg: "bg-teal-100", text: "text-teal-700" },
  { bg: "bg-indigo-100", text: "text-indigo-700" },
  { bg: "bg-red-100", text: "text-red-700" },
  { bg: "bg-fuchsia-100", text: "text-fuchsia-700" },
  { bg: "bg-amber-100", text: "text-amber-700" },
  { bg: "bg-green-100", text: "text-green-700" },
];

function getAvatarColor(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = (hash * 31 + name.charCodeAt(i)) | 0;
  }
  return avatarColors[Math.abs(hash) % avatarColors.length];
}

interface SubgroupMemberListProps {
  members: SubgroupMember[];
  isOwner: boolean;
  onRemove: (userId: string) => void;
}

export function SubgroupMemberList({ members, isOwner, onRemove }: SubgroupMemberListProps) {
  return (
    <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
      <div className="border-b border-border px-5 py-3.5">
        <span className="text-sm font-semibold text-foreground">组成员（{members.length}）</span>
      </div>
      <div className="flex-1 overflow-y-auto">
        {members.length === 0 && (
          <div className="px-5 py-6 text-center text-xs text-muted-foreground">暂无成员</div>
        )}
        {members.map((m, i) => {
          const color = getAvatarColor(m.user_name);
          return (
            <div key={m.id}>
              {i > 0 && <div className="h-px bg-border" />}
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
                    onClick={() => onRemove(m.user_id)}
                    className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-red-50 hover:text-red-500"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
