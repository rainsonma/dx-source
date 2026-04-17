// dx-web/src/features/web/home/components/hero-game-demo.tsx
"use client";

import dynamic from "next/dynamic";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useInView } from "@/features/web/home/hooks/use-in-view";
import { useRotatingScene } from "@/features/web/home/hooks/use-rotating-scene";

const SceneWordSentence = dynamic(
  () =>
    import("@/features/web/home/components/hero-game-demo-word-sentence").then(
      (m) => m.HeroGameDemoWordSentence,
    ),
  { ssr: false },
);

const SceneVocabMatch = dynamic(
  () =>
    import("@/features/web/home/components/hero-game-demo-vocab-match").then(
      (m) => m.HeroGameDemoVocabMatch,
    ),
  { ssr: false },
);

const SceneVocabElim = dynamic(
  () =>
    import("@/features/web/home/components/hero-game-demo-vocab-elim").then(
      (m) => m.HeroGameDemoVocabElim,
    ),
  { ssr: false },
);

const SceneVocabBattle = dynamic(
  () =>
    import("@/features/web/home/components/hero-game-demo-vocab-battle").then(
      (m) => m.HeroGameDemoVocabBattle,
    ),
  { ssr: false },
);

const SCENES = [
  {
    key: "word-sentence",
    label: "连词成句",
    ariaLabel: "切换到连词成句演示",
    Scene: SceneWordSentence,
  },
  {
    key: "vocab-battle",
    label: "词汇对轰",
    ariaLabel: "切换到词汇对轰演示",
    Scene: SceneVocabBattle,
  },
  {
    key: "vocab-match",
    label: "词汇配对",
    ariaLabel: "切换到词汇配对演示",
    Scene: SceneVocabMatch,
  },
  {
    key: "vocab-elim",
    label: "词汇消消乐",
    ariaLabel: "切换到词汇消消乐演示",
    Scene: SceneVocabElim,
  },
] as const;

export function HeroGameDemo() {
  const { ref, inView } = useInView<HTMLDivElement>({ threshold: 0.2 });
  const [hovered, setHovered] = useState(false);
  const [focused, setFocused] = useState(false);
  const paused = !inView || hovered || focused;
  const { index, setIndex } = useRotatingScene({
    total: SCENES.length,
    intervalMs: 4500,
    paused,
  });

  const ActiveScene = SCENES[index].Scene;

  return (
    <div
      ref={ref}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      className="relative aspect-[5/3] w-full overflow-hidden rounded-[20px] border border-slate-200 bg-white/90 p-4 shadow-[0_12px_40px_rgba(13,148,136,0.12)] backdrop-blur md:aspect-[15/7] md:p-6"
    >
      <div className="absolute left-4 top-4 z-10 flex gap-1 rounded-full border border-slate-200 bg-white p-1 text-xs shadow-sm md:left-6 md:top-6">
        {SCENES.map((s, i) => (
          <button
            key={s.key}
            type="button"
            onClick={() => setIndex(i)}
            aria-label={s.ariaLabel}
            className={cn(
              "rounded-full px-3 py-1 font-medium transition-colors",
              i === index
                ? "bg-teal-600 text-white"
                : "text-slate-500 hover:text-slate-900",
            )}
          >
            {s.label}
          </button>
        ))}
      </div>
      <div
        aria-hidden="true"
        className="flex h-full w-full items-center justify-center"
      >
        <ActiveScene key={SCENES[index].key} active={inView} />
      </div>
    </div>
  );
}
