"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
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
  const [timeLeft, setTimeLeft] = useState(60);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const id = requestAnimationFrame(() => setMounted(true));
    return () => cancelAnimationFrame(id);
  }, []);

  useEffect(() => {
    if (timeLeft <= 0) {
      onDecline();
      return;
    }
    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, onDecline]);

  if (!mounted) return null;

  return createPortal(
    <div style={{ position: "fixed", bottom: 24, right: 24, zIndex: 99999 }}>
      <div
        className="flex w-80 flex-col gap-3 rounded-2xl border-[3px] border-teal-600/60 p-4 shadow-2xl"
        style={{ background: "radial-gradient(ellipse at center, #1E1B4B, #0F0A2E)" }}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Swords className="h-4 w-4 text-teal-400" />
            <span className="text-sm font-bold text-white">PK 邀请</span>
          </div>
          <button
            type="button"
            onClick={onDecline}
            className="flex h-6 w-6 items-center justify-center rounded-md hover:bg-white/10"
          >
            <X className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
        <div className="flex flex-col gap-1">
          <p className="text-sm text-white">
            <span className="font-semibold text-teal-300">{initiatorName}</span> 邀请你进行 PK 对战
          </p>
          <p className="text-xs text-slate-400">
            {gameName} · {levelName}
          </p>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs text-slate-500">{timeLeft}s</span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={onDecline}
              className="rounded-lg border border-slate-600 px-4 py-1.5 text-sm font-medium text-slate-300 hover:bg-white/10"
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
    </div>,
    document.body
  );
}
