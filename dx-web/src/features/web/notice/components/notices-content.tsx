"use client";

import { useState } from "react";
import { Plus, Loader2, Megaphone } from "lucide-react";
import type { NoticeItem as NoticeItemType } from "@/features/web/notice/actions/notice.action";
import { useNoticeList } from "@/features/web/notice/hooks/use-notice-list";
import { NoticeItem } from "@/features/web/notice/components/notice-item";
import { PublishNoticeModal } from "@/features/web/notice/components/publish-notice-modal";
import { EditNoticeModal } from "@/features/web/notice/components/edit-notice-modal";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface NoticesContentProps {
  initialItems: NoticeItemType[];
  initialCursor: string | null;
  username: string | null;
}

/** Main notice list with infinite scroll and optional publish button */
export function NoticesContent({ initialItems, initialCursor, username }: NoticesContentProps) {
  const { items, isLoading, hasMore, sentinelRef, publishNotice, editNotice, removeNotice } =
    useNoticeList({ initialItems, initialCursor });
  const [showPublish, setShowPublish] = useState(false);
  const [editingNotice, setEditingNotice] = useState<NoticeItemType | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const isAdmin = username === "rainson";

  return (
    <>
      {/* Header row with optional publish button */}
      <div className="flex items-center justify-between">
        <span className="text-[13px] text-slate-400">
          共 {items.length} 条通知{hasMore ? "+" : ""}
        </span>
        {isAdmin && (
          <button
            type="button"
            onClick={() => setShowPublish(true)}
            className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-3.5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-teal-700"
          >
            <Plus className="h-4 w-4" />
            新通知
          </button>
        )}
      </div>

      {/* Notice list */}
      {items.length > 0 ? (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          {items.map((notice) => (
            <NoticeItem
              key={notice.id}
              notice={notice}
              isAdmin={isAdmin}
              onEdit={setEditingNotice}
              onDelete={setDeletingId}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center gap-3 py-16 text-slate-400">
          <Megaphone className="h-10 w-10" />
          <span className="text-sm">暂无通知</span>
        </div>
      )}

      {/* Loading spinner */}
      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
        </div>
      )}

      {/* Infinite scroll sentinel */}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      {/* Publish modal */}
      {isAdmin && (
        <PublishNoticeModal
          open={showPublish}
          onOpenChange={setShowPublish}
          onPublish={publishNotice}
        />
      )}

      {/* Edit modal */}
      {isAdmin && (
        <EditNoticeModal
          notice={editingNotice}
          open={editingNotice !== null}
          onOpenChange={(open) => {
            if (!open) setEditingNotice(null);
          }}
          onSave={editNotice}
        />
      )}

      {/* Delete confirmation */}
      {isAdmin && (
        <AlertDialog
          open={deletingId !== null}
          onOpenChange={(open) => {
            if (!open) setDeletingId(null);
          }}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>确定删除这条通知？</AlertDialogTitle>
              <AlertDialogDescription>
                删除后将不再对用户显示，此操作无法撤销。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>取消</AlertDialogCancel>
              <AlertDialogAction
                variant="destructive"
                onClick={async () => {
                  if (deletingId) {
                    await removeNotice(deletingId);
                    setDeletingId(null);
                  }
                }}
              >
                删除
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </>
  );
}
