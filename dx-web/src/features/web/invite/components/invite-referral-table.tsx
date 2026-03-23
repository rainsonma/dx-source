"use client";

import { useState, useTransition } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  REFERRAL_STATUS_LABELS,
} from "@/consts/referral-status";
import type { ReferralStatus } from "@/consts/referral-status";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { fetchReferralPage } from "@/features/web/invite/actions/invite.action";
import {
  AVATAR_COLORS,
  getDisplayName,
  maskEmail,
  formatDate,
  formatReward,
  getStatusClasses,
} from "@/features/web/invite/helpers/referral-table.helper";

type Referral = {
  id: string;
  status: string;
  rewardAmount: number;
  rewardedAt: Date | null;
  createdAt: Date;
  invitee: {
    id: string;
    username: string;
    nickname: string | null;
    email: string | null;
    grade: string;
  } | null;
};

type InviteReferralTableProps = {
  initialReferrals: Referral[];
  initialTotalPages: number;
};

/** Table displaying paginated invite referral records */
export function InviteReferralTable({
  initialReferrals,
  initialTotalPages,
}: InviteReferralTableProps) {
  const [referrals, setReferrals] = useState(initialReferrals);
  const [totalPages, setTotalPages] = useState(initialTotalPages);
  const [currentPage, setCurrentPage] = useState(1);
  const [isPending, startTransition] = useTransition();

  /** Load a specific page of referrals */
  const handlePageChange = (page: number) => {
    startTransition(async () => {
      const result = await fetchReferralPage(page);
      if ("data" in result && result.data) {
        setReferrals(result.data.referrals);
        setTotalPages(result.data.totalPages);
        setCurrentPage(page);
      }
    });
  };

  return (
    <div className={isPending ? "opacity-60 transition-opacity" : ""}>
      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted hover:bg-muted">
              <TableHead className="pl-6 text-xs font-semibold text-muted-foreground">
                好友信息
              </TableHead>
              <TableHead className="w-[120px] text-xs font-semibold text-muted-foreground">
                注册日期
              </TableHead>
              <TableHead className="w-[100px] text-xs font-semibold text-muted-foreground">
                状态
              </TableHead>
              <TableHead className="w-[120px] text-xs font-semibold text-muted-foreground">
                佣金
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {referrals.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="py-10 text-center text-sm text-muted-foreground">
                  暂无邀请记录
                </TableCell>
              </TableRow>
            )}
            {referrals.map((referral, index) => {
              const colors = AVATAR_COLORS[index % AVATAR_COLORS.length];
              const displayName = getDisplayName(referral.invitee);
              const initial = displayName.charAt(0);

              return (
                <TableRow key={referral.id}>
                  <TableCell className="pl-6">
                    <div className="flex items-center gap-2.5">
                      <div
                        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full ${colors.bg}`}
                      >
                        <span className={`text-[13px] font-semibold ${colors.text}`}>
                          {initial}
                        </span>
                      </div>
                      <div className="flex flex-col gap-0.5">
                        <span className="text-[13px] font-semibold text-foreground">
                          {displayName}
                        </span>
                        <span className="text-[11px] text-muted-foreground">
                          {maskEmail(referral.invitee?.email ?? null)}
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="text-[13px] text-muted-foreground">
                    {formatDate(referral.createdAt)}
                  </TableCell>
                  <TableCell>
                    <span
                      className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${getStatusClasses(referral.status)}`}
                    >
                      {REFERRAL_STATUS_LABELS[referral.status as ReferralStatus] ?? referral.status}
                    </span>
                  </TableCell>
                  <TableCell className="text-[13px] font-semibold text-teal-600">
                    {formatReward(referral.rewardAmount, referral.status)}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>

      <DataTablePagination
        currentPage={currentPage}
        totalPages={totalPages}
        onPageChange={handlePageChange}
      />
    </div>
  );
}
