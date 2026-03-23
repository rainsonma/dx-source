"use client";

import { useState, useTransition } from "react";
import useSWR from "swr";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { USER_GRADE_LABELS } from "@/consts/user-grade";
import type { UserGrade } from "@/consts/user-grade";

type Redeem = {
  id: string;
  code: string;
  grade: string;
  redeemedAt: Date | null;
};

type RedeemResponse = {
  items: Redeem[];
  total: number;
  page: number;
  pageSize: number;
};

const PAGE_SIZE = 15;

/** Table displaying the current user's redeem records with pagination */
export function RedeemHistoryTable() {
  const [currentPage, setCurrentPage] = useState(1);
  const [isPending, startTransition] = useTransition();

  const { data } = useSWR<RedeemResponse>(
    `/api/redeems?page=${currentPage}&pageSize=${PAGE_SIZE}`
  );

  const redeems = data?.items ?? [];
  const totalPages = data ? Math.ceil(data.total / data.pageSize) : 0;

  /** Load a specific page */
  const handlePageChange = (page: number) => {
    startTransition(() => {
      setCurrentPage(page);
    });
  };

  /** Format date for display */
  const formatDate = (date: Date | null) => {
    if (!date) return "-";
    return new Date(date).toLocaleDateString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  };

  return (
    <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="flex items-center justify-between px-4 py-5 md:px-6">
        <span className="text-base font-semibold text-foreground">兑换记录</span>
      </div>
      <div className="h-px w-full bg-border" />

      <div className={isPending ? "opacity-60 transition-opacity" : ""}>
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted hover:bg-muted">
                <TableHead className="pl-6 text-xs font-semibold text-muted-foreground">兑换码</TableHead>
                <TableHead className="w-[140px] text-xs font-semibold text-muted-foreground">兑换等级</TableHead>
                <TableHead className="w-[140px] text-xs font-semibold text-muted-foreground">兑换时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {redeems.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="py-10 text-center text-sm text-muted-foreground">
                    暂无兑换记录
                  </TableCell>
                </TableRow>
              )}
              {redeems.map((redeem) => (
                <TableRow key={redeem.id}>
                  <TableCell className="pl-6 font-mono text-[13px] font-medium text-foreground">
                    {redeem.code}
                  </TableCell>
                  <TableCell className="text-[13px] text-muted-foreground">
                    {USER_GRADE_LABELS[redeem.grade as UserGrade] ?? redeem.grade}
                  </TableCell>
                  <TableCell className="text-[13px] text-muted-foreground">
                    {formatDate(redeem.redeemedAt)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>

        <DataTablePagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={handlePageChange}
        />
      </div>
    </div>
  );
}
