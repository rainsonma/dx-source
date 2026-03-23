"use client";

import { useState } from "react";
import { CheckCircle2, CalendarDays, CalendarRange, Loader2 } from "lucide-react";
import { StatCard } from "@/components/in/stat-card";
import { WordTable } from "@/components/in/word-table";
import { DeleteConfirmDialog } from "@/components/in/delete-confirm-dialog";
import type { ColumnConfig } from "@/components/in/word-table";
import type { MasterItem, MasterStats } from "@/features/web/user-master/actions/master.action";
import { useMasterList } from "@/features/web/user-master/hooks/use-master-list";
import { formatDate } from "@/lib/format";

interface MasterContentProps {
  initialItems: MasterItem[];
  initialCursor: string | null;
  initialStats: MasterStats;
}

type FlatItem = { id: string; content: string; translation: string | null; gameName: string; masteredAt: string | null };

/** Flatten nested MasterItem for WordTable */
function flatten(item: MasterItem): FlatItem {
  return {
    id: item.id,
    content: item.contentItem.content,
    translation: item.contentItem.translation,
    gameName: item.gameName,
    masteredAt: item.masteredAt,
  };
}

const columns: ColumnConfig<FlatItem>[] = [
  { key: "content", label: "内容", width: "flex-1", render: () => null },
  {
    key: "gameName",
    label: "来源",
    width: "w-[140px]",
    render: (item) => <span className="text-xs text-slate-500">{item.gameName}</span>,
  },
  {
    key: "masteredAt",
    label: "掌握时间",
    width: "w-[120px]",
    render: (item) => <span className="text-xs text-slate-400">{formatDate(item.masteredAt)}</span>,
  },
];

export function MasterContent({ initialItems, initialCursor, initialStats }: MasterContentProps) {
  const {
    items, isLoading, hasMore, sentinelRef,
    selectedIds, setSelectedIds, stats, deleteOne, deleteSelected,
  } = useMasterList({ initialItems, initialCursor, initialStats });

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [showBatchDelete, setShowBatchDelete] = useState(false);

  const flatItems = items.map(flatten);

  const statCards = [
    { icon: CheckCircle2, iconBg: "bg-teal-100", iconColor: "text-teal-600", value: String(stats.total), label: "已掌握总数" },
    { icon: CalendarDays, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(stats.thisWeek), label: "本周掌握" },
    { icon: CalendarRange, iconBg: "bg-blue-100", iconColor: "text-blue-500", value: String(stats.thisMonth), label: "本月掌握" },
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
