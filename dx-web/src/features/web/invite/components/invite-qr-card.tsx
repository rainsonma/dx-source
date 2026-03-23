"use client";

import { useInviteQrcode } from "@/features/web/invite/hooks/use-invite-qrcode";

type InviteQrCardProps = {
  url: string;
  title: string;
  subtitle: string;
};

export function InviteQrCard({ url, title, subtitle }: InviteQrCardProps) {
  const containerRef = useInviteQrcode(url);

  return (
    <div className="flex w-full flex-col gap-3.5 rounded-[14px] border border-border bg-card p-5 lg:w-[260px]">
      <div className="flex items-center gap-4">
        <div
          ref={containerRef}
          className="flex h-[100px] w-[100px] items-center justify-center rounded-[10px] border border-border bg-muted p-2"
        />
        <div className="flex flex-col gap-2">
          <span className="text-sm font-semibold text-foreground">{title}</span>
          <span className="text-xs text-muted-foreground">{subtitle}</span>
          <button
            type="button"
            className="rounded-lg bg-teal-600 px-3 py-1.5 text-xs font-medium text-white"
          >
            保存图片
          </button>
        </div>
      </div>
    </div>
  );
}
