// dx-web/src/features/web/home/components/why-different-section.tsx
"use client";

import { motion } from "motion/react";
import { ArrowRight } from "lucide-react";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";

const ROWS = [
  { before: "背了就忘，靠意志力硬撑", after: "游戏化循环，大脑自发想再玩一局" },
  { before: "学的和用的两张皮", after: "连词成句、对话、对战都是真实语料" },
  { before: "一个人孤独地学", after: "好友开黑 · 学习群 · 排行榜 · 每日连胜" },
] as const;

export function WhyDifferentSection() {
  return (
    <section className="w-full bg-gradient-to-b from-white to-slate-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-[60px] lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-teal-600">
            为什么是斗学
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            不再死记硬背，让大脑爱上英语
          </h2>
        </motion.div>

        <motion.ul
          className="flex w-full max-w-[880px] flex-col gap-4"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
        >
          {ROWS.map((row) => (
            <motion.li
              key={row.before}
              variants={staggerChildVariants}
              className="flex flex-col items-stretch gap-3 rounded-2xl border border-slate-200 bg-white p-5 md:flex-row md:items-center md:gap-6 md:p-6"
            >
              <div className="flex-1 text-[15px] text-slate-400 line-through decoration-slate-300">
                {row.before}
              </div>
              <ArrowRight className="hidden h-5 w-5 shrink-0 text-slate-300 md:block" />
              <div className="flex-1 rounded-xl bg-gradient-to-r from-teal-50 to-violet-50 px-4 py-3 text-[15px] font-medium text-slate-900">
                {row.after}
              </div>
            </motion.li>
          ))}
        </motion.ul>
      </div>
    </section>
  );
}
