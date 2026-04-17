"use client";

import { motion } from "motion/react";
import { Check, Zap } from "lucide-react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface Props {
  active: boolean;
}

const PAIRS: ReadonlyArray<{ en: string; zh: string }> = [
  { en: "apple", zh: "苹果" },
  { en: "study", zh: "学习" },
  { en: "smile", zh: "微笑" },
];

// Right column in a pre-shuffled order for a more realistic feel
const RIGHT_ORDER = [2, 0, 1]; // shows 微笑, 苹果, 学习

export function HeroGameDemoVocabMatch({ active }: Props) {
  const reduced = usePrefersReducedMotion();
  if (!active) return <div className="h-full w-full" />;

  return (
    <motion.div
      className="flex w-full max-w-[640px] flex-col gap-4"
      initial={reduced ? false : { opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
    >
      <p className="text-xs text-slate-500">给每个英文单词找到对应的中文：</p>

      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-2">
          {PAIRS.map((p, i) => (
            <MatchChip
              key={p.en}
              text={p.en}
              side="left"
              matchAt={0.7 + i * 0.9}
              reduced={reduced}
            />
          ))}
        </div>
        <div className="flex flex-col gap-2">
          {RIGHT_ORDER.map((pairIdx, visualIdx) => {
            const p = PAIRS[pairIdx];
            return (
              <MatchChip
                key={p.zh}
                text={p.zh}
                side="right"
                // Each right chip matches slightly after its left partner.
                // pairIdx tells us which left chip is being matched.
                matchAt={0.95 + pairIdx * 0.9}
                // visualIdx used only to vary the entry stagger
                entryDelay={visualIdx * 0.08}
                reduced={reduced}
              />
            );
          })}
        </div>
      </div>

      <motion.div
        className="flex items-center gap-3"
        initial={reduced ? false : { opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={reduced ? { duration: 0 } : { delay: 3.4, duration: 0.4 }}
      >
        <span className="rounded-full bg-teal-600 px-3 py-1 text-xs font-semibold text-white">
          全部匹配
        </span>
        <span className="flex items-center gap-1 rounded-full bg-violet-100 px-3 py-1 text-xs font-semibold text-violet-700">
          <Zap className="h-3 w-3" /> 连击 x3
        </span>
        <span className="text-xs text-slate-500">+300 ★</span>
      </motion.div>
    </motion.div>
  );
}

function MatchChip({
  text,
  side,
  matchAt,
  entryDelay = 0,
  reduced,
}: {
  text: string;
  side: "left" | "right";
  matchAt: number;
  entryDelay?: number;
  reduced: boolean;
}) {
  return (
    <motion.div
      className="relative flex items-center justify-between rounded-lg border bg-white px-3 py-2 text-sm"
      initial={
        reduced
          ? { opacity: 1, x: 0, borderColor: "#14b8a6", backgroundColor: "#f0fdfa" }
          : { opacity: 0, x: side === "left" ? -12 : 12, borderColor: "#e2e8f0" }
      }
      animate={
        reduced
          ? { opacity: 1, x: 0, borderColor: "#14b8a6", backgroundColor: "#f0fdfa" }
          : {
              opacity: 1,
              x: 0,
              borderColor: ["#e2e8f0", "#14b8a6", "#14b8a6"],
              backgroundColor: ["#ffffff", "#f0fdfa", "#f0fdfa"],
            }
      }
      transition={
        reduced
          ? { duration: 0 }
          : {
              duration: 0.5,
              delay: entryDelay,
              borderColor: { duration: 0.4, times: [0, 0.5, 1], delay: matchAt },
              backgroundColor: { duration: 0.4, times: [0, 0.5, 1], delay: matchAt },
              opacity: { duration: 0.3, delay: entryDelay },
              x: { duration: 0.35, delay: entryDelay, ease: [0.22, 1, 0.36, 1] },
            }
      }
    >
      <span className="font-medium text-slate-900">{text}</span>
      <motion.span
        className="flex h-5 w-5 items-center justify-center rounded-full bg-teal-500 text-white"
        initial={reduced ? { scale: 1 } : { scale: 0 }}
        animate={{ scale: 1 }}
        transition={reduced ? { duration: 0 } : { duration: 0.25, delay: matchAt + 0.1 }}
      >
        <Check className="h-3 w-3" />
      </motion.span>
    </motion.div>
  );
}
