import Link from "next/link";
import { Rocket } from "lucide-react";

export function FinalCtaSection() {
  return (
    <section className="flex w-full flex-col items-center gap-8 bg-gradient-to-b from-pink-50 to-teal-50 px-[120px] py-[100px]">
      <div className="h-1 w-[800px] rounded-full bg-gradient-to-r from-teal-400/0 via-teal-400 via-30% to-violet-500/0" />
      <h2 className="text-center text-[52px] font-extrabold tracking-[-2px] text-slate-900">
        准备好开始你的冒险了吗？
      </h2>
      <p className="max-w-[600px] text-center text-lg leading-[1.6] text-slate-500">
        加入50万+学习者，将英语变成你最喜欢的游戏。邀请好友还能赚奖励！
      </p>
      <Link
        href="/auth/signup"
        className="flex items-center gap-3 rounded-[14px] bg-teal-600 px-11 py-[18px] shadow-[0_8px_40px_rgba(13,148,136,0.27)] hover:bg-teal-700"
      >
        <Rocket className="h-[22px] w-[22px] text-white" />
        <span className="text-[17px] font-bold text-white">开始你的斗学冒险</span>
      </Link>
    </section>
  );
}
