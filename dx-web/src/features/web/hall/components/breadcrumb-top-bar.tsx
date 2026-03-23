import { ArrowLeft, ChevronRight } from "lucide-react";
import Link from "next/link";
import { TopActions } from "@/features/web/hall/components/top-actions";

type BreadcrumbItem = {
  label: string;
  href?: string;
  maxChars: number;
};

/** Truncate text to maxChars, appending ellipsis if exceeded */
function truncate(text: string, maxChars: number) {
  return text.length > maxChars ? text.slice(0, maxChars) + "…" : text;
}

export function BreadcrumbTopBar({
  backHref,
  items,
}: {
  backHref: string;
  items: BreadcrumbItem[];
}) {
  return (
    <div className="flex w-full items-center justify-between">
      <div className="flex items-center gap-4">
        <Link
          href={backHref}
          aria-label="返回"
          className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-border bg-card text-muted-foreground hover:bg-accent"
        >
          <ArrowLeft className="h-[18px] w-[18px]" />
        </Link>
        <div className="flex flex-wrap items-center gap-2">
          {items.map((item, i) => {
            const isLast = i === items.length - 1;
            const text = truncate(item.label, item.maxChars);

            return (
              <span key={i} className="flex items-center gap-2">
                {item.href && !isLast ? (
                  <Link
                    href={item.href}
                    title={item.label}
                    className="text-sm font-medium text-muted-foreground hover:text-foreground"
                  >
                    {text}
                  </Link>
                ) : (
                  <span
                    title={item.label}
                    className={
                      isLast
                        ? "text-sm font-semibold text-foreground"
                        : "text-sm font-medium text-muted-foreground"
                    }
                  >
                    {text}
                  </span>
                )}
                {!isLast && (
                  <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
                )}
              </span>
            );
          })}
        </div>
      </div>
      <TopActions />
    </div>
  );
}
