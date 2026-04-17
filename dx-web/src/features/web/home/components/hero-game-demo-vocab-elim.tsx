"use client";

import { motion } from "motion/react";
import { Sparkles } from "lucide-react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface Props {
  active: boolean;
}

// 6-tile 2x3 grid. pair IDs determine which tiles eliminate together.
const TILES: ReadonlyArray<{ text: string; pair: 0 | 1 | 2 }> = [
  { text: "apple", pair: 0 },
  { text: "微笑", pair: 1 },
  { text: "学习", pair: 2 },
  { text: "苹果", pair: 0 },
  { text: "study", pair: 2 },
  { text: "smile", pair: 1 },
];

// Absolute timeline in seconds.
const TOTAL = 4; // duration of the full tile keyframe sequence
const ENTRY_END = 0.45; // all tiles finish entering by ~0.45s
const POP_AT: Record<0 | 1 | 2, number> = { 0: 0.95, 1: 1.85, 2: 2.75 };
const POP_PEAK = 0.12; // pop peak offset after match start
const FADE = 0.2; // fade delay after pop peak

export function HeroGameDemoVocabElim({ active }: Props) {
  const reduced = usePrefersReducedMotion();
  if (!active) return <div className="h-full w-full" />;

  return (
    <motion.div
      className="flex w-full max-w-[640px] flex-col gap-4"
      initial={reduced ? false : { opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
    >
      <p className="text-xs text-slate-500">
        找到并消除含义相同的单词对：
      </p>

      <div className="grid grid-cols-3 gap-2">
        {TILES.map((tile, i) => (
          <ElimTile key={i} tile={tile} index={i} reduced={reduced} />
        ))}
      </div>

      <motion.div
        className="flex items-center gap-3"
        initial={reduced ? false : { opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={reduced ? { duration: 0 } : { delay: 3.3, duration: 0.4 }}
      >
        <span className="rounded-full bg-pink-500 px-3 py-1 text-xs font-semibold text-white">
          全部消除
        </span>
        <span className="flex items-center gap-1 rounded-full bg-teal-100 px-3 py-1 text-xs font-semibold text-teal-700">
          <Sparkles className="h-3 w-3" /> 连击 x3
        </span>
        <span className="text-xs text-slate-500">+240 ★</span>
      </motion.div>
    </motion.div>
  );
}

function ElimTile({
  tile,
  index,
  reduced,
}: {
  tile: { text: string; pair: 0 | 1 | 2 };
  index: number;
  reduced: boolean;
}) {
  const entryAt = Math.min(ENTRY_END, 0.1 + index * 0.05);
  const popStart = POP_AT[tile.pair];
  const popPeak = popStart + POP_PEAK;
  const popEnd = popPeak + FADE;

  // Build keyframe timeline as fractions of TOTAL.
  const keyTimes = [
    0,
    entryAt / TOTAL,
    popStart / TOTAL,
    popPeak / TOTAL,
    popEnd / TOTAL,
    1,
  ];

  if (reduced) {
    return (
      <div className="flex h-10 items-center justify-center rounded-lg border border-slate-200 bg-slate-50 px-2 text-sm font-medium text-slate-500">
        {tile.text}
      </div>
    );
  }

  return (
    <motion.div
      className="flex h-10 items-center justify-center rounded-lg border px-2 text-sm font-medium"
      initial={{
        opacity: 0,
        scale: 0.9,
        borderColor: "#e2e8f0",
        backgroundColor: "#ffffff",
        color: "#1e293b",
      }}
      animate={{
        opacity: [0, 1, 1, 1, 0.35, 0.35],
        scale: [0.9, 1, 1, 1.08, 0.95, 0.95],
        borderColor: [
          "#e2e8f0",
          "#e2e8f0",
          "#ec4899",
          "#ec4899",
          "#e2e8f0",
          "#e2e8f0",
        ],
        backgroundColor: [
          "#ffffff",
          "#ffffff",
          "#fce7f3",
          "#fce7f3",
          "#f1f5f9",
          "#f1f5f9",
        ],
        color: ["#1e293b", "#1e293b", "#be185d", "#be185d", "#94a3b8", "#94a3b8"],
      }}
      transition={{ duration: TOTAL, times: keyTimes, ease: "linear" }}
    >
      <span>{tile.text}</span>
    </motion.div>
  );
}
