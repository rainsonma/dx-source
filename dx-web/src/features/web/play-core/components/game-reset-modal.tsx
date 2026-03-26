"use client";

import { useTransition } from "react";
import { useRouter } from "next/navigation";
import { RotateCcw, X, Loader2 } from "lucide-react";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { restartLevelSessionAction } from "@/features/web/play-core/actions/session.action";

export function GameResetModal() {
  const router = useRouter();
  const closeOverlay = useGameStore((s) => s.closeOverlay);
  const exitGame = useGameStore((s) => s.exitGame);
  const sessionId = useGameStore((s) => s.sessionId);
  const gameId = useGameStore((s) => s.gameId);
  const levelId = useGameStore((s) => s.levelId);
  const degree = useGameStore((s) => s.degree);
  const pattern = useGameStore((s) => s.pattern);
  const [isPending, startTransition] = useTransition();

  function handleConfirm() {
    if (!sessionId || !levelId || !gameId) return;
    startTransition(async () => {
      await restartLevelSessionAction(sessionId, levelId);
      exitGame();

      const params = new URLSearchParams();
      if (degree) params.set("degree", degree);
      params.set("level", levelId);
      if (pattern) params.set("pattern", pattern);
      router.push(`/hall/play-single/${gameId}?${params.toString()}`);
    });
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 px-4">
      <div className="flex w-full max-w-[440px] flex-col items-center gap-6 rounded-[20px] bg-card px-8 py-9">
        <RotateCcw className="h-12 w-12 text-amber-500" />
        <h2 className="text-[22px] font-bold text-foreground">重新开始</h2>
        <p className="text-center text-sm text-muted-foreground">
          确定重新开始吗？当前进度将被清除，重新从第一题开始。
        </p>
        <div className="flex w-full gap-3">
          <button
            type="button"
            onClick={closeOverlay}
            disabled={isPending}
            className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl border border-border bg-muted disabled:opacity-50"
          >
            <X className="h-[18px] w-[18px] text-muted-foreground" />
            <span className="text-[15px] font-medium text-muted-foreground">
              取消
            </span>
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={isPending}
            className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl bg-amber-500 disabled:opacity-50"
          >
            {isPending ? (
              <Loader2 className="h-[18px] w-[18px] animate-spin text-white" />
            ) : (
              <RotateCcw className="h-[18px] w-[18px] text-white" />
            )}
            <span className="text-[15px] font-semibold text-white">
              {isPending ? "重置中..." : "确认"}
            </span>
          </button>
        </div>
      </div>
    </div>
  );
}
