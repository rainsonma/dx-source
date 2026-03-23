"use client";

import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationEllipsis,
} from "@/components/ui/pagination";
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from "lucide-react";

type DataTablePaginationProps = {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
};

/** Generate the visible page numbers with ellipsis gaps */
function getPageNumbers(current: number, total: number): (number | "ellipsis")[] {
  if (total <= 5) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }

  const pages: (number | "ellipsis")[] = [1];

  if (current > 3) {
    pages.push("ellipsis");
  }

  const start = Math.max(2, current - 1);
  const end = Math.min(total - 1, current + 1);

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  if (current < total - 2) {
    pages.push("ellipsis");
  }

  pages.push(total);

  return pages;
}

/** Reusable pagination controls for data tables */
export function DataTablePagination({
  currentPage,
  totalPages,
  onPageChange,
}: DataTablePaginationProps) {
  if (totalPages <= 1) return null;

  const isFirst = currentPage === 1;
  const isLast = currentPage === totalPages;
  const pages = getPageNumbers(currentPage, totalPages);

  return (
    <div className="flex items-center justify-end gap-4 px-4 py-3">
      <span className="text-sm text-muted-foreground">
        第 {currentPage} / {totalPages} 页
      </span>
      <Pagination className="mx-0 w-auto">
        <PaginationContent>
          {/* First page */}
          <PaginationItem>
            <button
              type="button"
              aria-label="第一页"
              disabled={isFirst}
              onClick={() => onPageChange(1)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronsLeft className="h-4 w-4" />
            </button>
          </PaginationItem>

          {/* Previous page */}
          <PaginationItem>
            <button
              type="button"
              aria-label="上一页"
              disabled={isFirst}
              onClick={() => onPageChange(currentPage - 1)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
          </PaginationItem>

          {/* Page numbers */}
          {pages.map((page, i) =>
            page === "ellipsis" ? (
              <PaginationItem key={`ellipsis-${i}`}>
                <PaginationEllipsis />
              </PaginationItem>
            ) : (
              <PaginationItem key={page}>
                <PaginationLink
                  href="#"
                  isActive={page === currentPage}
                  onClick={(e) => {
                    e.preventDefault();
                    onPageChange(page);
                  }}
                  className="h-8 w-8 text-sm"
                >
                  {page}
                </PaginationLink>
              </PaginationItem>
            ),
          )}

          {/* Next page */}
          <PaginationItem>
            <button
              type="button"
              aria-label="下一页"
              disabled={isLast}
              onClick={() => onPageChange(currentPage + 1)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </PaginationItem>

          {/* Last page */}
          <PaginationItem>
            <button
              type="button"
              aria-label="最后一页"
              disabled={isLast}
              onClick={() => onPageChange(totalPages)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronsRight className="h-4 w-4" />
            </button>
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  );
}
