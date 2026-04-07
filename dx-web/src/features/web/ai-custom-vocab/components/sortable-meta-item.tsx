"use client";

import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, Trash2 } from "lucide-react";
import type { LevelMeta } from "@/features/web/ai-custom-vocab/actions/course-game.action";
import { SOURCE_FROM_LABELS } from "@/consts/source-from";
import type { SourceFrom } from "@/consts/source-from";

type SortableMetaItemProps = {
  meta: LevelMeta;
  isSelected?: boolean;
  onClick?: () => void;
  onDelete?: (id: string) => void;
};

function truncate(text: string, max: number) {
  return text.length > max ? text.slice(0, max) + "..." : text;
}

export function SortableMetaItem({
  meta,
  isSelected,
  onClick,
  onDelete,
}: SortableMetaItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: meta.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const sourceLabel =
    SOURCE_FROM_LABELS[meta.sourceFrom as SourceFrom] ?? meta.sourceFrom;

  return (
    <div
      ref={setNodeRef}
      style={style}
      onClick={onClick}
      className={`flex cursor-pointer items-center gap-2.5 rounded-[10px] border-3 px-3.5 py-2 ${
        isDragging
          ? "z-10 border-teal-400 bg-teal-50 opacity-90 shadow-md"
          : isSelected
            ? "border-teal-600 bg-teal-50"
            : "border-transparent bg-muted"
      }`}
    >
      <button
        type="button"
        className="shrink-0 cursor-grab touch-none active:cursor-grabbing"
        {...attributes}
        {...listeners}
      >
        <GripVertical className="h-3.5 w-3.5 text-muted-foreground" />
      </button>

      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="truncate text-sm font-semibold text-foreground">
          {truncate(meta.sourceData, 30)}
        </span>
        {meta.translation && (
          <span className="truncate text-[11px] text-muted-foreground">
            {truncate(meta.translation, 40)}
          </span>
        )}
      </div>

      <span className="ml-auto shrink-0 rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
        {sourceLabel}
      </span>
      <span className={`w-14 shrink-0 rounded-md px-2 py-0.5 text-center text-[11px] font-medium ${
        meta.isBreakDone
          ? "bg-teal-100 text-teal-600"
          : "bg-muted text-muted-foreground"
      }`}>
        {meta.isBreakDone ? "已分解" : "未分解"}
      </span>
      <span className={`w-14 shrink-0 rounded-md px-2 py-0.5 text-center text-[11px] font-medium ${
        meta.isItemDone
          ? "bg-teal-100 text-teal-600"
          : "bg-muted text-muted-foreground"
      }`}>
        {meta.isItemDone ? "已生成" : "未生成"}
      </span>

      {onDelete && (
        <button
          type="button"
          aria-label="删除"
          onClick={() => onDelete(meta.id)}
          className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-red-100"
        >
          <Trash2 className="h-3 w-3 text-red-500" />
        </button>
      )}
    </div>
  );
}
