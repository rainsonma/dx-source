import Link from "next/link";
import { Crown, ChevronRight } from "lucide-react";

import { USER_GRADE_LABELS } from "@/consts/user-grade";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Membership info block (grade, vipDueAt, beans) with upgrade link */
export function MembershipBlock({ profile }: { profile: MeProfile }) {
  const gradeLabel = USER_GRADE_LABELS[profile.grade];
  const dueDate = profile.vipDueAt
    ? new Date(profile.vipDueAt).toLocaleDateString("zh-CN")
    : "—";

  return (
    <div className="rounded-2xl border border-border bg-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-foreground">会员信息</h3>
        <Link
          href="/purchase/membership"
          className="flex items-center gap-1 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          升级会员
          <ChevronRight className="h-4 w-4" />
        </Link>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">会员等级</span>
          <div className="flex items-center gap-1.5">
            <Crown className="h-4 w-4 text-amber-500" />
            <span className="text-sm font-medium text-foreground">{gradeLabel}</span>
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">到期时间</span>
          <span className="text-sm text-foreground">{dueDate}</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">能量豆</span>
          <span className="text-sm font-medium text-foreground">{profile.beans.toLocaleString()}</span>
        </div>
      </div>
    </div>
  );
}
