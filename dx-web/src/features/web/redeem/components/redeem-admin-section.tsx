"use client";

import { useState, useTransition } from "react";
import useSWR from "swr";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { TicketPlus } from "lucide-react";
import { toast } from "sonner";
import { USER_GRADE_LABELS } from "@/consts/user-grade";
import type { UserGrade } from "@/consts/user-grade";
import { generateCodesAction } from "@/features/web/redeem/actions/redeem.action";
import { GenerateCodesModal } from "@/features/web/redeem/components/generate-codes-modal";

type AdminRedeem = {
  id: string;
  code: string;
  grade: string;
  userId: string | null;
  redeemedAt: Date | null;
  createdAt: Date;
  user: { username: string; nickname: string | null } | null;
};

type AdminRedeemResponse = {
  items: AdminRedeem[];
  total: number;
  page: number;
  pageSize: number;
};

const PAGE_SIZE = 15;

/** Admin-only section: generate button + all codes data table */
export function RedeemAdminSection() {
  const [currentPage, setCurrentPage] = useState(1);
  const [isPending, startTransition] = useTransition();
  const [modalOpen, setModalOpen] = useState(false);

  const { data, mutate } = useSWR<AdminRedeemResponse>(
    `/api/admin/redeems?page=${currentPage}&pageSize=${PAGE_SIZE}`
  );

  const redeems = data?.items ?? [];
  const totalPages = data ? Math.ceil(data.total / data.pageSize) : 0;

  /** Generate codes and refresh the table */
  const handleGenerate = async (input: { grade: string; quantity: string }): Promise<boolean> => {
    const result = await generateCodesAction(input);

    if ("error" in result) {
      toast.error(result.error);
      return false;
    }

    toast.success(`成功生成 ${input.quantity} 个兑换码`);
    setCurrentPage(1);
    mutate();
    return true;
  };

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

  /** Get display name for a redeemed user */
  const getUserDisplay = (redeem: AdminRedeem) => {
    if (!redeem.user) return "-";
    return redeem.user.nickname ?? redeem.user.username;
  };

  return (
    <>
      <button
        type="button"
        onClick={() => setModalOpen(true)}
        className="flex items-center gap-1.5 self-start rounded-lg bg-teal-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-teal-700"
      >
        <TicketPlus className="h-4 w-4" />
        生成兑换码
      </button>

      <GenerateCodesModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        onGenerate={handleGenerate}
      />

      <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <div className="flex items-center justify-between px-4 py-5 md:px-6">
          <span className="text-base font-semibold text-foreground">兑换码管理</span>
        </div>
        <div className="h-px w-full bg-border" />

        <div className={isPending ? "opacity-60 transition-opacity" : ""}>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted hover:bg-muted">
                  <TableHead className="pl-6 text-xs font-semibold text-muted-foreground">兑换码</TableHead>
                  <TableHead className="w-[100px] text-xs font-semibold text-muted-foreground">等级</TableHead>
                  <TableHead className="w-[80px] text-xs font-semibold text-muted-foreground">状态</TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-muted-foreground">兑换用户</TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-muted-foreground">兑换时间</TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-muted-foreground">创建时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {redeems.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="py-10 text-center text-sm text-muted-foreground">
                      暂无兑换码
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
                    <TableCell>
                      <span
                        className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${
                          redeem.userId
                            ? "bg-teal-600/10 text-teal-600"
                            : "bg-amber-500/10 text-amber-600"
                        }`}
                      >
                        {redeem.userId ? "已兑换" : "未使用"}
                      </span>
                    </TableCell>
                    <TableCell className="text-[13px] text-muted-foreground">
                      {getUserDisplay(redeem)}
                    </TableCell>
                    <TableCell className="text-[13px] text-muted-foreground">
                      {formatDate(redeem.redeemedAt)}
                    </TableCell>
                    <TableCell className="text-[13px] text-muted-foreground">
                      {formatDate(redeem.createdAt)}
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
    </>
  );
}
