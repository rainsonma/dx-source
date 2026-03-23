"use client";

import useSWR from "swr";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { RedeemContent } from "@/features/web/redeem/components/redeem-content";

export default function RedeemPage() {
  const { data: profile } = useSWR<{ username: string }>("/api/user/profile");

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="兑换码"
        subtitle="输入兑换码激活会员权益"
      />
      <RedeemContent username={profile?.username ?? null} />
    </div>
  );
}
