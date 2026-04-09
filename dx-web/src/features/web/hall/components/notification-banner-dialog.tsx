"use client";

import { createElement } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { NoticeItem } from "@/features/web/notice/actions/notice.action";
import { resolveNoticeIcon } from "@/features/web/notice/helpers/notice-icon";
import { formatRelativeTime } from "@/features/web/notice/helpers/notice-time";

interface NotificationBannerDialogProps {
  notice: NoticeItem | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/** ShadCN dialog showing a single notice, opened from NotificationBanner */
export function NotificationBannerDialog({
  notice,
  open,
  onOpenChange,
}: NotificationBannerDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="gap-0 overflow-hidden p-0 sm:max-w-lg">
        <DialogHeader className="border-b border-border bg-slate-50/60 px-6 py-4">
          <DialogTitle className="text-base font-bold text-foreground">
            斗学消息通知
          </DialogTitle>
          <DialogDescription className="sr-only">
            查看消息通知内容
          </DialogDescription>
        </DialogHeader>

        <div className="flex max-h-[60vh] flex-col gap-3 overflow-y-auto px-6 py-5">
          {notice && (
            <>
              <div className="flex items-start gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[10px] bg-teal-50">
                  {createElement(resolveNoticeIcon(notice.icon), {
                    className: "h-[18px] w-[18px] text-teal-600",
                  })}
                </div>
                <h3 className="pt-1.5 text-[15px] font-semibold text-foreground">
                  {notice.title}
                </h3>
              </div>
              {notice.content && (
                <p className="text-sm leading-relaxed whitespace-pre-wrap text-muted-foreground">
                  {notice.content}
                </p>
              )}
            </>
          )}
        </div>

        <div className="flex items-center justify-end border-t border-border px-6 py-3">
          <span className="text-xs text-muted-foreground">
            {notice ? formatRelativeTime(notice.createdAt) : ""}
          </span>
        </div>
      </DialogContent>
    </Dialog>
  );
}
