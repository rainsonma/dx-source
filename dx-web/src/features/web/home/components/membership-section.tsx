// dx-web/src/features/web/home/components/membership-section.tsx
"use client";

import Link from "next/link";
import { motion } from "motion/react";
import { Check, Sparkles, ArrowRight } from "lucide-react";
import {
  USER_GRADES,
  USER_GRADE_PRICES,
  USER_GRADE_LABELS,
  type UserGrade,
} from "@/consts/user-grade";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface TierConfig {
  grade: UserGrade;
  period: string;
  features: string[];
  badge?: { label: string; className: string };
  cardClassName: string;
  ctaClassName: string;
  pulse?: boolean;
}

const TIERS: TierConfig[] = [
  {
    grade: USER_GRADES.FREE,
    period: "",
    features: ["部分关卡", "基础游戏", "加入少量学习群"],
    cardClassName: "bg-white",
    ctaClassName: "border border-slate-200 bg-white text-slate-700 hover:bg-slate-50",
  },
  {
    grade: USER_GRADES.MONTH,
    period: "/月",
    features: ["全部关卡畅玩", "AI 随心学", "月度能量豆"],
    cardClassName: "bg-white",
    ctaClassName: "bg-teal-600 text-white hover:bg-teal-700",
  },
  {
    grade: USER_GRADES.YEAR,
    period: "/年",
    features: ["包含月度全部权益", "更多能量豆月度赠送", "优先客服支持"],
    badge: { label: "推荐", className: "bg-teal-600 text-white" },
    cardClassName: "bg-white ring-2 ring-teal-500",
    ctaClassName: "bg-teal-600 text-white hover:bg-teal-700",
    pulse: true,
  },
  {
    grade: USER_GRADES.LIFETIME,
    period: "",
    features: ["一次付费，终身生效", "最高能量豆赠送", "邀请好友首充享 30% 返佣"],
    badge: { label: "最超值", className: "bg-violet-600 text-white" },
    cardClassName:
      "bg-gradient-to-br from-violet-600 to-teal-600 text-white ring-2 ring-violet-500",
    ctaClassName: "bg-white text-teal-700 hover:bg-white/90",
  },
];

function TierCard({ tier, reduced }: { tier: TierConfig; reduced: boolean }) {
  const price = USER_GRADE_PRICES[tier.grade];
  const name = USER_GRADE_LABELS[tier.grade];
  const isDark = tier.grade === USER_GRADES.LIFETIME;
  const shouldPulse = tier.pulse && !reduced;

  return (
    <motion.div
      variants={staggerChildVariants}
      className={`relative flex flex-col gap-4 rounded-2xl p-6 shadow-[0_4px_16px_rgba(15,23,42,0.05)] ${tier.cardClassName}`}
      animate={shouldPulse ? { y: [0, -4, 0] } : undefined}
      transition={
        shouldPulse
          ? { duration: 2.4, repeat: Infinity, ease: "easeInOut" }
          : undefined
      }
    >
      {tier.badge && (
        <span
          className={`absolute -top-3 left-6 rounded-full px-3 py-1 text-xs font-semibold shadow-sm ${tier.badge.className}`}
        >
          {tier.badge.label}
        </span>
      )}
      <h3 className={`text-lg font-bold ${isDark ? "text-white" : "text-slate-900"}`}>
        {name}
      </h3>
      <div className="flex items-end gap-1">
        <span className={`text-4xl font-extrabold ${isDark ? "text-white" : "text-slate-900"}`}>
          ¥{price}
        </span>
        {tier.period && (
          <span className={`mb-1 text-sm ${isDark ? "text-white/70" : "text-slate-500"}`}>
            {tier.period}
          </span>
        )}
      </div>
      <ul className="flex flex-1 flex-col gap-2">
        {tier.features.map((f) => (
          <li key={f} className={`flex items-start gap-2 text-sm ${isDark ? "text-white/90" : "text-slate-600"}`}>
            <Check className={`mt-0.5 h-4 w-4 shrink-0 ${isDark ? "text-white" : "text-teal-600"}`} />
            <span>{f}</span>
          </li>
        ))}
      </ul>
      <Link
        href="/purchase/membership"
        className={`mt-2 flex items-center justify-center rounded-lg px-4 py-2.5 text-sm font-semibold transition-colors ${tier.ctaClassName}`}
      >
        {tier.grade === USER_GRADES.FREE ? "了解更多" : "立即开通"}
      </Link>
    </motion.div>
  );
}

export function MembershipSection() {
  const reduced = usePrefersReducedMotion();

  return (
    <section className="w-full bg-gradient-to-b from-pink-50 to-white py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-12 lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-teal-600">
            会员计划
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            选一个你最舒服的节奏，越早开始越划算
          </h2>
          <p className="max-w-[540px] text-[15px] text-slate-500">
            还有季度会员等更多选项，在会员页查看完整对比。
          </p>
        </motion.div>

        <motion.div
          className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 lg:gap-6"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
        >
          {TIERS.map((t) => (
            <TierCard key={t.grade} tier={t} reduced={reduced} />
          ))}
        </motion.div>

        <motion.div
          className="flex w-full flex-col items-start justify-between gap-4 rounded-2xl bg-gradient-to-r from-violet-600 via-teal-600 to-teal-500 p-6 text-white md:flex-row md:items-center md:p-8"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <div className="flex flex-col gap-2">
            <span className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-white/80">
              <Sparkles className="h-4 w-4" /> 邀请好友，赚终身返佣
            </span>
            <p className="text-lg font-bold md:text-xl">
              永久会员邀请好友首充，你拿 30% 返佣
            </p>
          </div>
          <Link
            href="/docs/invites/referral-program"
            className="inline-flex items-center gap-2 rounded-lg bg-white/15 px-5 py-3 text-sm font-semibold text-white backdrop-blur transition-colors hover:bg-white/25"
          >
            查看规则
            <ArrowRight className="h-4 w-4" />
          </Link>
        </motion.div>
      </div>
    </section>
  );
}
