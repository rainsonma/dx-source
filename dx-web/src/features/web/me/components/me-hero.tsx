"use client";

import { Zap, Flame, Coins } from "lucide-react";

import { getLevel } from "@/consts/user-level";
import { USER_GRADE_LABELS } from "@/consts/user-grade";
import { AvatarUploader } from "@/features/web/me/components/avatar-uploader";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Hero banner with avatar, display name, level, grade, and key stats */
export function MeHero({ profile, onProfileChanged }: { profile: MeProfile; onProfileChanged?: () => void }) {
  const displayName = profile.nickname ?? profile.username;
  const level = getLevel(profile.exp);
  const gradeLabel = USER_GRADE_LABELS[profile.grade];

  return (
    <div className="flex flex-col items-center gap-4 rounded-2xl border border-border bg-card p-8 md:flex-row md:items-center md:gap-6">
      <AvatarUploader
        userId={profile.id}
        avatarUrl={profile.avatarUrl}
        displayName={displayName}
        onProfileChanged={onProfileChanged}
      />

      <div className="flex flex-1 flex-col items-center gap-2 md:items-start">
        <div className="flex items-center gap-2">
          <h2 className="text-xl font-bold text-foreground">{displayName}</h2>
          <span className="rounded-full bg-indigo-100 px-2 py-0.5 text-[11px] font-bold text-indigo-600">
            Lv.{level}
          </span>
          <span className="rounded bg-border px-1.5 py-0.5 text-[11px] font-semibold text-muted-foreground">
            {gradeLabel}
          </span>
        </div>
        <p className="text-sm text-muted-foreground">@{profile.username}</p>

        <div className="mt-2 flex gap-6">
          <div className="flex items-center gap-1.5">
            <Zap className="h-4 w-4 text-teal-600" />
            <span className="text-sm font-semibold text-foreground">{profile.exp.toLocaleString()}</span>
            <span className="text-xs text-muted-foreground">经验</span>
          </div>
          <div className="flex items-center gap-1.5">
            <Flame className="h-4 w-4 text-orange-500" />
            <span className="text-sm font-semibold text-foreground">{profile.currentPlayStreak}</span>
            <span className="text-xs text-muted-foreground">天连续</span>
          </div>
          <div className="flex items-center gap-1.5">
            <Coins className="h-4 w-4 text-amber-500" />
            <span className="text-sm font-semibold text-foreground">{profile.beans.toLocaleString()}</span>
            <span className="text-xs text-muted-foreground">能量豆</span>
          </div>
        </div>
      </div>
    </div>
  );
}
