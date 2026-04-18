# Hero Button Rocket Animation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the static `Gamepad2` icon on the 开启斗学之旅 hero button with an animated `Rocket` icon that floats identically to the 开始你的斗学冒险 button in the final CTA section.

**Architecture:** Single file edit — convert `hero-section.tsx` from a server component to a client component, swap the icon, and wrap it in a `motion.span` with the existing float animation pattern already used in `final-cta-section.tsx`.

**Tech Stack:** Next.js 16, TailwindCSS v4, Framer Motion (`motion/react`), Lucide React

---

### Task 1: Add rocket animation to hero button

**Files:**
- Modify: `dx-web/src/features/web/home/components/hero-section.tsx`

- [ ] **Step 1: Write the complete new file content**

  Replace the entire content of `dx-web/src/features/web/home/components/hero-section.tsx` with:

  ```tsx
  // dx-web/src/features/web/home/components/hero-section.tsx
  "use client";

  import Link from "next/link";
  import { Rocket, ArrowRight } from "lucide-react";
  import { motion } from "motion/react";
  import { HeroGameDemo } from "@/features/web/home/components/hero-game-demo";
  import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

  interface HeroSectionProps {
    isLoggedIn: boolean;
  }

  export function HeroSection({ isLoggedIn }: HeroSectionProps) {
    const reduced = usePrefersReducedMotion();
    const primaryHref = isLoggedIn ? "/hall" : "/auth/signup";

    return (
      <section className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-8 px-5 pb-[80px] pt-16 md:px-10 md:pb-[100px] md:pt-20 lg:px-[120px]">
        <div className="flex items-center gap-2 rounded-full border border-slate-200 bg-white/70 px-5 py-2 backdrop-blur">
          <div className="h-2 w-2 rounded-full bg-green-400" />
          <span className="text-[13px] font-medium text-slate-500">
            斗学，让你玩着玩着就学会英语
          </span>
        </div>
        <div className="flex w-full flex-col items-center">
          <h1 className="text-center text-5xl font-extrabold leading-tight tracking-[-2px] text-slate-900 md:text-6xl lg:text-[72px] lg:tracking-[-3px]">
            别「学」英语了
          </h1>
          <h1 className="bg-gradient-to-r from-teal-400 to-violet-500 bg-clip-text text-center text-5xl font-extrabold leading-tight tracking-[-2px] text-transparent md:text-6xl lg:text-[72px] lg:tracking-[-3px]">
            玩着玩着就会了
          </h1>
        </div>
        <p className="max-w-[680px] text-center text-sm leading-[1.6] text-slate-500 md:text-base lg:text-lg">
          多种游戏模式 · AI 定制内容 · 和朋友一起闯关
          <br className="hidden md:block" />
          每天 10 分钟，英语悄悄就流利了
        </p>
        <p className="sr-only">
          下方演示了连词成句、词汇配对、词汇消消乐和词汇对轰四种玩法的循环动画。
        </p>
        <div className="flex flex-col items-center gap-4 md:flex-row">
          <Link
            href={primaryHref}
            className="flex items-center gap-2.5 rounded-xl bg-teal-600 px-9 py-4 shadow-[0_4px_30px_rgba(13,148,136,0.27)] transition-colors hover:bg-teal-700"
          >
            <motion.span
              animate={reduced ? undefined : { y: [0, -2, 0] }}
              transition={reduced ? undefined : { duration: 1.6, repeat: Infinity, ease: "easeInOut" }}
              className="flex items-center"
            >
              <Rocket className="h-5 w-5 text-white" />
            </motion.span>
            <span className="text-base font-semibold text-white">开启斗学之旅</span>
          </Link>
          <Link
            href="#features"
            className="flex items-center gap-2.5 rounded-xl border-[1.5px] border-slate-200 bg-white/70 px-9 py-4 transition-colors hover:bg-white"
          >
            <span className="text-base font-medium text-slate-900">了解更多</span>
            <ArrowRight className="h-[18px] w-[18px] text-slate-900" />
          </Link>
        </div>
        <div className="relative mt-4 w-full max-w-[900px]">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-[-40px] -z-10 rounded-[32px] bg-[radial-gradient(ellipse_at_center,rgba(94,234,212,0.35),transparent_70%)]"
          />
          <HeroGameDemo />
        </div>
      </section>
    );
  }
  ```

  Changes from the original:
  - Added `"use client"` directive at the top.
  - Replaced `Gamepad2` import with `Rocket`; added `motion` from `motion/react` and `usePrefersReducedMotion` from `@/features/web/home/hooks/use-in-view`.
  - Added `const reduced = usePrefersReducedMotion()` inside the component.
  - Replaced `<Gamepad2 className="h-5 w-5 text-white" />` with `<motion.span animate … ><Rocket … /></motion.span>`.
  - Everything else (layout, text, hrefs, other button, HeroGameDemo) is unchanged.

- [ ] **Step 2: Run lint**

  ```bash
  cd dx-web && npm run lint
  ```

  Expected: no errors.

- [ ] **Step 3: Commit**

  ```bash
  git add "dx-web/src/features/web/home/components/hero-section.tsx"
  git commit -m "fix(web): replace hero button Gamepad2 icon with animated Rocket"
  ```

- [ ] **Step 4: Build verification**

  ```bash
  cd dx-web && npm run build
  ```

  Expected: exits 0, no TypeScript or build errors.

- [ ] **Step 5: Visual check**

  ```bash
  cd dx-web && npm run dev
  ```

  Open `http://localhost:3000` and verify:
  1. The 开启斗学之旅 button shows a rocket icon (not a gamepad).
  2. The rocket gently floats up and down — the same 2px vertical loop as the 开始你的斗学冒险 button further down the page.
  3. All other elements on the hero section (heading, subtitle, 了解更多 button, game demo) are unchanged.
  4. If OS reduced-motion is enabled, the rocket is static (no animation).
