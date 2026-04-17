// dx-web/src/features/web/home/components/ai-features-section.tsx
"use client";

import { motion } from "motion/react";
import { Sparkles, Coins } from "lucide-react";
import { useEffect, useState } from "react";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";
import {
  useInView,
  usePrefersReducedMotion,
} from "@/features/web/home/hooks/use-in-view";

const BULLETS = [
  "输入任意主题或场景，AI 按你的水平生成课程",
  "CEFR A1–C2 全覆盖，难度智能匹配",
  "内容沉淀进你的词汇系统，复习自动推送",
];

const STREAM_LINES = [
  "negotiate · 谈判 — Let's negotiate the salary.",
  "résumé · 简历 — Please send me your résumé.",
  "confident · 自信 — Stay confident during the interview.",
  "leverage · 优势 — Leverage your strengths.",
  "follow up · 跟进 — I'll follow up next week.",
];

function useTypewriter(lines: string[], active: boolean): string[] {
  const [out, setOut] = useState<string[]>([]);
  const reduced = usePrefersReducedMotion();

  useEffect(() => {
    let cancelled = false;

    if (!active) return;

    if (reduced) {
      window.setTimeout(() => {
        if (!cancelled) setOut(lines);
      }, 0);
      return () => {
        cancelled = true;
      };
    }

    let lineIndex = 0;
    let charIndex = 0;

    function tick() {
      if (cancelled) return;
      if (lineIndex >= lines.length) {
        // hold then restart
        window.setTimeout(() => {
          if (cancelled) return;
          lineIndex = 0;
          charIndex = 0;
          setOut([]);
          tick();
        }, 2000);
        return;
      }
      charIndex += 1;
      const curLine = lineIndex;
      const curChar = charIndex;
      const curText = lines[curLine].slice(0, curChar);
      setOut((prev) => {
        const copy = [...prev];
        copy[curLine] = curText;
        return copy;
      });
      if (charIndex >= lines[lineIndex].length) {
        lineIndex += 1;
        charIndex = 0;
        window.setTimeout(tick, 300);
      } else {
        window.setTimeout(tick, 30);
      }
    }
    window.setTimeout(tick, 0);
    return () => {
      cancelled = true;
      setOut([]);
    };
  }, [active, reduced, lines]);

  return out;
}

export function AiFeaturesSection() {
  const { ref, inView } = useInView<HTMLDivElement>({ threshold: 0.3 });
  const stream = useTypewriter(STREAM_LINES, inView);

  return (
    <section className="w-full bg-gradient-to-b from-white to-teal-50 py-[80px] md:py-[100px]">
      <div
        ref={ref}
        className="mx-auto grid w-full max-w-[1280px] grid-cols-1 items-center gap-10 px-5 md:px-10 lg:grid-cols-2 lg:gap-16 lg:px-[120px]"
      >
        <motion.div
          className="flex flex-col gap-6"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <motion.span
            variants={staggerChildVariants}
            className="text-sm font-semibold tracking-wide text-teal-600"
          >
            AI 驱动 · 专属于你的学习
          </motion.span>
          <motion.h2
            variants={staggerChildVariants}
            className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl"
          >
            AI 帮你定制课程，你想学什么都可以
          </motion.h2>
          <motion.ul
            variants={staggerChildVariants}
            className="flex flex-col gap-3"
          >
            {BULLETS.map((b) => (
              <li key={b} className="flex items-start gap-3 text-slate-600">
                <Sparkles className="mt-1 h-4 w-4 shrink-0 text-teal-600" />
                <span className="text-[15px] leading-relaxed">{b}</span>
              </li>
            ))}
          </motion.ul>
        </motion.div>

        <motion.div
          className="rounded-2xl border border-slate-200 bg-white p-6 shadow-[0_8px_32px_rgba(13,148,136,0.08)] md:p-8"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.3 }}
        >
          <div className="mb-3 flex items-center gap-2">
            <span className="rounded-full bg-teal-50 px-3 py-1 text-xs font-semibold text-teal-700">
              AI 随心学
            </span>
            <motion.span
              initial={{ scale: 0.8 }}
              whileInView={{ scale: 1 }}
              viewport={{ once: true }}
              transition={{ type: "spring", stiffness: 300, damping: 18 }}
              className="rounded-full bg-violet-50 px-3 py-1 text-xs font-semibold text-violet-700"
            >
              CEFR B1–B2
            </motion.span>
          </div>
          <div className="mb-4 rounded-lg bg-slate-50 px-4 py-3 text-sm text-slate-600">
            输入主题：<span className="font-semibold text-slate-900">职场面试高频词</span>
          </div>
          <div className="mb-4 flex flex-col gap-2">
            {STREAM_LINES.map((full, i) => {
              const current = stream[i] ?? "";
              const isTyping = current.length < full.length;
              const isCaretRow =
                isTyping &&
                STREAM_LINES.slice(0, i).every(
                  (_, j) => (stream[j]?.length ?? 0) === STREAM_LINES[j].length,
                );
              return (
                <div
                  key={i}
                  className="text-sm leading-relaxed text-slate-700"
                >
                  <span className="mr-2 text-teal-600">›</span>
                  {current || "\u00A0"}
                  {isCaretRow && (
                    <span className="ml-0.5 inline-block h-4 w-[2px] animate-pulse bg-slate-400 align-middle" />
                  )}
                </div>
              );
            })}
          </div>
          <div className="flex items-center gap-2 border-t border-slate-100 pt-3 text-xs text-slate-500">
            <Coins className="h-3.5 w-3.5 text-amber-500" />
            <span>消耗 5 能量豆 · 失败全额退还</span>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
