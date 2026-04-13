"use client";

import { useEffect, useState } from "react";
import useSWR from "swr";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { InviteContent } from "@/features/web/invite/components/invite-content";
import type { InviteStats } from "@/features/web/invite/helpers/invite-stats.helper";
import type { ReferralItem } from "@/features/web/invite/actions/invite.action";
import type { ApiProfileData } from "@/features/web/me/types/me.types";
import type { UserGrade } from "@/consts/user-grade";

type ApiInviteData = {
  inviteCode: string;
  stats: { total: number; pending: number; paid: number; rewarded: number };
  referrals: ReferralItem[];
  totalReferrals: number;
};

const emptyStats: InviteStats = {
  totalReward: "¥ 0.00",
  totalFriends: 0,
  newThisMonth: 0,
  pendingCount: 0,
  conversionRate: "0%",
};

export default function InvitePage() {
  const [inviteUrl, setInviteUrl] = useState("");
  const [referrals, setReferrals] = useState<ReferralItem[]>([]);
  const [totalPages, setTotalPages] = useState(0);
  const [stats, setStats] = useState<InviteStats>(emptyStats);

  const { data: profileData, error: profileError } = useSWR<ApiProfileData>(
    "/api/user/profile"
  );

  const userGrade: UserGrade | null = profileError
    ? "free"
    : profileData
      ? (profileData.grade as UserGrade)
      : null;

  useEffect(() => {
    async function load() {
      const res = await apiClient.get<ApiInviteData>("/api/invite");

      if (res.code !== 0) return;

      const data = res.data;
      const url = `${window.location.protocol}//${window.location.host}/invite/${data.inviteCode}`;
      const pages = Math.ceil(data.totalReferrals / 15);
      const converted = data.stats.paid + data.stats.rewarded;
      const rate =
        data.stats.total > 0
          ? Math.round((converted / data.stats.total) * 100)
          : 0;

      setInviteUrl(url);
      setReferrals(data.referrals);
      setTotalPages(pages);
      setStats({
        totalReward: "¥ 0.00",
        totalFriends: data.stats.total,
        newThisMonth: 0,
        pendingCount: data.stats.pending,
        conversionRate: `${rate}%`,
      });
    }

    load();
  }, []);

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="邀请推广"
        subtitle="邀请好友加入斗学，成功即可获得佣金奖励"
      />
      <InviteContent
        inviteUrl={inviteUrl}
        referrals={referrals}
        totalPages={totalPages}
        stats={stats}
        userGrade={userGrade}
      />
    </div>
  );
}
