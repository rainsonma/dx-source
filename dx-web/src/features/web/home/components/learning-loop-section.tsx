// dx-web/src/features/web/home/components/learning-loop-section.tsx
"use client";

import { motion } from "motion/react";
import { ArrowRight, BookOpen, RefreshCw, CircleCheck, Check } from "lucide-react";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

const LEVELS = [
  { label: "Lv.01", done: true },
  { label: "Lv.02", done: true },
  { label: "Lv.03", done: false },
];

const VOCAB = [
  {
    title: "生词本",
    desc: "持续沉淀你不会的词汇",
    Icon: BookOpen,
    iconColor: "#EC4899",
    bg: "bg-pink-500/10",
  },
  {
    title: "复习本",
    desc: "[1, 3, 7, 14, 30, 90] 天节奏智能推送",
    Icon: RefreshCw,
    iconColor: "#7B61FF",
    bg: "bg-violet-500/10",
  },
  {
    title: "已掌握",
    desc: "看得见的词汇量增长",
    Icon: CircleCheck,
    iconColor: "#0d9488",
    bg: "bg-teal-500/10",
  },
];

export function LearningLoopSection() {
  const reduced = usePrefersReducedMotion();

  return (
    <section className="w-full bg-gradient-to-b from-teal-50 to-violet-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-[60px] lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-violet-500">
            学习闭环 · 从陌生到掌握
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            一套系统，追踪你每一个学习单元的命运
          </h2>
        </motion.div>

        <motion.div
          className="flex w-full flex-col items-stretch gap-6 lg:flex-row lg:items-center lg:justify-between"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
        >
          <motion.div
            variants={staggerChildVariants}
            className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-6 lg:w-[280px]"
          >
            <span className="text-xs font-semibold uppercase tracking-wider text-violet-600">
              闯关式课程
            </span>
            {LEVELS.map((lv) => (
              <div
                key={lv.label}
                className={`flex items-center justify-between rounded-lg border px-4 py-3 ${
                  lv.done
                    ? "border-teal-200 bg-teal-50"
                    : "border-violet-200 bg-white"
                }`}
              >
                <span className="text-sm font-semibold text-slate-900">
                  {lv.label}
                </span>
                {lv.done ? (
                  <Check className="h-4 w-4 text-teal-600" />
                ) : (
                  <motion.span
                    className="h-2 w-2 rounded-full bg-violet-500"
                    animate={reduced ? undefined : { scale: [1, 1.4, 1] }}
                    transition={reduced ? undefined : { duration: 1.6, repeat: Infinity }}
                  />
                )}
              </div>
            ))}
          </motion.div>

          <motion.div
            variants={staggerChildVariants}
            className="relative flex flex-1 items-center justify-center py-4 lg:px-6"
          >
            <div className="relative h-px w-full bg-gradient-to-r from-teal-300 via-teal-500 to-violet-500">
              <motion.span
                className="absolute -top-1 h-3 w-3 rounded-full bg-teal-500 shadow-[0_0_12px_rgba(13,148,136,0.8)]"
                initial={reduced ? { left: "100%" } : { left: "0%" }}
                whileInView={{ left: "100%" }}
                viewport={{ once: true, amount: 0.4 }}
                transition={reduced ? { duration: 0 } : { duration: 1.4, ease: "easeInOut" }}
              />
            </div>
            <div className="absolute -bottom-6 left-1/2 -translate-x-1/2 whitespace-nowrap text-xs text-slate-500">
              通过每轮学习不断沉淀 <ArrowRight className="inline h-3 w-3" />
            </div>
          </motion.div>

          <motion.div
            variants={staggerChildVariants}
            className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-6 lg:w-[320px]"
          >
            <span className="text-xs font-semibold uppercase tracking-wider text-teal-600">
              词汇三本
            </span>
            {VOCAB.map(({ title, desc, Icon, iconColor, bg }) => (
              <div
                key={title}
                className="flex items-center gap-3 rounded-lg border border-slate-100 bg-slate-50 px-3 py-2.5"
              >
                <div
                  className={`flex h-9 w-9 items-center justify-center rounded-lg ${bg}`}
                >
                  <Icon className="h-4 w-4" style={{ color: iconColor }} />
                </div>
                <div className="flex flex-col">
                  <span className="text-sm font-semibold text-slate-900">
                    {title}
                  </span>
                  <span className="text-xs text-slate-500">{desc}</span>
                </div>
              </div>
            ))}
          </motion.div>
        </motion.div>

        <motion.p
          className="max-w-[640px] text-center text-[15px] text-slate-500"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          艾宾浩斯遗忘曲线智能推送复习，你只需要玩。
        </motion.p>
      </div>
    </section>
  );
}
