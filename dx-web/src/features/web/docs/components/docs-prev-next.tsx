import Link from "next/link";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { TopicRef } from "@/features/web/docs/types";

type Props = {
  prev: TopicRef | null;
  next: TopicRef | null;
};

export function DocsPrevNext({ prev, next }: Props) {
  return (
    <div className="flex items-stretch justify-between gap-4">
      {prev ? (
        <Link
          href={`/docs/${prev.category.slug}/${prev.topic.slug}`}
          className="flex flex-col items-start gap-1 rounded-lg border border-slate-200 px-4 py-3 hover:border-slate-300"
        >
          <span className="text-xs text-slate-400">上一页</span>
          <div className="flex items-center gap-1.5">
            <ChevronLeft
              className="h-3.5 w-3.5 text-slate-500"
              aria-hidden="true"
            />
            <span className="text-sm font-medium text-slate-700">
              {prev.topic.title}
            </span>
          </div>
        </Link>
      ) : (
        <div />
      )}

      {next ? (
        <Link
          href={`/docs/${next.category.slug}/${next.topic.slug}`}
          className="flex flex-col items-end gap-1 rounded-lg border border-slate-200 px-4 py-3 hover:border-slate-300"
        >
          <span className="text-xs text-slate-400">下一页</span>
          <div className="flex items-center gap-1.5">
            <span className="text-sm font-medium text-slate-700">
              {next.topic.title}
            </span>
            <ChevronRight
              className="h-3.5 w-3.5 text-slate-500"
              aria-hidden="true"
            />
          </div>
        </Link>
      ) : (
        <div />
      )}
    </div>
  );
}
