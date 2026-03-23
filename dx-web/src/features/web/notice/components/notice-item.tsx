import { Pencil, Trash2 } from "lucide-react";
import type { NoticeItem as NoticeItemType } from "@/features/web/notice/actions/notice.action";
import { resolveNoticeIcon } from "@/features/web/notice/helpers/notice-icon";
import { formatRelativeTime } from "@/features/web/notice/helpers/notice-time";

interface NoticeItemProps {
  notice: NoticeItemType;
  isAdmin?: boolean;
  onEdit?: (notice: NoticeItemType) => void;
  onDelete?: (id: string) => void;
}

/** Renders a single notice row with dynamic icon */
export function NoticeItem({ notice, isAdmin, onEdit, onDelete }: NoticeItemProps) {
  const Icon = resolveNoticeIcon(notice.icon);

  return (
    <div className="flex gap-3.5 border-b border-slate-200 px-4 py-4 last:border-b-0 lg:px-5">
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[10px] bg-teal-50">
        <Icon className="h-[18px] w-[18px] text-teal-600" />
      </div>
      <div className="flex flex-1 flex-col gap-1.5">
        <span className="text-sm font-semibold text-slate-900">
          {notice.title}
        </span>
        {notice.content && (
          <span className="text-[13px] leading-[1.5] text-slate-500">
            {notice.content}
          </span>
        )}
        <span className="text-xs text-slate-400">
          {formatRelativeTime(notice.createdAt)}
        </span>
      </div>
      {isAdmin && (
        <div className="flex shrink-0 items-start gap-1 pt-0.5">
          <button
            type="button"
            onClick={() => onEdit?.(notice)}
            className="flex h-7 w-7 items-center justify-center rounded-md text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
          >
            <Pencil className="h-3.5 w-3.5" />
          </button>
          <button
            type="button"
            onClick={() => onDelete?.(notice.id)}
            className="flex h-7 w-7 items-center justify-center rounded-md text-slate-400 transition-colors hover:bg-red-50 hover:text-red-500"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      )}
    </div>
  );
}
