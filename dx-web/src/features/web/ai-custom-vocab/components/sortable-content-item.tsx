"use client";

import { useState, useRef, useCallback } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  GripVertical,
  Pencil,
  Trash2,
  ChevronUp,
  ChevronDown,
  ChevronsUp, ChevronsDown
} from "lucide-react";
import type { LevelContentItem } from "@/features/web/ai-custom-vocab/actions/course-game.action";

type SortableContentItemProps = {
  item: LevelContentItem;
  index: number;
  actionsDisabled?: boolean;
  onUpdateText?: (itemId: string, content: string, translation: string | null) => Promise<boolean>;
  onInsert?: (itemId: string, direction: "above" | "below") => void;
  onCopy?: (itemId: string, direction: "above" | "below") => void;
  onDelete?: (itemId: string) => void;
};

export function SortableContentItem({ item, index, actionsDisabled, onUpdateText, onInsert, onCopy, onDelete }: SortableContentItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: item.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const itemsArray = Array.isArray(item.items) ? item.items : [];

  const [editingField, setEditingField] = useState<"content" | "translation" | null>(null);
  const [editValue, setEditValue] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const startEditing = useCallback((field: "content" | "translation") => {
    const value = field === "content" ? item.content : (item.translation ?? "");
    setEditingField(field);
    setEditValue(value);
    setTimeout(() => inputRef.current?.focus(), 0);
  }, [item.content, item.translation]);

  const cancelEditing = useCallback(() => {
    setEditingField(null);
    setEditValue("");
  }, []);

  const saveEditing = useCallback(async () => {
    if (!editingField || !onUpdateText) {
      cancelEditing();
      return;
    }

    const trimmed = editValue.trim();
    const originalValue = editingField === "content" ? item.content : (item.translation ?? "");

    if (trimmed === originalValue || (editingField === "content" && trimmed === "")) {
      cancelEditing();
      return;
    }

    const newContent = editingField === "content" ? trimmed : item.content;
    const newTranslation = editingField === "translation" ? (trimmed || null) : item.translation;

    cancelEditing();
    await onUpdateText(item.id, newContent, newTranslation);
  }, [editingField, editValue, item, onUpdateText, cancelEditing]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      saveEditing();
    } else if (e.key === "Escape") {
      cancelEditing();
    }
  }, [saveEditing, cancelEditing]);

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`flex flex-col rounded-xl border bg-card ${
        isDragging
          ? "z-10 border-teal-400 opacity-90 shadow-md"
          : "border-border"
      }`}
    >
      {/* Title bar */}
      <div className="flex items-center justify-between px-2 pt-2">
        <div className="flex items-center gap-1.5">
          <button
            type="button"
            className="shrink-0 cursor-grab touch-none active:cursor-grabbing"
            {...attributes}
            {...listeners}
          >
            <GripVertical className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
          <div className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-teal-50">
            <span className="text-[10px] font-bold text-teal-600">{index}</span>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <span className={`rounded-md px-1.5 py-0.5 text-[10px] font-medium ${
            itemsArray.length > 0
              ? "bg-teal-100 text-teal-600"
              : "bg-muted text-muted-foreground"
          }`}>
            {itemsArray.length > 0 ? "已生成" : "未生成"}
          </span>
          {(onInsert || onCopy || onDelete) && (
            <div className="flex items-center gap-1">
              <button
                type="button"
                aria-label="向上插入"
                disabled={actionsDisabled}
                onClick={() => onInsert?.(item.id, "above")}
                className="flex h-6 w-6 items-center justify-center rounded-md bg-muted hover:bg-accent disabled:opacity-50"
              >
                <ChevronUp className="h-3 w-3 text-muted-foreground" />
              </button>
              <button
                type="button"
                aria-label="向下插入"
                disabled={actionsDisabled}
                onClick={() => onInsert?.(item.id, "below")}
                className="flex h-6 w-6 items-center justify-center rounded-md bg-muted hover:bg-accent disabled:opacity-50"
              >
                <ChevronDown className="h-3 w-3 text-muted-foreground" />
              </button>
              <button
                type="button"
                aria-label="向上复制"
                disabled={actionsDisabled}
                onClick={() => onCopy?.(item.id, "above")}
                className="flex h-6 w-6 items-center justify-center rounded-md bg-muted hover:bg-accent disabled:opacity-50"
              >
                <ChevronsUp className="h-3 w-3 text-muted-foreground" />
              </button>
              <button
                type="button"
                aria-label="向下复制"
                disabled={actionsDisabled}
                onClick={() => onCopy?.(item.id, "below")}
                className="flex h-6 w-6 items-center justify-center rounded-md bg-muted hover:bg-accent disabled:opacity-50"
              >
                <ChevronsDown className="h-3 w-3 text-muted-foreground" />
              </button>
              <button
                type="button"
                aria-label="删除"
                onClick={() => onDelete?.(item.id)}
                className="flex h-6 w-6 items-center justify-center rounded-md bg-red-100 hover:bg-red-200"
              >
                <Trash2 className="h-3 w-3 text-red-500" />
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Content body */}
      <div className="px-3 py-2.5 lg:px-4">
        <div className="flex min-w-0 flex-col gap-0.5">
          {editingField === "content" ? (
            <input
              ref={inputRef}
              type="text"
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onBlur={saveEditing}
              onKeyDown={handleKeyDown}
              className="truncate rounded border border-teal-300 bg-teal-50 px-1 text-sm font-semibold text-foreground outline-none focus:ring-1 focus:ring-teal-400"
            />
          ) : (
            <button
              type="button"
              onClick={() => onUpdateText && startEditing("content")}
              className={`group flex min-w-0 items-center gap-1 text-left ${!onUpdateText ? "cursor-default" : ""}`}
            >
              <span className="truncate text-sm font-semibold text-foreground">
                {item.content}
              </span>
              {onUpdateText && (
                <Pencil className="h-3 w-3 shrink-0 text-muted-foreground group-hover:text-teal-500" />
              )}
            </button>
          )}
          {editingField === "translation" ? (
            <input
              ref={inputRef}
              type="text"
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onBlur={saveEditing}
              onKeyDown={handleKeyDown}
              className="truncate rounded border border-teal-300 bg-teal-50 px-1 text-[11px] text-muted-foreground outline-none focus:ring-1 focus:ring-teal-400"
            />
          ) : (
            <button
              type="button"
              onClick={() => onUpdateText && startEditing("translation")}
              className={`group flex min-w-0 items-center gap-1 text-left ${!onUpdateText ? "cursor-default" : ""}`}
            >
              <span className="truncate text-[11px] text-muted-foreground">
                {item.translation || (onUpdateText ? "添加翻译..." : "")}
              </span>
              {onUpdateText && (
                <Pencil className="h-2.5 w-2.5 shrink-0 text-muted-foreground group-hover:text-teal-400" />
              )}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
