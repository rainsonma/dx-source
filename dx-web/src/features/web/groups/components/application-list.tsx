"use client";

import { Check, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { GroupApplication } from "../types/group";

interface ApplicationListDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  applications: GroupApplication[];
  onAccept: (appId: string) => void;
  onReject: (appId: string) => void;
}

export function ApplicationListDialog({
  open,
  onOpenChange,
  applications,
  onAccept,
  onReject,
}: ApplicationListDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>加入待审批（{applications.length}）</DialogTitle>
          <DialogDescription>审批用户的加入申请</DialogDescription>
        </DialogHeader>
        <div className="max-h-80 overflow-y-auto">
          {applications.length === 0 && (
            <div className="py-6 text-center text-xs text-muted-foreground">暂无申请</div>
          )}
          {applications.map((app, i) => (
            <div key={app.id}>
              {i > 0 && <div className="h-px bg-border" />}
              <div className="flex items-center gap-3 py-2.5">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-amber-100">
                  <span className="text-xs font-semibold text-amber-700">{app.user_name[0]}</span>
                </div>
                <div className="flex flex-1 flex-col gap-0.5">
                  <span className="text-[13px] font-semibold text-foreground">{app.user_name}</span>
                  <span className="text-[11px] text-muted-foreground">申请加入</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <button
                    type="button"
                    onClick={() => onAccept(app.id)}
                    className="flex h-7 w-7 items-center justify-center rounded-lg bg-teal-600 text-white hover:bg-teal-700"
                  >
                    <Check className="h-3.5 w-3.5" />
                  </button>
                  <button
                    type="button"
                    onClick={() => onReject(app.id)}
                    className="flex h-7 w-7 items-center justify-center rounded-lg border border-border text-muted-foreground hover:bg-red-50 hover:text-red-500"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
