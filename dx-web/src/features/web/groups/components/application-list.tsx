"use client";

import { Check, X } from "lucide-react";
import type { GroupApplication } from "../types/group";

interface ApplicationListProps {
  applications: GroupApplication[];
  onAccept: (appId: string) => void;
  onReject: (appId: string) => void;
}

export function ApplicationList({ applications, onAccept, onReject }: ApplicationListProps) {
  return (
    <div className="flex w-full flex-col overflow-hidden rounded-[14px] border border-border bg-card">
      <div className="border-b border-border px-5 py-3.5">
        <span className="text-sm font-semibold text-foreground">待审批（{applications.length}）</span>
      </div>
      <div className="flex-1 overflow-y-auto">
        {applications.length === 0 && (
          <div className="px-5 py-6 text-center text-xs text-muted-foreground">暂无申请</div>
        )}
        {applications.map((app, i) => (
          <div key={app.id}>
            {i > 0 && <div className="h-px bg-border" />}
            <div className="flex items-center gap-3 px-5 py-2.5">
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
    </div>
  );
}
