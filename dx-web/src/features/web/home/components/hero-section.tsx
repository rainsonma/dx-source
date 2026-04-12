import Link from "next/link";
import { Gamepad2, ArrowRight } from "lucide-react";

export function HeroSection() {
  return (
    <section className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-8 px-[120px] pb-[100px] pt-20">
      <div className="flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-5 py-2">
        <div className="h-2 w-2 rounded-full bg-green-400" />
        <span className="text-[13px] font-medium text-slate-500">
          斗学助你快速升级英语技能树
        </span>
      </div>
      <div className="flex w-full flex-col items-center">
        <h1 className="text-center text-[72px] font-extrabold leading-tight tracking-[-3px] text-slate-900">
          别「学」英语了
        </h1>
        <h1 className="bg-gradient-to-r from-teal-400 to-violet-500 bg-clip-text text-center text-[72px] font-extrabold leading-tight tracking-[-3px] text-transparent">
          玩着玩着就会了
        </h1>
      </div>
      <p className="max-w-[680px] text-center text-lg leading-[1.6] text-slate-500">
        多种游戏模式 + AI 智能辅助 + 社区互动
        <br />
        斗学重新定义英语学习体验，不知不觉英语熟练了，不知不觉英语流利了
      </p>
      <div className="flex flex-col items-center gap-4 md:flex-row">
        <Link
          href="/hall"
          className="flex items-center gap-2.5 rounded-xl bg-teal-600 px-9 py-4 shadow-[0_4px_30px_rgba(13,148,136,0.27)] hover:bg-teal-700"
        >
          <Gamepad2 className="h-5 w-5 text-white" />
          <span className="text-base font-semibold text-white">开始课程游戏</span>
        </Link>
        <Link
          href="/features"
          className="flex items-center gap-2.5 rounded-xl border-[1.5px] border-slate-200 bg-white/70 px-9 py-4 hover:bg-white"
        >
          <span className="text-base font-medium text-slate-900">
            了解更多
          </span>
          <ArrowRight className="h-[18px] w-[18px] text-slate-900" />
        </Link>
      </div>
      <div className="mt-4 flex h-[420px] w-[900px] items-start justify-center rounded-[20px] border border-slate-200 bg-[#F0F2F5] p-10 shadow-[0_12px_40px_rgba(123,97,255,0.09)]">
        <div className="h-[340px] w-[820px] rounded-xl border border-slate-200 bg-white p-6">
          <div className="h-full w-full rounded-lg bg-slate-50" />
        </div>
      </div>
    </section>
  );
}
