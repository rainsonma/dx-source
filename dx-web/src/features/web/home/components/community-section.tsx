// dx-web/src/features/web/home/components/community-section.tsx
"use client";

import { motion } from "motion/react";
import {
  Trophy,
  MessageSquare,
  Users,
  Flame,
  type LucideIcon,
} from "lucide-react";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface Card {
  Icon: LucideIcon;
  iconColor: string;
  bgClassName: string;
  title: string;
  description: string;
  extra?: React.ReactNode;
}

function StreakHeatmap() {
  const reduced = usePrefersReducedMotion();
  const cells = Array.from({ length: 7 });
  return (
    <motion.div
      className="mt-2 flex gap-1"
      initial="hidden"
      whileInView={reduced ? undefined : "show"}
      viewport={{ once: true, amount: 0.6 }}
      variants={{
        hidden: {},
        show: { transition: { staggerChildren: 0.08 } },
      }}
    >
      {cells.map((_, i) => (
        <motion.span
          key={i}
          className="h-5 w-5 rounded"
          variants={{
            hidden: { backgroundColor: i < 6 ? "#0d9488" : "#99f6e4" },
            show: {
              backgroundColor: i < 6 ? "#0d9488" : "#99f6e4",
            },
          }}
          transition={{ duration: 0 }}
        />
      ))}
    </motion.div>
  );
}

const CARDS: Card[] = [
  {
    Icon: Trophy,
    iconColor: "#F59E0B",
    bgClassName: "bg-amber-500/10",
    title: "排行榜",
    description: "经验值与在线时长，按日、周、月六种榜单，登上领奖台。",
  },
  {
    Icon: MessageSquare,
    iconColor: "#EA580C",
    bgClassName: "bg-orange-600/10",
    title: "斗学社",
    description: "发帖、评论、点赞、关注，把学习心得变成社交动态。",
  },
  {
    Icon: Users,
    iconColor: "#3B82F6",
    bgClassName: "bg-blue-500/10",
    title: "学习群",
    description: "组建小组一起闯关，群内可直接发起课程对战。",
  },
  {
    Icon: Flame,
    iconColor: "#0d9488",
    bgClassName: "bg-teal-500/10",
    title: "连续打卡",
    description: "每天玩至少一次，保持连胜；错过一天从 1 重来。",
    extra: <StreakHeatmap />,
  },
];

function CommunityCard({ card }: { card: Card }) {
  const { Icon, iconColor, bgClassName, title, description, extra } = card;
  return (
    <motion.div
      variants={staggerChildVariants}
      whileHover={{ y: -3 }}
      className="flex flex-col items-start gap-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-[0_4px_16px_rgba(15,23,42,0.03)] transition-shadow hover:shadow-[0_8px_24px_rgba(13,148,136,0.08)]"
    >
      <div
        className={`flex h-12 w-12 items-center justify-center rounded-[12px] ${bgClassName}`}
      >
        <Icon className="h-6 w-6" style={{ color: iconColor }} />
      </div>
      <h3 className="text-lg font-bold text-slate-900">{title}</h3>
      <p className="text-sm leading-relaxed text-slate-500">{description}</p>
      {extra}
    </motion.div>
  );
}

export function CommunitySection() {
  return (
    <section className="w-full bg-gradient-to-b from-violet-50 to-pink-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-[60px] lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-orange-600">
            一起玩才好玩
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            和朋友开黑，排行榜上见
          </h2>
        </motion.div>
        <motion.div
          className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 lg:gap-6"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
        >
          {CARDS.map((c) => (
            <CommunityCard key={c.title} card={c} />
          ))}
        </motion.div>
      </div>
    </section>
  );
}
