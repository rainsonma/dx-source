"use client";

import { useEffect, useState } from "react";
import { Swords, X } from "lucide-react";

interface PkInvitationPopupProps {
  pkId: string;
  gameName: string;
  levelName: string;
  initiatorName: string;
  onAccept: () => void;
  onDecline: () => void;
}

export function PkInvitationPopup({
  gameName,
  levelName,
  initiatorName,
  onAccept,
  onDecline,
}: PkInvitationPopupProps) {
  const [timeLeft, setTimeLeft] = useState(30);

  useEffect(() => {
    if (timeLeft <= 0) {
      onDecline();
      return;
    }
    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, onDecline]);

  return (
    <div className="fixed bottom-6 right-6 z-[9999]">
      <div className="flex w-80 flex-col gap-3 rounded-2xl border border-border bg-card p-4 shadow-lg">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Swords className="h-4 w-4 text-teal-600" />
            <span className="text-sm font-bold text-foreground">PK 邀请</span>
          </div>
          <button
            type="button"
            onClick={onDecline}
            className="flex h-6 w-6 items-center justify-center rounded-md hover:bg-muted"
          >
            <X className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>
        <div className="flex flex-col gap-1">
          <p className="text-sm text-foreground">
            <span className="font-semibold">{initiatorName}</span> 邀请你进行 PK 对战
          </p>
          <p className="text-xs text-muted-foreground">
            {gameName} · {levelName}
          </p>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs text-muted-foreground">{timeLeft}s</span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={onDecline}
              className="rounded-lg border border-border px-4 py-1.5 text-sm font-medium text-muted-foreground hover:bg-muted"
            >
              拒绝
            </button>
            <button
              type="button"
              onClick={onAccept}
              className="rounded-lg bg-teal-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-teal-700"
            >
              接受
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
