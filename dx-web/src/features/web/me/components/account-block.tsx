"use client";

import { useState } from "react";
import { Pencil } from "lucide-react";

import { EditEmailDialog } from "@/features/web/me/components/edit-email-dialog";
import type { MeProfile } from "@/features/web/me/types/me.types";

/** Account info block (username, email, phone) with email edit dialog */
export function AccountBlock({ profile, onProfileChanged }: { profile: MeProfile; onProfileChanged?: () => void }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="rounded-2xl border border-border bg-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-foreground">账号信息</h3>
        <button
          onClick={() => setOpen(true)}
          className="flex items-center gap-1.5 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          <Pencil className="h-3.5 w-3.5" />
          编辑
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <InfoItem label="用户名" value={profile.username} />
        <InfoItem label="邮箱" value={profile.email ?? "未设置"} />
        <InfoItem label="手机号" value={profile.phone ?? "暂未开放"} muted />
      </div>

      <EditEmailDialog
        open={open}
        onOpenChange={setOpen}
        currentEmail={profile.email}
        onProfileChanged={onProfileChanged}
      />
    </div>
  );
}

function InfoItem({ label, value, muted }: { label: string; value: string; muted?: boolean }) {
  const isEmpty = value === "未设置" || value === "暂未开放";
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className={`text-sm ${isEmpty || muted ? "text-muted-foreground" : "text-foreground"}`}>{value}</span>
    </div>
  );
}
