// dx-web/src/features/web/home/components/hero-game-demo-vocab-battle.tsx
"use client";

import { motion } from "motion/react";
import { Swords, Bot, User } from "lucide-react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface Props {
  active: boolean;
}

const LETTERS = ["c", "o", "u", "r", "a", "g", "e"];

export function HeroGameDemoVocabBattle({ active }: Props) {
  const reduced = usePrefersReducedMotion();
  if (!active) return <div className="h-full w-full" />;

  return (
    <motion.div
      className="grid w-full max-w-[680px] grid-cols-2 items-center gap-6 text-sm"
      initial={reduced ? false : { opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
    >
      <div className="flex flex-col items-start gap-2">
        <div className="flex items-center gap-2 text-slate-500">
          <User className="h-4 w-4" /> 你
        </div>
        <div className="flex flex-wrap gap-1">
          {LETTERS.map((ch, i) => (
            <motion.span
              key={i}
              className="inline-flex h-7 w-7 items-center justify-center rounded-md border border-teal-200 bg-teal-50 text-sm font-semibold text-teal-700"
              initial={reduced ? false : { opacity: 0, y: -6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={
                reduced
                  ? { duration: 0 }
                  : { delay: 0.6 + 0.1 * i, duration: 0.2 }
              }
            >
              {ch}
            </motion.span>
          ))}
        </div>
        <motion.div
          className="mt-2 flex items-center gap-2 rounded-full bg-teal-600 px-3 py-1 text-xs font-semibold text-white"
          initial={reduced ? false : { opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={reduced ? { duration: 0 } : { delay: 1.7, duration: 0.3 }}
        >
          <Swords className="h-3 w-3" /> combo x3
        </motion.div>
      </div>

      <div className="relative flex flex-col items-end gap-2">
        <div className="flex items-center gap-2 text-slate-500">
          对手 <Bot className="h-4 w-4" />
        </div>
        <motion.span
          className="rounded-lg bg-rose-100 px-3 py-2 text-sm font-semibold text-rose-700"
          initial={reduced ? false : { opacity: 0, x: 24 }}
          animate={{ opacity: 1, x: 0 }}
          transition={reduced ? { duration: 0 } : { duration: 0.4 }}
        >
          勇气
        </motion.span>
        <div className="relative h-2 w-full overflow-hidden rounded-full bg-slate-100">
          <motion.div
            className="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-rose-400 to-rose-600"
            initial={reduced ? { width: "55%" } : { width: "100%" }}
            animate={{ width: reduced ? "55%" : ["100%", "100%", "55%"] }}
            transition={
              reduced
                ? { duration: 0 }
                : { duration: 2, times: [0, 0.7, 1], ease: "easeOut" }
            }
          />
        </div>
        <motion.span
          className="text-xs font-semibold text-rose-600"
          initial={reduced ? false : { opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={reduced ? { duration: 0 } : { delay: 1.4, duration: 0.3 }}
        >
          -15 HP
        </motion.span>
      </div>

      {!reduced && (
        <motion.div
          aria-hidden="true"
          className="pointer-events-none col-span-2 -mt-12 h-2 rounded-full bg-gradient-to-r from-teal-400 to-transparent"
          initial={{ scaleX: 0, originX: 0 }}
          animate={{ scaleX: [0, 1, 0] }}
          transition={{ duration: 1.6, delay: 1.2, times: [0, 0.6, 1] }}
        />
      )}
    </motion.div>
  );
}
