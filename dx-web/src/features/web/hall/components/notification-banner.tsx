"use client";

import { createElement, useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import type { NoticeItem } from "@/features/web/notice/actions/notice.action";
import { resolveNoticeIcon } from "@/features/web/notice/helpers/notice-icon";
import { formatRelativeTime } from "@/features/web/notice/helpers/notice-time";
import { NotificationBannerDialog } from "./notification-banner-dialog";

const ROTATION_INTERVAL_MS = 5000;

type NoticeListResponse = {
  items: NoticeItem[];
  nextCursor: string;
  hasMore: boolean;
};

/** Rotating notification ticker above the /hall 3-card row */
export function NotificationBanner() {
  const [notices, setNotices] = useState<NoticeItem[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [hovered, setHovered] = useState(false);
  const [dialogNotice, setDialogNotice] = useState<NoticeItem | null>(null);

  const paused = hovered || dialogNotice !== null;

  // Fetch latest 3 notices once on mount. Silent failure: non-zero code or
  // thrown error leaves `notices` empty, which causes the render to return null.
  useEffect(() => {
    let cancelled = false;
    apiClient
      .get<NoticeListResponse>("/api/notices?limit=3")
      .then((res) => {
        if (cancelled) return;
        if (res.code === 0) setNotices(res.data.items ?? []);
        setLoaded(true);
      })
      .catch(() => {
        if (!cancelled) setLoaded(true);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // Auto-rotate every ROTATION_INTERVAL_MS while unpaused and more than 1 notice.
  useEffect(() => {
    if (notices.length < 2 || paused) return;
    const id = setInterval(() => {
      setCurrentIndex((i) => (i + 1) % notices.length);
    }, ROTATION_INTERVAL_MS);
    return () => clearInterval(id);
  }, [notices.length, paused]);

  if (!loaded || notices.length === 0) return null;

  const current = notices[currentIndex];
  const Icon = resolveNoticeIcon(current.icon);

  return (
    <>
      <button
        type="button"
        onClick={() => setDialogNotice(current)}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        onFocus={() => setHovered(true)}
        onBlur={() => setHovered(false)}
        className="group flex w-full items-center gap-3 rounded-[14px] border border-border bg-card px-4 py-2.5 text-left transition-colors hover:border-teal-200 hover:bg-teal-50/30 focus-visible:ring-2 focus-visible:ring-teal-500/50 focus-visible:ring-offset-2 focus-visible:outline-hidden sm:px-5"
        aria-label={`查看消息通知：${current.title}`}
      >
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] bg-teal-50 transition-colors group-hover:bg-teal-100">
          {createElement(Icon, {
            className: "h-[18px] w-[18px] text-teal-600",
          })}
        </div>

        <div
          key={currentIndex}
          className="animate-in fade-in-0 slide-in-from-bottom-1 flex min-w-0 flex-1 items-center gap-2.5 duration-300 sm:gap-3"
        >
          <span className="shrink-0 text-sm font-semibold text-foreground">
            {current.title}
          </span>
          {current.content && (
            <span className="hidden truncate text-sm text-muted-foreground sm:inline">
              {current.content}
            </span>
          )}
        </div>

        {notices.length > 1 && (
          <div className="hidden shrink-0 items-center gap-1 lg:flex">
            {notices.map((n, i) => (
              <span
                key={n.id}
                aria-hidden="true"
                className={cn(
                  "h-1.5 rounded-full transition-all",
                  i === currentIndex ? "w-3.5 bg-teal-500" : "w-1.5 bg-slate-300",
                )}
              />
            ))}
          </div>
        )}

        <span className="shrink-0 text-xs text-muted-foreground">
          {formatRelativeTime(current.createdAt)}
        </span>
      </button>

      <NotificationBannerDialog
        notice={dialogNotice}
        open={dialogNotice !== null}
        onOpenChange={(open) => {
          if (!open) setDialogNotice(null);
        }}
      />
    </>
  );
}
