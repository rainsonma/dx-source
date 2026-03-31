"use client";

import Link from "next/link";
import { Plus, ChevronRight } from "lucide-react";
import type { Group } from "../types/group";

type GroupVariant = "teal" | "amber" | "indigo" | "red" | "green" | "purple";

const variantClasses: Record<GroupVariant, { avatarBg: string; avatarColor: string; tagBg: string; tagColor: string }> = {
  teal: { avatarBg: "bg-teal-100", avatarColor: "text-teal-700", tagBg: "bg-teal-50", tagColor: "text-teal-600" },
  amber: { avatarBg: "bg-amber-100", avatarColor: "text-amber-700", tagBg: "bg-amber-50", tagColor: "text-amber-700" },
  indigo: { avatarBg: "bg-indigo-100", avatarColor: "text-indigo-700", tagBg: "bg-indigo-50", tagColor: "text-indigo-700" },
  red: { avatarBg: "bg-red-100", avatarColor: "text-red-700", tagBg: "bg-red-50", tagColor: "text-red-700" },
  green: { avatarBg: "bg-green-50", avatarColor: "text-green-700", tagBg: "bg-green-50", tagColor: "text-green-700" },
  purple: { avatarBg: "bg-purple-50", avatarColor: "text-purple-700", tagBg: "bg-purple-50", tagColor: "text-purple-700" },
};

const variants: GroupVariant[] = ["teal", "amber", "indigo", "red", "green", "purple"];

function hashToVariant(str: string): GroupVariant {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash * 31 + str.charCodeAt(i)) | 0;
  }
  return variants[Math.abs(hash) % variants.length];
}

interface GroupCardProps {
  group: Group;
  isMember?: boolean;
  highlighted?: boolean;
  onJoin?: () => void;
}

export function GroupCard({ group, isMember = true, highlighted = false, onJoin }: GroupCardProps) {
  const variant = hashToVariant(group.id);
  const v = variantClasses[variant];
  const letter = group.name[0];

  const cardContent = (
    <div
      className={`flex flex-col gap-3.5 rounded-[14px] p-5 transition-colors hover:shadow-sm ${
        highlighted
          ? "border-2 border-teal-600 bg-teal-50/50"
          : "border border-border bg-card"
      }`}
    >
      {/* Header */}
      <div className="flex items-center gap-3">
        <div
          className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-xl ${v.avatarBg}`}
        >
          <span className={`text-lg font-bold ${v.avatarColor}`}>{letter}</span>
        </div>
        <div className="flex flex-1 flex-col gap-0.5">
          <span className="text-[15px] font-semibold text-foreground">{group.name}</span>
          <div className="flex items-center gap-1.5">
            <div className={`flex h-[18px] w-[18px] items-center justify-center rounded-full ${v.avatarBg}`}>
              <span className={`text-[8px] font-semibold ${v.avatarColor}`}>{group.owner_name[0]}</span>
            </div>
            <span className="text-xs font-medium text-muted-foreground">{group.owner_name}</span>
            <span className="text-xs text-muted-foreground">·</span>
            <span className="text-xs text-muted-foreground">{group.member_count} 人</span>
          </div>
        </div>
        {isMember ? (
          <span className="rounded-md bg-teal-600/10 px-2.5 py-1 text-[11px] font-medium text-teal-600">已加入</span>
        ) : group.has_applied ? (
          <span className="rounded-md bg-amber-500/10 px-2.5 py-1 text-[11px] font-medium text-amber-600">申请审核中...</span>
        ) : (
          <button
            type="button"
            onClick={(e) => { e.preventDefault(); onJoin?.(); }}
            className="flex items-center gap-1 rounded-md bg-muted px-2.5 py-1 text-[11px] font-medium text-muted-foreground"
          >
            <Plus className="h-[11px] w-[11px]" />
            加入
          </button>
        )}
      </div>

      {group.description && (
        <p className="line-clamp-2 text-[13px] leading-relaxed text-muted-foreground">{group.description}</p>
      )}

      <div className="flex items-center justify-end">
        <div className="flex h-7 w-7 items-center justify-center rounded-full bg-slate-400/10">
          <ChevronRight className="h-4 w-4 text-muted-foreground" />
        </div>
      </div>
    </div>
  );

  if (isMember) {
    return <Link href={`/hall/groups/${group.id}`} className="block">{cardContent}</Link>;
  }
  return <div className="block">{cardContent}</div>;
}
