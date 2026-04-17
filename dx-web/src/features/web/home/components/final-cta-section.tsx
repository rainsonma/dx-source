// dx-web/src/features/web/home/components/final-cta-section.tsx
"use client";

import Link from "next/link";
import { motion } from "motion/react";
import { Rocket, ArrowRight } from "lucide-react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface FinalCtaSectionProps {
  isLoggedIn: boolean;
}

export function FinalCtaSection({ isLoggedIn }: FinalCtaSectionProps) {
  const reduced = usePrefersReducedMotion();
  const primaryHref = isLoggedIn ? "/hall" : "/auth/signup";

  return (
    <section className="w-full bg-gradient-to-b from-slate-50 to-teal-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-8 px-5 md:px-10 lg:px-[120px]">
        <div className="h-1 w-full max-w-[600px] rounded-full bg-gradient-to-r from-teal-400/0 via-teal-400 via-30% to-violet-500/0" />
        <motion.h2
          className="text-center text-4xl font-extrabold tracking-[-1px] text-slate-900 md:text-5xl lg:text-[52px] lg:tracking-[-2px]"
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.4 }}
          transition={{ duration: 0.5 }}
        >
          准备好把英语打通关了吗？
        </motion.h2>
        <p className="max-w-[600px] text-center text-base leading-[1.6] text-slate-500 md:text-lg">
          30 秒注册，今晚就能陪你玩到手滑。
        </p>
        <div className="flex flex-col items-center gap-4 md:flex-row">
          <Link
            href={primaryHref}
            className="group flex items-center gap-3 rounded-[14px] bg-teal-600 px-10 py-[18px] shadow-[0_8px_40px_rgba(13,148,136,0.27)] transition-colors hover:bg-teal-700"
          >
            <motion.span
              animate={reduced ? undefined : { y: [0, -2, 0] }}
              transition={reduced ? undefined : { duration: 1.6, repeat: Infinity, ease: "easeInOut" }}
              className="flex items-center"
            >
              <Rocket className="h-[22px] w-[22px] text-white" />
            </motion.span>
            <span className="text-[17px] font-bold text-white">
              开始你的斗学冒险
            </span>
          </Link>
          <Link
            href="/wiki"
            className="flex items-center gap-2.5 rounded-[14px] border-[1.5px] border-slate-200 bg-white/70 px-8 py-[18px] transition-colors hover:bg-white"
          >
            <span className="text-[15px] font-medium text-slate-900">
              查看 Wiki
            </span>
            <ArrowRight className="h-[18px] w-[18px] text-slate-900" />
          </Link>
        </div>
      </div>
    </section>
  );
}
