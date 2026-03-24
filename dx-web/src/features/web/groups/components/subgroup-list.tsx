"use client";

import { Plus, Trash2 } from "lucide-react";
import type { Subgroup } from "../types/group";

const subgroupColors = [
  "bg-blue-50",
  "bg-amber-100",
  "bg-green-50",
  "bg-red-100",
  "bg-purple-50",
  "bg-teal-50",
];

function getSubgroupBg(index: number) {
  return subgroupColors[index % subgroupColors.length];
}

interface SubgroupListProps {
  subgroups: Subgroup[];
  isOwner: boolean;
  selectedId: string | null;
  onSelect: (id: string) => void;
  onCreate: () => void;
  onDelete: (id: string) => void;
}

export function SubgroupList({ subgroups, isOwner, selectedId, onSelect, onCreate, onDelete }: SubgroupListProps) {
  return (
    <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
      <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
        <span className="text-[15px] font-semibold text-foreground">群小组（{subgroups.length}）</span>
        {isOwner && (
          <button
            type="button"
            onClick={onCreate}
            className="flex items-center gap-1 rounded-lg bg-teal-600 px-2.5 py-1 text-[11px] font-semibold text-white hover:bg-teal-700"
          >
            <Plus className="h-3 w-3" />
            新建
          </button>
        )}
      </div>
      <div className="flex-1 overflow-y-auto">
        {subgroups.length === 0 && (
          <div className="px-5 py-6 text-center text-xs text-muted-foreground">暂无小组</div>
        )}
        {subgroups.map((sg, i) => (
          <div key={sg.id}>
            {i > 0 && <div className="h-px bg-border" />}
            <button
              type="button"
              onClick={() => onSelect(sg.id)}
              className={`flex w-full items-center gap-3.5 px-5 py-3.5 text-left transition-colors hover:bg-muted/50 ${
                selectedId === sg.id ? "bg-teal-50/50" : ""
              }`}
            >
              <div className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-[10px] ${getSubgroupBg(i)}`}>
                <span className="text-sm font-semibold text-foreground">{sg.name[0]}</span>
              </div>
              <div className="flex flex-1 flex-col gap-0.5">
                <span className="text-[13px] font-semibold text-foreground">{sg.name}</span>
                <span className="text-[11px] text-muted-foreground">{sg.member_count} 名成员</span>
              </div>
              {isOwner && (
                <span
                  role="button"
                  tabIndex={0}
                  onClick={(e) => { e.stopPropagation(); onDelete(sg.id); }}
                  onKeyDown={(e) => { if (e.key === "Enter") { e.stopPropagation(); onDelete(sg.id); } }}
                  className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-red-50 hover:text-red-500"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </span>
              )}
              <span className="text-sm font-bold text-teal-600">{sg.order}</span>
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
