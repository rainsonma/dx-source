// dx-web/src/features/web/home/components/game-mode-preview.tsx
"use client";

import { motion } from "motion/react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

export type GameModeKey =
  | "word-sentence"
  | "vocab-match"
  | "vocab-elim"
  | "vocab-battle";

interface Props {
  mode: GameModeKey;
  play: boolean;
}

export function GameModePreview({ mode, play }: Props) {
  const reduced = usePrefersReducedMotion();
  const animate = play && !reduced;

  switch (mode) {
    case "word-sentence":
      return <WordSentencePreview animate={animate} />;
    case "vocab-match":
      return <VocabMatchPreview animate={animate} />;
    case "vocab-elim":
      return <VocabElimPreview animate={animate} />;
    case "vocab-battle":
      return <VocabBattlePreview animate={animate} />;
  }
}

function Strip({ children }: { children: React.ReactNode }) {
  return (
    <div className="mt-4 flex h-[72px] w-full items-center gap-1.5 overflow-hidden rounded-xl bg-slate-50 px-3">
      {children}
    </div>
  );
}

function WordSentencePreview({ animate }: { animate: boolean }) {
  const words = ["I", "love", "this", "game"];
  return (
    <Strip>
      {words.map((w, i) => (
        <motion.span
          key={w}
          className="rounded-md bg-white px-2 py-1 text-xs font-semibold text-violet-700 shadow-sm"
          animate={animate ? { y: [0, -4, 0] } : undefined}
          transition={
            animate
              ? { duration: 1.2, delay: 0.1 * i, repeat: Infinity, repeatDelay: 0.6 }
              : undefined
          }
        >
          {w}
        </motion.span>
      ))}
    </Strip>
  );
}

function VocabMatchPreview({ animate }: { animate: boolean }) {
  const pairs: [string, string][] = [
    ["apple", "苹果"],
    ["study", "学习"],
    ["smile", "微笑"],
  ];
  return (
    <Strip>
      <div className="grid flex-1 grid-cols-2 gap-1">
        {pairs.map(([en, zh], i) => (
          <div key={en} className="contents">
            <motion.span
              className="rounded-md bg-white px-2 py-0.5 text-[11px] font-medium text-teal-700 shadow-sm"
              animate={animate ? { backgroundColor: ["#ffffff", "#ccfbf1", "#ffffff"] } : undefined}
              transition={
                animate
                  ? { duration: 1.4, delay: 0.2 * i, repeat: Infinity, repeatDelay: 0.8 }
                  : undefined
              }
            >
              {en}
            </motion.span>
            <motion.span
              className="rounded-md bg-white px-2 py-0.5 text-[11px] font-medium text-slate-600 shadow-sm"
              animate={animate ? { backgroundColor: ["#ffffff", "#ccfbf1", "#ffffff"] } : undefined}
              transition={
                animate
                  ? { duration: 1.4, delay: 0.2 * i, repeat: Infinity, repeatDelay: 0.8 }
                  : undefined
              }
            >
              {zh}
            </motion.span>
          </div>
        ))}
      </div>
    </Strip>
  );
}

// 2×4 tile grid with 4 matching pairs; tiles pulse when their pair activates.
const ELIM_TILES: ReadonlyArray<{ text: string; pair: 0 | 1 | 2 | 3 }> = [
  { text: "apple", pair: 0 },
  { text: "微笑", pair: 1 },
  { text: "学习", pair: 2 },
  { text: "书", pair: 3 },
  { text: "苹果", pair: 0 },
  { text: "study", pair: 2 },
  { text: "smile", pair: 1 },
  { text: "book", pair: 3 },
];

function VocabElimPreview({ animate }: { animate: boolean }) {
  return (
    <Strip>
      <div className="grid h-full flex-1 grid-cols-4 grid-rows-2 gap-1 py-1">
        {ELIM_TILES.map((tile, i) => (
          <motion.div
            key={i}
            className="flex items-center justify-center rounded-md bg-white text-[10px] font-semibold text-slate-700 shadow-sm"
            animate={
              animate
                ? {
                    backgroundColor: ["#ffffff", "#fce7f3", "#f1f5f9", "#ffffff"],
                    color: ["#334155", "#be185d", "#94a3b8", "#334155"],
                    scale: [1, 1.06, 0.95, 1],
                  }
                : undefined
            }
            transition={
              animate
                ? {
                    duration: 2.4,
                    times: [0, 0.35, 0.55, 1],
                    delay: 0.3 * tile.pair,
                    repeat: Infinity,
                    repeatDelay: 0.6,
                  }
                : undefined
            }
          >
            {tile.text}
          </motion.div>
        ))}
      </div>
    </Strip>
  );
}

function VocabBattlePreview({ animate }: { animate: boolean }) {
  return (
    <Strip>
      <span className="text-[11px] font-semibold text-teal-700">你</span>
      <motion.div
        className="h-2 flex-1 rounded-full bg-gradient-to-r from-teal-400 to-teal-600"
        animate={animate ? { scaleX: [1, 0.9, 1] } : undefined}
        style={{ transformOrigin: "left" }}
        transition={animate ? { duration: 1.4, repeat: Infinity } : undefined}
      />
      <motion.span
        className="text-sm"
        animate={animate ? { x: [0, 100, 0], opacity: [0, 1, 0] } : undefined}
        transition={animate ? { duration: 1.4, repeat: Infinity, repeatDelay: 0.6 } : undefined}
      >
        ✦
      </motion.span>
      <motion.div
        className="h-2 flex-1 rounded-full bg-gradient-to-r from-rose-400 to-rose-600"
        animate={animate ? { scaleX: [1, 0.6, 0.6, 1] } : undefined}
        style={{ transformOrigin: "right" }}
        transition={
          animate ? { duration: 2.2, times: [0, 0.5, 0.8, 1], repeat: Infinity } : undefined
        }
      />
      <span className="text-[11px] font-semibold text-rose-600">对</span>
    </Strip>
  );
}
