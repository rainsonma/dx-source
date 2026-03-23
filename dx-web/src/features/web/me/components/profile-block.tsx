"use client";

import { useState } from "react";
import { Pencil } from "lucide-react";

import { EditProfileDialog } from "@/features/web/me/components/edit-profile-dialog";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Profile info block (nickname, city, introduction) with edit dialog */
export function ProfileBlock({ profile, onProfileChanged }: { profile: MeProfile; onProfileChanged?: () => void }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="rounded-2xl border border-border bg-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-foreground">个人资料</h3>
        <button
          onClick={() => setOpen(true)}
          className="flex items-center gap-1.5 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          <Pencil className="h-3.5 w-3.5" />
          编辑
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <InfoItem label="昵称" value={profile.nickname ?? "未设置"} />
        <InfoItem label="城市" value={profile.city ?? "未设置"} />
        <div className="md:col-span-2">
          <InfoItem label="简介" value={profile.introduction ?? "未设置"} />
        </div>
      </div>

      <EditProfileDialog
        open={open}
        onOpenChange={setOpen}
        nickname={profile.nickname}
        city={profile.city}
        introduction={profile.introduction}
        onProfileChanged={onProfileChanged}
      />
    </div>
  );
}

function InfoItem({ label, value }: { label: string; value: string }) {
  const isEmpty = value === "未设置";
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className={`text-sm ${isEmpty ? "text-muted-foreground" : "text-foreground"}`}>{value}</span>
    </div>
  );
}
