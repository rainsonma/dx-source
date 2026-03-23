"use client";

import { Trash2 } from "lucide-react";

export interface WordRow {
  id: string;
  content: string;
  translation: string | null;
}

export interface ColumnConfig<T extends WordRow> {
  key: string;
  label: string;
  width: string;
  render: (item: T) => React.ReactNode;
}

interface WordTableProps<T extends WordRow> {
  items: T[];
  columns: ColumnConfig<T>[];
  selectedIds: Set<string>;
  onSelectChange: (ids: Set<string>) => void;
  onDelete: (id: string) => void;
  onDeleteBatch: () => void;
}

/** Generic word table with checkboxes, single and batch delete */
export function WordTable<T extends WordRow>({
  items,
  columns,
  selectedIds,
  onSelectChange,
  onDelete,
  onDeleteBatch,
}: WordTableProps<T>) {
  const allSelected = items.length > 0 && selectedIds.size === items.length;

  /** Toggle all checkboxes */
  const handleToggleAll = () => {
    if (allSelected) {
      onSelectChange(new Set());
    } else {
      onSelectChange(new Set(items.map((i) => i.id)));
    }
  };

  /** Toggle a single checkbox */
  const handleToggle = (id: string) => {
    const next = new Set(selectedIds);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    onSelectChange(next);
  };

  return (
    <div className="overflow-x-auto">
      {/* Batch toolbar */}
      {selectedIds.size > 0 && (
        <div className="mb-2 flex items-center gap-3 rounded-lg bg-red-50 px-4 py-2">
          <span className="text-sm text-red-700">
            已选择 {selectedIds.size} 项
          </span>
          <button
            onClick={onDeleteBatch}
            className="rounded-md bg-red-500 px-3 py-1 text-xs font-medium text-white hover:bg-red-600"
          >
            删除
          </button>
        </div>
      )}

      <div className="min-w-[600px] overflow-hidden rounded-xl border border-border bg-card">
        {/* Header */}
        <div className="flex items-center gap-4 bg-muted px-5 py-3.5">
          <input
            type="checkbox"
            checked={allSelected}
            onChange={handleToggleAll}
            className="h-4 w-4 shrink-0 rounded border-border"
          />
          <span className="flex-1 text-xs font-semibold text-muted-foreground">
            {columns[0]?.label ?? "词汇"}
          </span>
          {columns.slice(1).map((col) => (
            <span
              key={col.key}
              className={`${col.width} text-xs font-semibold text-muted-foreground`}
            >
              {col.label}
            </span>
          ))}
          <div className="w-5" />
        </div>

        {/* Rows */}
        {items.map((item) => (
          <div
            key={item.id}
            className="flex items-center gap-4 border-t border-border px-5 py-3.5"
          >
            <input
              type="checkbox"
              checked={selectedIds.has(item.id)}
              onChange={() => handleToggle(item.id)}
              className="h-4 w-4 shrink-0 rounded border-border"
            />
            <div className="flex flex-1 flex-col gap-0.5">
              <span className="text-sm font-semibold text-foreground">
                {item.content}
              </span>
              <span className="text-xs text-muted-foreground">
                {item.translation ?? ""}
              </span>
            </div>
            {columns.slice(1).map((col) => (
              <div key={col.key} className={col.width}>
                {col.render(item)}
              </div>
            ))}
            <Trash2
              className="h-[18px] w-[18px] shrink-0 cursor-pointer text-muted-foreground hover:text-foreground"
              onClick={() => onDelete(item.id)}
            />
          </div>
        ))}

        {/* Empty state */}
        {items.length === 0 && (
          <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
            暂无数据
          </div>
        )}
      </div>
    </div>
  );
}
