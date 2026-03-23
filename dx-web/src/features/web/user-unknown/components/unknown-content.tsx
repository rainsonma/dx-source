"use client";

import { useState } from "react";
import { BookOpen, Clock, Layers, Loader2 } from "lucide-react";
import { StatCard } from "@/components/in/stat-card";
import { WordTable } from "@/components/in/word-table";
import { DeleteConfirmDialog } from "@/components/in/delete-confirm-dialog";
import type { ColumnConfig } from "@/components/in/word-table";
import type { UnknownItem, UnknownStats } from "@/features/web/user-unknown/actions/unknown.action";
import { useUnknownList } from "@/features/web/user-unknown/hooks/use-unknown-list";
import { formatDate } from "@/lib/format";

interface UnknownContentProps {
  initialItems: UnknownItem[];
  initialCursor: string | null;
  initialStats: UnknownStats;
}

type FlatItem = { id: string; content: string; translation: string | null; gameName: string; createdAt: string };

/** Flatten nested UnknownItem for WordTable */
function flatten(item: UnknownItem): FlatItem {
  return {
    id: item.id,
    content: item.contentItem.content,
    translation: item.contentItem.translation,
    gameName: item.gameName,
    createdAt: item.createdAt,
  };
}

const columns: ColumnConfig<FlatItem>[] = [
  { key: "content", label: "生词", width: "flex-1", render: () => null },
  {
    key: "gameName",
    label: "来源",
    width: "w-[140px]",
    render: (item) => <span className="text-xs text-slate-500">{item.gameName}</span>,
  },
  {
    key: "createdAt",
    label: "添加时间",
    width: "w-[120px]",
    render: (item) => <span className="text-xs text-slate-400">{formatDate(item.createdAt)}</span>,
  },
];

export function UnknownContent({ initialItems, initialCursor, initialStats }: UnknownContentProps) {
  const {
    items, isLoading, hasMore, sentinelRef,
    selectedIds, setSelectedIds, stats, deleteOne, deleteSelected,
  } = useUnknownList({ initialItems, initialCursor, initialStats });

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [showBatchDelete, setShowBatchDelete] = useState(false);

  const flatItems = items.map(flatten);

  const statCards = [
    { icon: BookOpen, iconBg: "bg-red-100", iconColor: "text-red-500", value: String(stats.total), label: "全部生词" },
    { icon: Clock, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(stats.today), label: "今日添加" },
    { icon: Layers, iconBg: "bg-blue-100", iconColor: "text-blue-500", value: String(stats.lastThreeDays), label: "最近三天" },
  ];

  return (
    <>
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
        {statCards.map((stat) => (
          <StatCard key={stat.label} {...stat} />
        ))}
      </div>

      <WordTable
        items={flatItems}
        columns={columns}
        selectedIds={selectedIds}
        onSelectChange={setSelectedIds}
        onDelete={(id) => setDeleteTarget(id)}
        onDeleteBatch={() => setShowBatchDelete(true)}
      />

      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      )}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      <DeleteConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null); }}
        count={1}
        onConfirm={() => { if (deleteTarget) { deleteOne(deleteTarget); setDeleteTarget(null); } }}
      />
      <DeleteConfirmDialog
        open={showBatchDelete}
        onOpenChange={setShowBatchDelete}
        count={selectedIds.size}
        onConfirm={() => { deleteSelected(); setShowBatchDelete(false); }}
      />
    </>
  );
}
