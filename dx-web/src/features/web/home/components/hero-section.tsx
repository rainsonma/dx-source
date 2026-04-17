// dx-web/src/features/web/home/components/hero-section.tsx
import Link from "next/link";
import { Gamepad2, ArrowRight } from "lucide-react";
import { HeroGameDemo } from "@/features/web/home/components/hero-game-demo";

interface HeroSectionProps {
  isLoggedIn: boolean;
}

export function HeroSection({ isLoggedIn }: HeroSectionProps) {
  const primaryHref = isLoggedIn ? "/hall" : "/auth/signup";

  return (
    <section className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-8 px-5 pb-[80px] pt-16 md:px-10 md:pb-[100px] md:pt-20 lg:px-[120px]">
      <div className="flex items-center gap-2 rounded-full border border-slate-200 bg-white/70 px-5 py-2 backdrop-blur">
        <div className="h-2 w-2 rounded-full bg-green-400" />
        <span className="text-[13px] font-medium text-slate-500">
          斗学，让你玩着玩着就学会英语
        </span>
      </div>
      <div className="flex w-full flex-col items-center">
        <h1 className="text-center text-5xl font-extrabold leading-tight tracking-[-2px] text-slate-900 md:text-6xl lg:text-[72px] lg:tracking-[-3px]">
          别「学」英语了
        </h1>
        <h1 className="bg-gradient-to-r from-teal-400 to-violet-500 bg-clip-text text-center text-5xl font-extrabold leading-tight tracking-[-2px] text-transparent md:text-6xl lg:text-[72px] lg:tracking-[-3px]">
          玩着玩着就会了
        </h1>
      </div>
      <p className="max-w-[680px] text-center text-sm leading-[1.6] text-slate-500 md:text-base lg:text-lg">
        多种游戏模式 · AI 定制内容 · 和朋友一起闯关
        <br className="hidden md:block" />
        每天 10 分钟，英语悄悄就流利了
      </p>
      <p className="sr-only">
        下方演示了连词成句和词汇对轰两种玩法的循环动画。
      </p>
      <div className="flex flex-col items-center gap-4 md:flex-row">
        <Link
          href={primaryHref}
          className="flex items-center gap-2.5 rounded-xl bg-teal-600 px-9 py-4 shadow-[0_4px_30px_rgba(13,148,136,0.27)] transition-colors hover:bg-teal-700"
        >
          <Gamepad2 className="h-5 w-5 text-white" />
          <span className="text-base font-semibold text-white">开始斗学之旅</span>
        </Link>
        <Link
          href="#features"
          className="flex items-center gap-2.5 rounded-xl border-[1.5px] border-slate-200 bg-white/70 px-9 py-4 transition-colors hover:bg-white"
        >
          <span className="text-base font-medium text-slate-900">了解更多</span>
          <ArrowRight className="h-[18px] w-[18px] text-slate-900" />
        </Link>
      </div>
      <div className="relative mt-4 w-full max-w-[900px]">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-[-40px] -z-10 rounded-[32px] bg-[radial-gradient(ellipse_at_center,rgba(94,234,212,0.35),transparent_70%)]"
        />
        <HeroGameDemo />
      </div>
    </section>
  );
}
