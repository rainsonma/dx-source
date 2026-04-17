// dx-web/src/features/web/home/components/features-section.tsx
"use client";

import { motion } from "motion/react";
import {
  Keyboard,
  Swords,
  Shuffle,
  Crosshair,
  type LucideIcon,
} from "lucide-react";
import { useState } from "react";
import {
  GameModePreview,
  type GameModeKey,
} from "@/features/web/home/components/game-mode-preview";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";

interface Feature {
  key: GameModeKey;
  icon: LucideIcon;
  iconClassName: string;
  bgClassName: string;
  title: string;
  description: string;
}

const FEATURES: Feature[] = [
  {
    key: "word-sentence",
    icon: Keyboard,
    iconClassName: "text-violet-600",
    bgClassName: "bg-violet-100",
    title: "连词成句",
    description:
      "看到中文秒拼出英文句子，越快越高分。真实语料替你练习语序和搭配。",
  },
  {
    key: "vocab-match",
    icon: Swords,
    iconClassName: "text-teal-600",
    bgClassName: "bg-blue-100",
    title: "词汇配对",
    description:
      "英文与中文快速配对，限时给分。巩固词汇量和中译英反应速度。",
  },
  {
    key: "vocab-elim",
    icon: Shuffle,
    iconClassName: "text-emerald-600",
    bgClassName: "bg-pink-100",
    title: "词汇消消乐",
    description:
      "记忆配对消除，越快消除越高分。玩着玩着就把生词牢牢记住。",
  },
  {
    key: "vocab-battle",
    icon: Crosshair,
    iconClassName: "text-red-500",
    bgClassName: "bg-red-100",
    title: "词汇对轰",
    description:
      "和 AI 对手拼炮弹。拼对拼快就发射，紧张刺激的单词对战。",
  },
];

function FeatureCard({ feature }: { feature: Feature }) {
  const Icon = feature.icon;
  const [hovered, setHovered] = useState(false);

  return (
    <motion.div
      variants={staggerChildVariants}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      whileHover={{ y: -3 }}
      className="flex flex-col gap-5 rounded-2xl border border-slate-200 bg-white p-8 transition-shadow hover:shadow-[0_8px_24px_rgba(13,148,136,0.08)]"
    >
      <div
        className={`flex h-14 w-14 items-center justify-center rounded-2xl ${feature.bgClassName}`}
      >
        <Icon className={`h-7 w-7 ${feature.iconClassName}`} />
      </div>
      <h3 className="text-xl font-bold text-slate-900">{feature.title}</h3>
      <p className="text-[15px] leading-relaxed text-slate-500">
        {feature.description}
      </p>
      {/* Mobile: autoplay via md:hidden CSS trick — use `play` prop always true on mobile */}
      <div className="md:hidden">
        <GameModePreview mode={feature.key} play={true} />
      </div>
      <div className="hidden md:block">
        <GameModePreview mode={feature.key} play={hovered} />
      </div>
    </motion.div>
  );
}

export function FeaturesSection() {
  return (
    <section
      id="features"
      className="w-full bg-gradient-to-b from-slate-50 to-white py-[80px] md:py-[100px]"
    >
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:gap-[60px] md:px-10 lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-violet-500">
            核心玩法 · 4 种模式，覆盖听说读写
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            挑一个就能上手，全都玩转就起飞
          </h2>
        </motion.div>
        <motion.div
          className="grid w-full grid-cols-1 gap-6 md:grid-cols-2"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
        >
          {FEATURES.map((feature) => (
            <FeatureCard key={feature.title} feature={feature} />
          ))}
        </motion.div>
      </div>
    </section>
  );
}
