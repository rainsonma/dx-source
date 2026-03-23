"use client";

import { useState } from "react";
import { Clock, AlertTriangle, CheckCircle2, Loader2 } from "lucide-react";
import Link from "next/link";
import { StatCard } from "@/components/in/stat-card";
import { WordTable } from "@/components/in/word-table";
import { Badge } from "@/components/in/badge";
import { DeleteConfirmDialog } from "@/components/in/delete-confirm-dialog";
import type { ColumnConfig } from "@/components/in/word-table";
import type { ReviewItem, ReviewStats } from "@/features/web/user-review/actions/review.action";
import { useReviewList } from "@/features/web/user-review/hooks/use-review-list";
import { formatDate } from "@/lib/format";

interface ReviewContentProps {
  initialItems: ReviewItem[];
  initialCursor: string | null;
  initialStats: ReviewStats;
}

/** Derive urgency level from nextReviewAt */
function getUrgency(nextReviewAt: string | null): { label: string; bg: string; text: string } {
  if (!nextReviewAt) return { label: "正常", bg: "bg-blue-100", text: "text-blue-700" };
  const now = new Date();
  const diff = new Date(nextReviewAt).getTime() - now.getTime();
  const days = diff / (1000 * 60 * 60 * 24);
  if (days < 0) return { label: "紧急", bg: "bg-red-100", text: "text-red-700" };
  if (days < 1) return { label: "较高", bg: "bg-red-100", text: "text-red-700" };
  if (days < 3) return { label: "中等", bg: "bg-amber-100", text: "text-amber-700" };
  return { label: "较低", bg: "bg-teal-100", text: "text-teal-700" };
}

type FlatItem = {
  id: string;
  content: string;
  translation: string | null;
  gameId: string;
  gameName: string;
  lastReviewAt: string | null;
  nextReviewAt: string | null;
  reviewCount: number;
};

/** Flatten nested ReviewItem for WordTable */
function flatten(item: ReviewItem): FlatItem {
  return {
    id: item.id,
    content: item.contentItem.content,
    translation: item.contentItem.translation,
    gameId: item.gameId,
    gameName: item.gameName,
    lastReviewAt: item.lastReviewAt,
    nextReviewAt: item.nextReviewAt,
    reviewCount: item.reviewCount,
  };
}

const columns: ColumnConfig<FlatItem>[] = [
  { key: "content", label: "词汇", width: "flex-1", render: () => null },
  {
    key: "lastReview",
    label: "上次复习",
    width: "w-[100px]",
    render: (item) => <span className="text-xs text-slate-400">{formatDate(item.lastReviewAt)}</span>,
  },
  {
    key: "nextReview",
    label: "下次复习",
    width: "w-[100px]",
    render: (item) => {
      const isOverdue = item.nextReviewAt && new Date(item.nextReviewAt) < new Date();
      return (
        <span className={`text-xs ${isOverdue ? "font-semibold text-red-500" : "text-slate-400"}`}>
          {formatDate(item.nextReviewAt)}
        </span>
      );
    },
  },
  {
    key: "urgency",
    label: "紧急度",
    width: "w-[80px]",
    render: (item) => {
      const u = getUrgency(item.nextReviewAt);
      return <Badge label={u.label} bg={u.bg} text={u.text} />;
    },
  },
  {
    key: "action",
    label: "",
    width: "w-[72px]",
    render: (item) => (
      <Link
        href={`/hall/games/${item.gameId}`}
        className="rounded-md bg-indigo-50 px-2.5 py-1 text-[11px] font-semibold text-indigo-600 hover:bg-indigo-100"
      >
        去复习
      </Link>
    ),
  },
];

export function ReviewContent({ initialItems, initialCursor, initialStats }: ReviewContentProps) {
  const {
    items, isLoading, hasMore, sentinelRef,
    selectedIds, setSelectedIds, stats, deleteOne, deleteSelected,
  } = useReviewList({ initialItems, initialCursor, initialStats });

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [showBatchDelete, setShowBatchDelete] = useState(false);

  const flatItems = items.map(flatten);

  const statCards = [
    { icon: Clock, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(stats.pending), label: "待复习" },
    { icon: AlertTriangle, iconBg: "bg-red-100", iconColor: "text-red-500", value: String(stats.overdue), label: "已逾期" },
    { icon: CheckCircle2, iconBg: "bg-teal-100", iconColor: "text-teal-600", value: String(stats.reviewedToday), label: "今日已复习" },
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
