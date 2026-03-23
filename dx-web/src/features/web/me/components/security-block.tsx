"use client";

import { useState } from "react";
import { Pencil, Lock } from "lucide-react";

import { ChangePasswordDialog } from "@/features/web/me/components/change-password-dialog";

/** Security settings block (password) with change password dialog */
export function SecurityBlock() {
  const [open, setOpen] = useState(false);

  return (
    <div className="rounded-2xl border border-border bg-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-foreground">安全设置</h3>
        <button
          onClick={() => setOpen(true)}
          className="flex items-center gap-1.5 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          <Pencil className="h-3.5 w-3.5" />
          编辑
        </button>
      </div>

      <div className="flex items-center gap-2">
        <Lock className="h-4 w-4 text-muted-foreground" />
        <span className="text-xs text-muted-foreground">密码</span>
        <span className="text-sm text-foreground">••••••••</span>
      </div>

      <ChangePasswordDialog open={open} onOpenChange={setOpen} />
    </div>
  );
}
