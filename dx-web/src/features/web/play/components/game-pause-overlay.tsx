"use client";

import { useRouter } from "next/navigation";
import { CirclePause, Play, LogOut } from "lucide-react";

interface GamePauseOverlayProps {
  elapsedTime: string;
  gameId: string;
  onResume: () => void;
}

export function GamePauseOverlay({ elapsedTime, gameId, onResume }: GamePauseOverlayProps) {
  const router = useRouter();

  /** Navigate to game details page — store cleanup handled by GamePlayShell's stale-state useEffect */
  const handleExit = () => {
    router.push(`/hall/games/${gameId}`);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 px-4">
      <div className="flex w-full max-w-[380px] flex-col items-center gap-6 rounded-[20px] bg-card px-8 py-9">
        <CirclePause className="h-12 w-12 text-teal-600" />
        <h2 className="text-[22px] font-bold text-foreground">游戏暂停</h2>
        <span className="-tracking-wider text-4xl font-extrabold text-muted-foreground">
          {elapsedTime}
        </span>
        <div className="flex w-full flex-col gap-3">
          <button
            type="button"
            onClick={onResume}
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-teal-600"
          >
            <Play className="h-[18px] w-[18px] text-white" />
            <span className="text-[15px] font-semibold text-white">
              继续游戏
            </span>
          </button>
          <button
            type="button"
            onClick={handleExit}
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl border border-border bg-muted"
          >
            <LogOut className="h-[18px] w-[18px] text-muted-foreground" />
            <span className="text-[15px] font-medium text-muted-foreground">
              退出游戏
            </span>
          </button>
        </div>
      </div>
    </div>
  );
}
