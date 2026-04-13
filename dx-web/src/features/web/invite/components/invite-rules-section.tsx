"use client";

import Link from "next/link";
import { ScrollText, Lock, Crown } from "lucide-react";
import type { UserGrade } from "@/consts/user-grade";
import {
  inviteRules,
  commissionTiers,
  commissionRewardKeys,
  rewardRowLabels,
  inviteeDiscounts,
  formatRewardValue,
  type CommissionTier,
} from "@/features/web/invite/helpers/invite-rules";

type InviteRulesSectionProps = {
  userGrade: UserGrade | null;
};

export function InviteRulesSection({ userGrade }: InviteRulesSectionProps) {
  const isLoading = userGrade === null;
  const isFreeUser = userGrade === "free";

  return (
    <div className="flex flex-col gap-5 rounded-[14px] border border-border bg-card p-4 lg:gap-6 lg:p-6">
      <SectionHeader />
      <RulesList />
      {isLoading && <LoadingPlaceholder />}
      {!isLoading && isFreeUser && <LockedHint />}
      {!isLoading && !isFreeUser && (
        <>
          <CommissionTiersBlock />
          <InviteeDiscountsBlock />
        </>
      )}
    </div>
  );
}

function SectionHeader() {
  return (
    <div className="flex items-center gap-2">
      <ScrollText className="h-[18px] w-[18px] text-teal-600" />
      <span className="text-base font-semibold text-foreground">活动规则</span>
    </div>
  );
}

function SubHeader({ title }: { title: string }) {
  return (
    <div className="flex items-center gap-2">
      <span className="h-3 w-1 rounded-full bg-teal-600" />
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
    </div>
  );
}

function RulesList() {
  return (
    <div className="flex flex-col gap-3">
      {inviteRules.map((rule, i) => (
        <div key={i} className="flex gap-2.5">
          <span className="text-sm font-semibold text-teal-600">{i + 1}.</span>
          <span className="text-sm text-muted-foreground">{rule}</span>
        </div>
      ))}
    </div>
  );
}

function CommissionTiersBlock() {
  return (
    <div className="flex flex-col gap-3">
      <SubHeader title="佣金体系" />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {commissionTiers.map((tier) => (
          <CommissionTierCard key={tier.id} tier={tier} />
        ))}
      </div>
    </div>
  );
}

function CommissionTierCard({ tier }: { tier: CommissionTier }) {
  const isLifetime = tier.id === "lifetime";
  const cardClass = isLifetime
    ? "rounded-[10px] border border-teal-500/40 bg-gradient-to-b from-teal-50/40 to-transparent p-4"
    : "rounded-[10px] border border-border bg-card p-4";

  return (
    <div className={`flex flex-col ${cardClass}`}>
      <div className="flex items-center gap-1.5">
        <span className="text-sm font-semibold text-foreground">{tier.label}</span>
        {isLifetime && <Crown className="h-3.5 w-3.5 text-teal-600" />}
      </div>
      <span className="pb-2 pt-0.5 text-xs text-muted-foreground">{tier.sublabel}</span>
      <div className="flex flex-col">
        {commissionRewardKeys.map((key) => (
          <div
            key={key}
            className="flex items-center justify-between border-b border-border/40 py-2 last:border-0"
          >
            <span className="text-sm text-muted-foreground">{rewardRowLabels[key]}</span>
            <span className="text-sm font-semibold text-foreground">
              {formatRewardValue(tier.rewards[key])}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function InviteeDiscountsBlock() {
  return (
    <div className="flex flex-col gap-3">
      <SubHeader title="被邀请者专属折扣" />
      <div className="flex flex-col">
        {inviteeDiscounts.map((discount) => (
          <div
            key={discount.grade}
            className="flex items-center justify-between border-b border-border/40 py-2 last:border-0"
          >
            <span className="text-sm text-muted-foreground">{discount.label}</span>
            <span className="text-sm font-semibold text-foreground">
              {formatRewardValue(discount.value)}
            </span>
          </div>
        ))}
      </div>
      <p className="pt-1 text-xs text-muted-foreground">
        * 未获邀请直接购买者无折扣优惠
      </p>
    </div>
  );
}

function LockedHint() {
  return (
    <div className="flex min-h-[280px] flex-col items-center justify-center gap-3 rounded-[10px] border border-dashed border-border/60 bg-muted/30 p-6">
      <Lock className="h-6 w-6 text-muted-foreground" />
      <div className="flex flex-col items-center gap-1 text-center">
        <span className="text-sm font-medium text-foreground">
          佣金体系与会员折扣仅会员可见
        </span>
        <span className="text-xs text-muted-foreground">
          升级会员解锁完整邀请奖励
        </span>
      </div>
      <Link
        href="/purchase/membership"
        className="inline-flex items-center gap-1.5 rounded-[10px] bg-teal-600 px-4 py-2 text-xs font-semibold text-white outline-none transition-colors hover:bg-teal-700 focus-visible:ring-[3px] focus-visible:ring-ring/50"
      >
        升级会员
      </Link>
    </div>
  );
}

function LoadingPlaceholder() {
  return (
    <div className="min-h-[280px] animate-pulse rounded-[10px] border border-border/40 bg-muted/10" />
  );
}
