// dx-web/src/features/web/home/components/why-different-section.tsx
"use client";

import { motion } from "motion/react";
import { ArrowRight } from "lucide-react";
import {
  revealVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

const ROWS = [
  { before: "背了就忘，靠意志力硬撑", after: "游戏化循环，大脑自发想再玩一局" },
  { before: "学的和用的两张皮", after: "连词成句、对话、对战都是真实语料" },
  { before: "一个人孤独地学", after: "好友开黑 · 学习群 · 排行榜 · 每日连胜" },
] as const;

export function WhyDifferentSection() {
  const reduced = usePrefersReducedMotion();

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
            <Row key={row.before} row={row} reduced={reduced} />
          ))}
        </motion.ul>
      </div>
    </section>
  );
}

function Row({
  row,
  reduced,
}: {
  row: { before: string; after: string };
  reduced: boolean;
}) {
  return (
    <motion.li
      variants={{
        hidden: {},
        show: { transition: { staggerChildren: reduced ? 0 : 0.12 } },
      }}
      whileHover={reduced ? undefined : { y: -2 }}
      transition={reduced ? undefined : { type: "spring", stiffness: 260, damping: 22 }}
      className="group relative flex flex-col items-stretch gap-3 overflow-hidden rounded-2xl border border-slate-200 bg-white p-5 shadow-[0_1px_2px_rgba(15,23,42,0.03)] transition-shadow hover:shadow-[0_8px_24px_rgba(13,148,136,0.08)] md:flex-row md:items-center md:gap-6 md:p-6"
    >
      {/* Before cell — slides in from the left with the strike-through drawing across */}
      <motion.div
        variants={{
          hidden: reduced
            ? { opacity: 1, x: 0 }
            : { opacity: 0, x: -16 },
          show: { opacity: 1, x: 0 },
        }}
        transition={reduced ? { duration: 0 } : { duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
        className="relative flex-1 text-[15px] text-slate-400"
      >
        <span className="relative inline-block">
          {row.before}
          <motion.span
            aria-hidden="true"
            className="pointer-events-none absolute left-0 top-1/2 h-[1.5px] bg-slate-300"
            initial={reduced ? { scaleX: 1, originX: 0 } : { scaleX: 0, originX: 0 }}
            whileInView={{ scaleX: 1 }}
            viewport={{ once: true, amount: 0.6 }}
            transition={reduced ? { duration: 0 } : { duration: 0.5, delay: 0.25, ease: "easeOut" }}
            style={{ width: "100%" }}
          />
        </span>
      </motion.div>

      {/* Arrow — travelling dot slides L→R when row reveals */}
      <motion.div
        variants={{
          hidden: { opacity: 0 },
          show: { opacity: 1 },
        }}
        transition={reduced ? { duration: 0 } : { duration: 0.3, delay: 0.15 }}
        className="relative hidden h-5 w-5 shrink-0 items-center justify-center md:flex"
      >
        <ArrowRight className="h-5 w-5 text-teal-400" />
        {!reduced && (
          <motion.span
            aria-hidden="true"
            className="absolute inset-0 rounded-full bg-teal-300/30 blur-md"
            initial={{ scale: 0, opacity: 0 }}
            whileInView={{ scale: [0, 1.2, 0], opacity: [0, 0.8, 0] }}
            viewport={{ once: true, amount: 0.6 }}
            transition={{ duration: 0.9, delay: 0.2, times: [0, 0.5, 1] }}
          />
        )}
      </motion.div>

      {/* After cell — rises with a soft teal bloom behind it */}
      <motion.div
        variants={{
          hidden: reduced
            ? { opacity: 1, y: 0 }
            : { opacity: 0, y: 16 },
          show: { opacity: 1, y: 0 },
        }}
        transition={
          reduced
            ? { duration: 0 }
            : { type: "spring", stiffness: 200, damping: 22, delay: 0.05 }
        }
        className="relative flex-1"
      >
        {!reduced && (
          <motion.span
            aria-hidden="true"
            className="pointer-events-none absolute -inset-2 rounded-2xl bg-[radial-gradient(ellipse_at_center,rgba(94,234,212,0.55),transparent_70%)]"
            initial={{ opacity: 0, scale: 0.8 }}
            whileInView={{ opacity: [0, 1, 0], scale: [0.8, 1.05, 1.05] }}
            viewport={{ once: true, amount: 0.6 }}
            transition={{ duration: 1.1, delay: 0.25, times: [0, 0.4, 1] }}
          />
        )}
        <div className="relative rounded-xl bg-gradient-to-r from-teal-50 to-violet-50 px-4 py-3 text-[15px] font-medium text-slate-900 transition-colors group-hover:from-teal-100 group-hover:to-violet-100">
          {row.after}
        </div>
      </motion.div>
    </motion.li>
  );
}
