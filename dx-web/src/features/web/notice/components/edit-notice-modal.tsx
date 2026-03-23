"use client";

import { useState, useEffect } from "react";
import { Pencil, Loader2 } from "lucide-react";
import type { NoticeItem } from "@/features/web/notice/actions/notice.action";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";

interface EditNoticeModalProps {
  notice: NoticeItem | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (input: {
    id: string;
    title: string;
    content?: string;
    icon?: string;
  }) => Promise<boolean>;
}

/** Modal form for editing an existing notice */
export function EditNoticeModal({
  notice,
  open,
  onOpenChange,
  onSave,
}: EditNoticeModalProps) {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [icon, setIcon] = useState("");
  const [submitting, setSubmitting] = useState(false);

  /** Sync form fields when notice changes */
  useEffect(() => {
    if (notice) {
      setTitle(notice.title);
      setContent(notice.content ?? "");
      setIcon(notice.icon ?? "");
    }
  }, [notice]);

  async function handleSubmit() {
    if (!notice || !title.trim()) return;
    setSubmitting(true);
    const ok = await onSave({
      id: notice.id,
      title: title.trim(),
      content: content.trim() || undefined,
      icon: icon.trim() || undefined,
    });
    setSubmitting(false);
    if (ok) onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton
        className="max-w-[520px] gap-0 rounded-[20px] border-none p-0"
      >
        <div className="flex flex-col gap-5 px-7 pt-7 pb-6">
          {/* Header */}
          <DialogTitle className="flex items-center gap-2.5 text-xl font-bold text-slate-900">
            <Pencil className="h-[18px] w-[18px] text-teal-600" />
            编辑通知
          </DialogTitle>
          <DialogDescription className="sr-only">
            编辑通知标题和内容
          </DialogDescription>

          <div className="h-px bg-slate-100" />

          {/* Form */}
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <label
                htmlFor="edit-notice-title"
                className="text-[13px] font-medium text-slate-700"
              >
                标题 *
              </label>
              <input
                id="edit-notice-title"
                placeholder="通知标题"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                disabled={submitting}
                maxLength={200}
                className="h-10 rounded-lg border border-slate-200 bg-white px-3.5 text-[13px] text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
            </div>

            <div className="flex flex-col gap-2">
              <label
                htmlFor="edit-notice-content"
                className="text-[13px] font-medium text-slate-700"
              >
                内容
              </label>
              <textarea
                id="edit-notice-content"
                placeholder="通知内容（可选）"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                disabled={submitting}
                maxLength={2000}
                rows={4}
                className="rounded-lg border border-slate-200 bg-white px-3.5 py-2.5 text-[13px] leading-relaxed text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
            </div>

            <div className="flex flex-col gap-2">
              <label
                htmlFor="edit-notice-icon"
                className="text-[13px] font-medium text-slate-700"
              >
                图标
              </label>
              <input
                id="edit-notice-icon"
                placeholder="message-circle-more"
                value={icon}
                onChange={(e) => setIcon(e.target.value)}
                disabled={submitting}
                maxLength={50}
                className="h-10 rounded-lg border border-slate-200 bg-white px-3.5 text-[13px] text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
              <span className="text-xs text-slate-400">
                Lucide 图标名称，留空使用默认图标
              </span>
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-2.5">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
              className="rounded-lg border border-slate-200 px-4 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={submitting || !title.trim()}
              className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-teal-700 disabled:opacity-50"
            >
              {submitting && (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              )}
              保存
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
