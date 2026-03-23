import Link from "next/link";
import { Flame, Play } from "lucide-react";

export function DailyChallengeCard() {
  return (
    <div className="flex w-full flex-col gap-4 rounded-[14px] bg-teal-600 p-6">
      <div className="flex items-center gap-2">
        <Flame className="h-4 w-4 text-amber-300" />
        <span className="text-xs font-semibold text-teal-100">今日挑战</span>
      </div>
      <p className="text-base font-bold leading-[1.5] text-white">
        完成今天的听说读写
        <br />
        赢取双倍经验值！
      </p>
      <Link
        href="/hall/games"
        className="flex h-11 w-full items-center justify-center gap-2 rounded-[10px] bg-white text-[13px] font-semibold text-teal-700 hover:bg-white/90"
      >
        <Play className="h-4 w-4" />
        开始挑战
      </Link>
    </div>
  );
}
