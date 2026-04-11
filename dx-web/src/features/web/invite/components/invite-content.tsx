"use client";

import { useState } from "react";
import {
  Copy,
  Check,
  Users,
  DollarSign,
  UserCheck,
  TrendingUp,
  ScrollText,
  Info,
  Share2,
} from "lucide-react";

import { InviteQrCard } from "@/features/web/invite/components/invite-qr-card";
import { InviteReferralTable } from "@/features/web/invite/components/invite-referral-table";
import { ShareSnippetsModal } from "@/features/web/invite/components/share-snippets-modal";
import type { InviteStats } from "@/features/web/invite/helpers/invite-stats.helper";
import type { ReferralItem } from "@/features/web/invite/actions/invite.action";

type InviteContentProps = {
  inviteUrl: string;
  referrals: ReferralItem[];
  totalPages: number;
  stats: InviteStats;
};

const rules = [
  "邀请好友通过您的专属链接注册斗学平台",
  "好友成功注册并完成首次购买会员即算邀请成功",
  "月度会员佣金 ¥9.90，季度会员 ¥29.70，年度会员 ¥89.70",
  "佣金每月 15 日统一结算，可提现至绑定的银行账户",
  "邀请人数不设上限，邀请越多佣金越多",
];

export function InviteContent({ inviteUrl, referrals, totalPages, stats }: InviteContentProps) {
  const [shareOpen, setShareOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const statCards = [
    { icon: DollarSign, iconBg: "bg-teal-100", iconColor: "text-teal-600", value: stats.totalReward, label: "累计获得推广佣金" },
    { icon: Users, iconBg: "bg-blue-100", iconColor: "text-blue-500", value: String(stats.totalFriends), label: `本月新增 ${stats.newThisMonth} 位好友` },
    { icon: UserCheck, iconBg: "bg-amber-100", iconColor: "text-amber-500", value: String(stats.pendingCount), label: "好友已注册待验证" },
    { icon: TrendingUp, iconBg: "bg-purple-100", iconColor: "text-purple-600", value: stats.conversionRate, label: "邀请成功转化比例" },
  ];

  /** Copy invite URL to clipboard and flash a check icon for 2 seconds */
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(inviteUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Silently fail if clipboard access is denied
    }
  };

  return (
    <>
      {/* Banner row */}
      <div className="flex w-full flex-col gap-4 lg:flex-row lg:flex-wrap">
        {/* Main banner */}
        <div className="flex flex-1 flex-col gap-3 rounded-[14px] bg-gradient-to-b from-teal-600 to-teal-700 p-5">
          <div className="flex items-center justify-between">
            <span className="text-base font-bold text-white">
              邀请好友，成功订阅即获佣金奖励！
            </span>
            <button
              type="button"
              onClick={() => setShareOpen(true)}
              className="flex items-center gap-1.5 rounded-lg bg-yellow-100 px-3 py-1.5 text-xs font-semibold text-yellow-800"
            >
              <Share2 className="h-3.5 w-3.5" />
              快速分享词
            </button>
          </div>
          <p className="text-xs leading-relaxed text-white/80">
            分享您的专属链接给好友，好友注册并订阅后，您即可获得相应佣金奖励，邀请无上限！
          </p>
          <div className="flex flex-col gap-3 md:flex-row">
            <div className="flex h-9 flex-1 items-center rounded-[10px] border border-white/25 bg-white/10 px-3.5">
              <span className="truncate text-[13px] text-white/60">
                {inviteUrl}
              </span>
            </div>
            <button
              type="button"
              onClick={handleCopy}
              className="flex items-center justify-center gap-1.5 rounded-[10px] bg-white px-4 py-2 text-[13px] font-semibold text-teal-700"
            >
              {copied ? (
                <Check className="h-3.5 w-3.5" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
              复制链接
            </button>
          </div>
        </div>

        {/* QR cards */}
        <InviteQrCard
          url={inviteUrl}
          title="邀请链接"
          subtitle="分享此二维码邀请好友"
        />
        <InviteQrCard
          url={inviteUrl}
          title="微信扫码"
          subtitle="扫码分享到微信好友"
        />
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {statCards.map((stat) => (
          <div
            key={stat.label}
            className="flex flex-col gap-2 rounded-[14px] border border-border bg-card p-4 lg:p-5"
          >
            <div className="flex items-center gap-2">
              <stat.icon className={`h-[18px] w-[18px] ${stat.iconColor}`} />
            </div>
            <span className="text-xl font-extrabold text-foreground lg:text-[28px]">
              {stat.value}
            </span>
            <span className="text-xs text-muted-foreground">{stat.label}</span>
          </div>
        ))}
      </div>

      {/* Invited clients table */}
      <div className="overflow-hidden rounded-[14px] border border-border bg-card">
        <div className="flex items-center justify-between px-4 py-4 lg:px-6">
          <div className="flex items-center gap-2">
            <Users className="h-[18px] w-[18px] text-teal-600" />
            <span className="text-base font-semibold text-foreground">
              邀请记录
            </span>
          </div>
          <div className="flex items-center gap-2">
            <button type="button" className="rounded-full bg-muted px-3 py-1 text-xs font-medium text-muted-foreground">
              全部
            </button>
            <button type="button" className="rounded-full px-3 py-1 text-xs font-medium text-muted-foreground">
              待激活
            </button>
            <button type="button" className="rounded-full px-3 py-1 text-xs font-medium text-muted-foreground">
              已激活
            </button>
          </div>
        </div>
        <InviteReferralTable
          initialReferrals={referrals}
          initialTotalPages={totalPages}
        />
      </div>

      {/* Rules card */}
      <div className="flex flex-col gap-4 rounded-[14px] border border-border bg-card p-4 lg:p-6">
        <div className="flex items-center gap-2">
          <ScrollText className="h-[18px] w-[18px] text-teal-600" />
          <span className="text-base font-semibold text-foreground">活动规则</span>
        </div>
        <div className="flex flex-col gap-3">
          {rules.map((rule, i) => (
            <div key={i} className="flex gap-2.5">
              <span className="text-sm font-semibold text-teal-600">{i + 1}.</span>
              <span className="text-sm text-muted-foreground">{rule}</span>
            </div>
          ))}
        </div>
        <div className="flex gap-2 rounded-[10px] border border-amber-500/20 bg-amber-50/10 p-3">
          <Info className="h-3.5 w-3.5 shrink-0 text-amber-500" />
          <span className="text-xs leading-relaxed text-amber-800">
            注意：佣金奖励仅限被邀请好友首次购买会员时发放，续费订单不参与佣金计算。
          </span>
        </div>
      </div>

      {/* Share snippets modal */}
      <ShareSnippetsModal
        open={shareOpen}
        onOpenChange={setShareOpen}
        inviteUrl={inviteUrl}
      />
    </>
  );
}
