"use client";

import { useMemo } from "react";
import useSWR from "swr";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { MeHero } from "@/features/web/me/components/me-hero";
import { ProfileBlock } from "@/features/web/me/components/profile-block";
import { AccountBlock } from "@/features/web/me/components/account-block";
import { SecurityBlock } from "@/features/web/me/components/security-block";
import { MembershipBlock } from "@/features/web/me/components/membership-block";
import { StatsBlock } from "@/features/web/me/components/stats-block";
import { InviteBlock } from "@/features/web/me/components/invite-block";
import { toMeProfile } from "@/features/web/me/types/me.types";
import type { ApiProfileData } from "@/features/web/me/types/me.types";

export default function MePage() {
  const { data: raw, mutate } = useSWR<ApiProfileData>("/api/user/profile");
  const profile = useMemo(() => raw ? toMeProfile(raw) : null, [raw]);

  if (!profile) return null;

  const refreshProfile = () => { mutate(); };

  return (
    <div className="flex min-h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="个人中心"
        subtitle="管理你的个人资料和账号信息"
      />
      <MeHero profile={profile} onProfileChanged={refreshProfile} />
      <div className="flex flex-col gap-5">
        <ProfileBlock profile={profile} onProfileChanged={refreshProfile} />
        <AccountBlock profile={profile} onProfileChanged={refreshProfile} />
        <SecurityBlock />
        <MembershipBlock profile={profile} />
        <StatsBlock profile={profile} />
        <InviteBlock inviteCode={profile.inviteCode} />
      </div>
    </div>
  );
}
