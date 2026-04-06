"use client";

import { useRouter } from "next/navigation";
import { X, Play, Layers } from "lucide-react";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";

interface PkExitModalProps {
  gameId: string;
}

export function PkExitModal({ gameId }: PkExitModalProps) {
  const router = useRouter();
  const closeOverlay = useGameStore((s) => s.closeOverlay);

  function handleExit(href: string) {
    router.push(href);
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 px-4">
      <div className="flex w-full max-w-[380px] flex-col gap-6 rounded-[20px] bg-card px-8 py-9">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-foreground">退出游戏</h2>
          <button
            type="button"
            onClick={closeOverlay}
            className="flex h-8 w-8 items-center justify-center rounded-lg"
          >
            <X className="h-5 w-5 text-muted-foreground" />
          </button>
        </div>
        <div className="flex w-full flex-col gap-3">
          <button
            type="button"
            onClick={() => handleExit(`/hall/games/${gameId}`)}
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl border border-border bg-muted"
          >
            <Layers className="h-[18px] w-[18px] text-muted-foreground" />
            <span className="text-[15px] font-medium text-muted-foreground">
              返回关卡列表
            </span>
          </button>
          <button
            type="button"
            onClick={closeOverlay}
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-teal-600"
          >
            <Play className="h-[18px] w-[18px] text-white" />
            <span className="text-[15px] font-semibold text-white">
              继续当前游戏
            </span>
          </button>
        </div>
      </div>
    </div>
  );
}
