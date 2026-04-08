import Link from "next/link";
import { Flame, MessageCircle, Play } from "lucide-react";

/** 今日打卡 — grouped card with two daily task items */
export function DailyChallengeCard() {
  return (
    <div className="flex h-full w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-5">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-orange-50">
          <Flame className="h-5 w-5 text-orange-500" />
        </div>
        <h3 className="text-base font-bold text-foreground">今日打卡</h3>
      </div>

      {/* Task 1 — Game challenge */}
      <div className="flex flex-col gap-3 rounded-xl bg-teal-600 p-5">
        <p className="text-sm font-bold leading-relaxed text-white">
          完成今天的连词成句
          <br />
          赢取双倍经验值！
        </p>
        <Link
          href="/hall/games"
          className="flex h-10 w-full items-center justify-center gap-2 rounded-[10px] bg-white text-[13px] font-semibold text-teal-700 hover:bg-white/90"
        >
          <Play className="h-4 w-4" />
          开始挑战
        </Link>
      </div>

      {/* Task 2 — Community post */}
      <div className="flex flex-col gap-3 rounded-xl bg-gradient-to-br from-teal-500 to-teal-700 p-5">
        <p className="text-sm font-bold leading-relaxed text-white">
          前往「斗学社」发表一条英文动态贴，进步需要坚持不懈!
        </p>
        <Link
          href="/hall/community"
          className="flex h-10 w-full items-center justify-center gap-2 rounded-[10px] bg-white text-[13px] font-semibold text-teal-700 hover:bg-white/90"
        >
          <MessageCircle className="h-4 w-4" />
          去发帖
        </Link>
      </div>
    </div>
  );
}
