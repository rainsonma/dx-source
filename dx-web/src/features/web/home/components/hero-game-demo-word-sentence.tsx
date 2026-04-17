// dx-web/src/features/web/home/components/hero-game-demo-word-sentence.tsx
"use client";

import { motion } from "motion/react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface Props {
  active: boolean;
}

const WORDS = [
  { text: "I", className: "bg-teal-100 text-teal-700" },
  { text: "eat", className: "bg-violet-100 text-violet-700" },
  { text: "fresh", className: "bg-pink-100 text-pink-700" },
  { text: "apples", className: "bg-amber-100 text-amber-700" },
  { text: "every", className: "bg-rose-100 text-rose-700" },
  { text: "morning", className: "bg-blue-100 text-blue-700" },
] as const;

export function HeroGameDemoWordSentence({ active }: Props) {
  const reduced = usePrefersReducedMotion();
  if (!active) return <div className="h-full w-full" />;

  return (
    <motion.div
      className="flex w-full max-w-[640px] flex-col gap-4"
      initial={reduced ? false : { opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
    >
      <p className="text-xs text-slate-500">请把下面的单词拼成正确的句子：</p>
      <div className="flex flex-wrap gap-2">
        {WORDS.map((w, i) => (
          <motion.span
            key={w.text}
            className={`rounded-lg px-3 py-2 text-sm font-semibold shadow-sm ${w.className}`}
            initial={reduced ? false : { opacity: 0, y: -12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={
              reduced ? { duration: 0 } : { delay: 0.12 * i, duration: 0.35, ease: [0.22, 1, 0.36, 1] }
            }
          >
            {w.text}
          </motion.span>
        ))}
      </div>
      <div className="relative h-1 w-full overflow-hidden rounded-full bg-slate-100">
        <motion.div
          className="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-teal-400 to-violet-500"
          initial={reduced ? { width: "100%" } : { width: 0 }}
          animate={{ width: "100%" }}
          transition={reduced ? { duration: 0 } : { delay: 1.4, duration: 1.0, ease: "easeOut" }}
        />
      </div>
      <motion.div
        className="flex items-center gap-3"
        initial={reduced ? false : { opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={reduced ? { duration: 0 } : { delay: 2.2, duration: 0.4 }}
      >
        <span className="rounded-full bg-teal-600 px-3 py-1 text-xs font-semibold text-white">
          +128 ★
        </span>
        <span className="rounded-full bg-violet-100 px-3 py-1 text-xs font-semibold text-violet-700">
          combo x2
        </span>
        <span className="text-xs text-slate-500">句子正确！</span>
      </motion.div>
    </motion.div>
  );
}
