# Home Landing Page Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the mocked home landing at `/` with a truthful, motion-rich redesign using the existing teal theme — install `motion`, rebuild 9 sections with real copy/data, drop fake stats + fake testimonials, add live hero demo + membership/referral + FAQ, verify accessibility and responsive behavior.

**Architecture:** All work lives under `dx-web/src/features/web/home/` (plus a `motion` install in `dx-web/package.json` and a page composition edit in `dx-web/src/app/(web)/(home)/page.tsx`). Shared motion presets + `useInView` + `useRotatingScene` hooks are defined once and consumed by sections. Hero demo is client-only and dynamically imported. Every section's data is traceable to real product features, docs, or `/purchase/membership` pricing.

**Tech Stack:** Next.js 16 (app router), React 19, Tailwind CSS v4, shadcn/ui primitives (`accordion`, `tabs`), `motion` npm package (successor to framer-motion), TypeScript. No backend (`dx-api/`) changes.

**Spec reference:** `docs/superpowers/specs/2026-04-17-home-landing-redesign-design.md`

**Testing conventions:** `dx-web` has no unit-test framework. Verification is:
- `npx tsc --noEmit` — type check (must be clean)
- `npm run lint` — ESLint (zero warnings, zero errors)
- `npm run build` — Next build (at milestone tasks only)
- Manual browser check on `npm run dev` (http://localhost:3000) at milestone tasks

Every task ends with a commit.

---

## File Structure

### New files (under `dx-web/src/features/web/home/`)

| File | Responsibility |
|---|---|
| `helpers/motion-presets.ts` | Shared reveal/stagger/spring config |
| `hooks/use-in-view.ts` | IntersectionObserver wrapper + `prefers-reduced-motion` |
| `hooks/use-rotating-scene.ts` | 6s scene rotation with visibility/reduced-motion guards |
| `components/hero-game-demo.tsx` | Client-only frame + scene switcher, dynamic import |
| `components/hero-game-demo-word-sentence.tsx` | Scene A — 连词成句 animated loop |
| `components/hero-game-demo-vocab-battle.tsx` | Scene B — 词汇对轰 animated loop |
| `components/why-different-section.tsx` | "为什么是斗学" contrast band |
| `components/game-mode-preview.tsx` | 4 mini animated preview variants per mode |
| `components/learning-loop-section.tsx` | Merged course progression + vocab manager |
| `components/community-section.tsx` | Community cards incl. new streak card |
| `components/membership-section.tsx` | Tier cards + referral band |
| `components/faq-section.tsx` | Accordion with real doc links |

### Modified files

| File | Change |
|---|---|
| `dx-web/package.json` + lockfile | Add `motion` dependency |
| `dx-web/src/features/web/home/components/hero-section.tsx` | Copy polish + embed `<HeroGameDemo/>` |
| `dx-web/src/features/web/home/components/features-section.tsx` | Add per-card mini `<GameModePreview/>` strip + reveal stagger |
| `dx-web/src/features/web/home/components/ai-features-section.tsx` | Keep **only** AI 随心学; replace content with typewriter mock |
| `dx-web/src/features/web/home/components/final-cta-section.tsx` | Tighten copy; add rocket idle float + secondary doc CTA |
| `dx-web/src/app/(web)/(home)/page.tsx` | New composition order; drop deleted sections; slim top gradient |

### Deleted files

| File | Reason |
|---|---|
| `dx-web/src/features/web/home/components/stats-section.tsx` | Fake 500K/10M/92%/4.9 — replaced by `why-different-section` |
| `dx-web/src/features/web/home/components/testimonials-section.tsx` | 9 fabricated quotes — replaced by `faq-section` |
| `dx-web/src/features/web/home/components/smart-vocabulary-section.tsx` | Merged into `learning-loop-section` |
| `dx-web/src/features/web/home/components/course-platform-section.tsx` | Merged into `learning-loop-section` |
| `dx-web/src/features/web/home/components/social-community-section.tsx` | Renamed/refactored into `community-section` |

No changes in `dx-api/`, `deploy/`, or `dx-web/src/components/ui/` (shadcn-managed).

---

## Task 0: Baseline verification

**Files:** none (verification only)

- [ ] **Step 1: Confirm branch and clean tree**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git status
git log -1 --oneline
```

Expected: on `main`, clean working tree, HEAD at the design-spec commit (`d0fce40 docs: home landing redesign design spec`) or newer as long as the tree is clean.

- [ ] **Step 2: Baseline tsc + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass with no errors.

---

## Task 1: Install `motion`

**Files:**
- Modify: `dx-web/package.json`, `dx-web/package-lock.json`

- [ ] **Step 1: Install the dependency**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm i motion
```

- [ ] **Step 2: Confirm only `motion` was added**

```bash
git diff --stat package.json package-lock.json
```

Expected: `package.json` gains `"motion"` under `dependencies`; lockfile updated. No other unexpected edits.

- [ ] **Step 3: Sanity import check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
```

Expected: passes (motion ships its own types).

- [ ] **Step 4: Commit**

```bash
git add package.json package-lock.json
git commit -m "chore(web): add motion dependency for home landing"
```

---

## Task 2: Shared motion presets + `useInView` + `useRotatingScene`

**Files:**
- Create: `dx-web/src/features/web/home/helpers/motion-presets.ts`
- Create: `dx-web/src/features/web/home/hooks/use-in-view.ts`
- Create: `dx-web/src/features/web/home/hooks/use-rotating-scene.ts`

- [ ] **Step 1: Write `motion-presets.ts`**

```ts
// dx-web/src/features/web/home/helpers/motion-presets.ts
import type { Variants } from "motion/react";

export const revealEase = { duration: 0.45, ease: [0.22, 1, 0.36, 1] } as const;
export const revealSpring = {
  type: "spring",
  stiffness: 160,
  damping: 24,
  mass: 0.6,
} as const;

export const revealVariants: Variants = {
  hidden: { opacity: 0, y: 24 },
  show: { opacity: 1, y: 0, transition: revealEase },
};

export const staggerContainerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

export const staggerChildVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0, transition: revealEase },
};
```

- [ ] **Step 2: Write `use-in-view.ts`**

```ts
// dx-web/src/features/web/home/hooks/use-in-view.ts
"use client";

import { useEffect, useRef, useState } from "react";

export function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    const apply = () => setReduced(mq.matches);
    apply();
    mq.addEventListener("change", apply);
    return () => mq.removeEventListener("change", apply);
  }, []);

  return reduced;
}

export function useInView<T extends Element>(
  options: IntersectionObserverInit = { threshold: 0.2 },
): { ref: React.RefObject<T | null>; inView: boolean } {
  const ref = useRef<T | null>(null);
  const [inView, setInView] = useState(false);

  useEffect(() => {
    const node = ref.current;
    if (!node) return;
    const obs = new IntersectionObserver(([entry]) => {
      setInView(entry.isIntersecting);
    }, options);
    obs.observe(node);
    return () => obs.disconnect();
    // options is a stable object literal at call sites; re-running on identity change would be churn.
     
  }, []);

  return { ref, inView };
}
```

- [ ] **Step 3: Write `use-rotating-scene.ts`**

```ts
// dx-web/src/features/web/home/hooks/use-rotating-scene.ts
"use client";

import { useEffect, useState } from "react";

interface Options {
  total: number;
  intervalMs?: number;
  paused?: boolean;
}

/** Cycles an index [0..total) while not paused. */
export function useRotatingScene({
  total,
  intervalMs = 6000,
  paused = false,
}: Options): { index: number; setIndex: (i: number) => void } {
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (paused || total <= 1) return;
    const id = window.setInterval(() => {
      setIndex((i) => (i + 1) % total);
    }, intervalMs);
    return () => window.clearInterval(id);
  }, [paused, total, intervalMs]);

  return { index, setIndex };
}
```

- [ ] **Step 4: Type check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
```

Expected: passes.

- [ ] **Step 5: Commit**

```bash
git add src/features/web/home/helpers/motion-presets.ts \
        src/features/web/home/hooks/use-in-view.ts \
        src/features/web/home/hooks/use-rotating-scene.ts
git commit -m "feat(web): add motion presets + in-view & rotating-scene hooks"
```

---

## Task 3: Refresh `hero-section.tsx` copy + gradient halo (demo wiring deferred)

**Files:**
- Modify: `dx-web/src/features/web/home/components/hero-section.tsx`

- [ ] **Step 1: Replace file contents**

```tsx
// dx-web/src/features/web/home/components/hero-section.tsx
import Link from "next/link";
import { Gamepad2, ArrowRight } from "lucide-react";
import { HeroGameDemo } from "@/features/web/home/components/hero-game-demo";

interface HeroSectionProps {
  isLoggedIn: boolean;
}

export function HeroSection({ isLoggedIn }: HeroSectionProps) {
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
        下方演示了连词成句和词汇对轰两种玩法的循环动画。
      </p>
      <div className="flex flex-col items-center gap-4 md:flex-row">
        <Link
          href={primaryHref}
          className="flex items-center gap-2.5 rounded-xl bg-teal-600 px-9 py-4 shadow-[0_4px_30px_rgba(13,148,136,0.27)] transition-colors hover:bg-teal-700"
        >
          <Gamepad2 className="h-5 w-5 text-white" />
          <span className="text-base font-semibold text-white">开始斗学之旅</span>
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

- [ ] **Step 2: Type check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
```

Expected: **fails** with "Cannot find module `@/features/web/home/components/hero-game-demo`" — that's OK, we create it in Task 4. Leave the edit in place and continue.

- [ ] **Step 3: Do not commit yet** — commit happens at the end of Task 4 together with the demo component.

---

## Task 4: Build `hero-game-demo.tsx` container + dynamic-import wiring

**Files:**
- Create: `dx-web/src/features/web/home/components/hero-game-demo.tsx`
- Create (placeholder scenes referenced from this file; filled in Tasks 5 & 6): none yet

- [ ] **Step 1: Write the container**

```tsx
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

const SceneVocabBattle = dynamic(
  () =>
    import("@/features/web/home/components/hero-game-demo-vocab-battle").then(
      (m) => m.HeroGameDemoVocabBattle,
    ),
  { ssr: false },
);

const SCENES = [
  { key: "word-sentence", label: "连词成句", Scene: SceneWordSentence },
  { key: "vocab-battle", label: "词汇对轰", Scene: SceneVocabBattle },
] as const;

export function HeroGameDemo() {
  const { ref, inView } = useInView<HTMLDivElement>({ threshold: 0.2 });
  const [hovered, setHovered] = useState(false);
  const paused = !inView || hovered;
  const { index, setIndex } = useRotatingScene({
    total: SCENES.length,
    intervalMs: 6000,
    paused,
  });

  const ActiveScene = SCENES[index].Scene;

  return (
    <div
      ref={ref}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      className="relative aspect-[5/3] w-full overflow-hidden rounded-[20px] border border-slate-200 bg-white/90 p-4 shadow-[0_12px_40px_rgba(13,148,136,0.12)] backdrop-blur md:aspect-[15/7] md:p-6"
      aria-hidden="true"
    >
      <div className="absolute left-4 top-4 z-10 flex gap-1 rounded-full border border-slate-200 bg-white p-1 text-xs shadow-sm md:left-6 md:top-6">
        {SCENES.map((s, i) => (
          <button
            key={s.key}
            type="button"
            onClick={() => setIndex(i)}
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
      <div className="flex h-full w-full items-center justify-center">
        <ActiveScene key={SCENES[index].key} active={inView} />
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Write scene-type stubs so imports resolve**

Create two stub files to keep tsc happy; Tasks 5 and 6 replace their contents.

```tsx
// dx-web/src/features/web/home/components/hero-game-demo-word-sentence.tsx
"use client";

interface Props {
  active: boolean;
}

export function HeroGameDemoWordSentence({ active: _active }: Props) {
  return <div className="h-full w-full" />;
}
```

```tsx
// dx-web/src/features/web/home/components/hero-game-demo-vocab-battle.tsx
"use client";

interface Props {
  active: boolean;
}

export function HeroGameDemoVocabBattle({ active: _active }: Props) {
  return <div className="h-full w-full" />;
}
```

- [ ] **Step 3: Update `page.tsx` to pass `isLoggedIn` to `HeroSection`**

Open `dx-web/src/app/(web)/(home)/page.tsx` and change the hero usage:

```tsx
// before
<HeroSection />
// after
<HeroSection isLoggedIn={isLoggedIn} />
```

Do not change anything else in `page.tsx` yet (Task 17 rewrites the full composition).

- [ ] **Step 4: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass. If lint complains about unused `_active`, the leading underscore exempts it under `@typescript-eslint/no-unused-vars`; otherwise change signature to `(_props: Props)`.

- [ ] **Step 5: Browser check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run dev
```

Open `http://localhost:3000/`. Expected: hero renders with a large teal-glowing empty card. The tab pill shows `连词成句 | 词汇对轰`. Clicking the pill switches active tab styling. Page should not crash. Kill the dev server with `Ctrl+C` after confirming.

- [ ] **Step 6: Commit**

```bash
git add src/features/web/home/components/hero-section.tsx \
        src/features/web/home/components/hero-game-demo.tsx \
        src/features/web/home/components/hero-game-demo-word-sentence.tsx \
        src/features/web/home/components/hero-game-demo-vocab-battle.tsx \
        src/app/\(web\)/\(home\)/page.tsx
git commit -m "feat(web/home): hero copy polish + HeroGameDemo container + scene stubs"
```

---

## Task 5: Scene A — 连词成句 animated loop

**Files:**
- Modify: `dx-web/src/features/web/home/components/hero-game-demo-word-sentence.tsx`

- [ ] **Step 1: Replace file contents**

```tsx
// dx-web/src/features/web/home/components/hero-game-demo-word-sentence.tsx
"use client";

import { motion } from "motion/react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface Props {
  active: boolean;
}

const WORDS = [
  { text: "I", className: "bg-teal-100 text-teal-700" },
  { text: "eat", className: "bg-violet-100 text-violet-700" },
  { text: "fresh", className: "bg-pink-100 text-pink-700" },
  { text: "apples", className: "bg-amber-100 text-amber-700" },
  { text: "every", className: "bg-rose-100 text-rose-700" },
  { text: "morning", className: "bg-blue-100 text-blue-700" },
] as const;

export function HeroGameDemoWordSentence({ active }: Props) {
  const reduced = usePrefersReducedMotion();
  if (!active) return <div className="h-full w-full" />;

  return (
    <motion.div
      className="flex w-full max-w-[640px] flex-col gap-4"
      initial={reduced ? false : { opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
    >
      <p className="text-xs text-slate-500">请把下面的单词拼成正确的句子：</p>
      <div className="flex flex-wrap gap-2">
        {WORDS.map((w, i) => (
          <motion.span
            key={w.text}
            className={`rounded-lg px-3 py-2 text-sm font-semibold shadow-sm ${w.className}`}
            initial={reduced ? false : { opacity: 0, y: -12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={
              reduced ? { duration: 0 } : { delay: 0.12 * i, duration: 0.35, ease: [0.22, 1, 0.36, 1] }
            }
          >
            {w.text}
          </motion.span>
        ))}
      </div>
      <div className="relative h-1 w-full overflow-hidden rounded-full bg-slate-100">
        <motion.div
          className="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-teal-400 to-violet-500"
          initial={reduced ? { width: "100%" } : { width: 0 }}
          animate={{ width: "100%" }}
          transition={reduced ? { duration: 0 } : { delay: 1.4, duration: 1.0, ease: "easeOut" }}
        />
      </div>
      <motion.div
        className="flex items-center gap-3"
        initial={reduced ? false : { opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={reduced ? { duration: 0 } : { delay: 2.2, duration: 0.4 }}
      >
        <span className="rounded-full bg-teal-600 px-3 py-1 text-xs font-semibold text-white">
          +128 ★
        </span>
        <span className="rounded-full bg-violet-100 px-3 py-1 text-xs font-semibold text-violet-700">
          combo x2
        </span>
        <span className="text-xs text-slate-500">句子正确！</span>
      </motion.div>
    </motion.div>
  );
}
```

- [ ] **Step 2: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Browser check**

Run `npm run dev`, open `/`, confirm: word chips drop in sequentially, progress bar fills teal→violet, score badge pops. Switching to 词汇对轰 tab still shows the empty stub. Kill dev server.

- [ ] **Step 4: Commit**

```bash
git add src/features/web/home/components/hero-game-demo-word-sentence.tsx
git commit -m "feat(web/home): implement Scene A 连词成句 demo"
```

---

## Task 6: Scene B — 词汇对轰 animated loop

**Files:**
- Modify: `dx-web/src/features/web/home/components/hero-game-demo-vocab-battle.tsx`

- [ ] **Step 1: Replace file contents**

```tsx
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
            initial={{ width: "100%" }}
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
```

- [ ] **Step 2: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Browser check**

Run `npm run dev`, open `/`, click the `词汇对轰` pill. Confirm: Chinese prompt slides in, letters type out, opponent HP bar drops from full to 55%, projectile sweep occurs, combo badge pops. Kill dev server.

- [ ] **Step 4: Commit**

```bash
git add src/features/web/home/components/hero-game-demo-vocab-battle.tsx
git commit -m "feat(web/home): implement Scene B 词汇对轰 demo"
```

---

## Task 7: Why-different contrast section

**Files:**
- Create: `dx-web/src/features/web/home/components/why-different-section.tsx`

- [ ] **Step 1: Write the section**

```tsx
// dx-web/src/features/web/home/components/why-different-section.tsx
"use client";

import { motion } from "motion/react";
import { ArrowRight } from "lucide-react";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";

const ROWS = [
  { before: "背了就忘，靠意志力硬撑", after: "游戏化循环，大脑自发想再玩一局" },
  { before: "学的和用的两张皮", after: "连词成句、对话、对战都是真实语料" },
  { before: "一个人孤独地学", after: "好友开黑 · 学习群 · 排行榜 · 每日连胜" },
] as const;

export function WhyDifferentSection() {
  return (
    <section className="w-full bg-gradient-to-b from-white to-slate-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-[60px] lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-teal-600">
            为什么是斗学
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            不再死记硬背，让大脑爱上英语
          </h2>
        </motion.div>

        <motion.ul
          className="flex w-full max-w-[880px] flex-col gap-4"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
        >
          {ROWS.map((row) => (
            <motion.li
              key={row.before}
              variants={staggerChildVariants}
              className="flex flex-col items-stretch gap-3 rounded-2xl border border-slate-200 bg-white p-5 md:flex-row md:items-center md:gap-6 md:p-6"
            >
              <div className="flex-1 text-[15px] text-slate-400 line-through decoration-slate-300">
                {row.before}
              </div>
              <ArrowRight className="hidden h-5 w-5 shrink-0 text-slate-300 md:block" />
              <div className="flex-1 rounded-xl bg-gradient-to-r from-teal-50 to-violet-50 px-4 py-3 text-[15px] font-medium text-slate-900">
                {row.after}
              </div>
            </motion.li>
          ))}
        </motion.ul>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/why-different-section.tsx
git commit -m "feat(web/home): add why-different contrast section"
```

---

## Task 8: Mini game-mode previews component

**Files:**
- Create: `dx-web/src/features/web/home/components/game-mode-preview.tsx`

- [ ] **Step 1: Write the component**

```tsx
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

function VocabElimPreview({ animate }: { animate: boolean }) {
  const cells = Array.from({ length: 9 });
  return (
    <Strip>
      <div className="grid flex-1 grid-cols-9 gap-1">
        {cells.map((_, i) => (
          <motion.div
            key={i}
            className="h-6 rounded-md bg-white shadow-sm"
            animate={
              animate
                ? {
                    backgroundColor: [
                      "#ffffff",
                      i % 2 === 0 ? "#fce7f3" : "#ccfbf1",
                      "#ffffff",
                    ],
                    scale: [1, 0.9, 1],
                  }
                : undefined
            }
            transition={
              animate
                ? { duration: 1.2, delay: 0.06 * i, repeat: Infinity, repeatDelay: 0.6 }
                : undefined
            }
          />
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
      <span className="text-[11px] font-semibold text-rose-600">AI</span>
    </Strip>
  );
}
```

- [ ] **Step 2: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/game-mode-preview.tsx
git commit -m "feat(web/home): add GameModePreview mini loops"
```

---

## Task 9: Refresh `features-section.tsx` (4 game modes with previews)

**Files:**
- Modify: `dx-web/src/features/web/home/components/features-section.tsx`

- [ ] **Step 1: Replace file contents**

```tsx
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
```

- [ ] **Step 2: Type check + lint + browser check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass. Run `npm run dev`, open `/`, scroll to the features section, hover each card — preview plays on hover desktop; on a narrow viewport the preview autoplays.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/features-section.tsx
git commit -m "feat(web/home): refresh game-modes section with mini previews and reveal stagger"
```

---

## Task 10: Refresh `ai-features-section.tsx` — 随心学 only

**Files:**
- Modify: `dx-web/src/features/web/home/components/ai-features-section.tsx`

- [ ] **Step 1: Replace file contents**

```tsx
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
  const [out, setOut] = useState<string[]>(active ? lines : []);
  const reduced = usePrefersReducedMotion();

  useEffect(() => {
    if (!active) {
      setOut([]);
      return;
    }
    if (reduced) {
      setOut(lines);
      return;
    }
    setOut([]);
    let cancelled = false;
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
      setOut((prev) => {
        const copy = [...prev];
        copy[lineIndex] = lines[lineIndex].slice(0, charIndex);
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
    tick();
    return () => {
      cancelled = true;
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
            {stream.map((line, i) => (
              <div
                key={i}
                className="text-sm leading-relaxed text-slate-700"
              >
                <span className="mr-2 text-teal-600">›</span>
                {line}
                {i === stream.length - 1 && stream[i]?.length < STREAM_LINES[i]?.length && (
                  <span className="ml-0.5 inline-block h-4 w-[2px] animate-pulse bg-slate-400 align-middle" />
                )}
              </div>
            ))}
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
```

- [ ] **Step 2: Type check + lint + browser check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass. Run dev, scroll to AI section — streaming typewriter should roll, CEFR pill pops, 能量豆 note visible.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/ai-features-section.tsx
git commit -m "feat(web/home): rebuild AI section as 随心学-only with streaming mock"
```

---

## Task 11: Build `learning-loop-section.tsx` (merged)

**Files:**
- Create: `dx-web/src/features/web/home/components/learning-loop-section.tsx`

- [ ] **Step 1: Write the section**

```tsx
// dx-web/src/features/web/home/components/learning-loop-section.tsx
"use client";

import { motion } from "motion/react";
import { ArrowRight, BookOpen, RefreshCw, CircleCheck, Check } from "lucide-react";
import {
  revealVariants,
  staggerChildVariants,
  staggerContainerVariants,
} from "@/features/web/home/helpers/motion-presets";

const LEVELS = [
  { label: "Lv.01", done: true },
  { label: "Lv.02", done: true },
  { label: "Lv.03", done: false },
];

const VOCAB = [
  {
    title: "生词本",
    desc: "自动沉淀你不会的词",
    Icon: BookOpen,
    iconColor: "#EC4899",
    bg: "bg-pink-500/10",
  },
  {
    title: "复习本",
    desc: "[1, 3, 7, 14, 30, 90] 天节奏智能推送",
    Icon: RefreshCw,
    iconColor: "#7B61FF",
    bg: "bg-violet-500/10",
  },
  {
    title: "已掌握",
    desc: "看得见的词汇量增长",
    Icon: CircleCheck,
    iconColor: "#0d9488",
    bg: "bg-teal-500/10",
  },
];

export function LearningLoopSection() {
  return (
    <section className="w-full bg-gradient-to-b from-teal-50 to-violet-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-[60px] lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-violet-500">
            学习闭环 · 从陌生到掌握
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            一套系统，追踪你每一个单词的命运
          </h2>
        </motion.div>

        <motion.div
          className="flex w-full flex-col items-stretch gap-6 lg:flex-row lg:items-center lg:justify-between"
          variants={staggerContainerVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
        >
          <motion.div
            variants={staggerChildVariants}
            className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-6 lg:w-[280px]"
          >
            <span className="text-xs font-semibold uppercase tracking-wider text-violet-600">
              闯关式课程
            </span>
            {LEVELS.map((lv) => (
              <div
                key={lv.label}
                className={`flex items-center justify-between rounded-lg border px-4 py-3 ${
                  lv.done
                    ? "border-teal-200 bg-teal-50"
                    : "border-violet-200 bg-white"
                }`}
              >
                <span className="text-sm font-semibold text-slate-900">
                  {lv.label}
                </span>
                {lv.done ? (
                  <Check className="h-4 w-4 text-teal-600" />
                ) : (
                  <motion.span
                    className="h-2 w-2 rounded-full bg-violet-500"
                    animate={{ scale: [1, 1.4, 1] }}
                    transition={{ duration: 1.6, repeat: Infinity }}
                  />
                )}
              </div>
            ))}
          </motion.div>

          <motion.div
            variants={staggerChildVariants}
            className="relative flex flex-1 items-center justify-center py-4 lg:px-6"
          >
            <div className="relative h-px w-full bg-gradient-to-r from-teal-300 via-teal-500 to-violet-500">
              <motion.span
                className="absolute -top-1 h-3 w-3 rounded-full bg-teal-500 shadow-[0_0_12px_rgba(13,148,136,0.8)]"
                initial={{ left: "0%" }}
                whileInView={{ left: "100%" }}
                viewport={{ once: true, amount: 0.4 }}
                transition={{ duration: 1.4, ease: "easeInOut" }}
              />
            </div>
            <div className="absolute -bottom-6 left-1/2 -translate-x-1/2 whitespace-nowrap text-xs text-slate-500">
              每次游戏自动沉淀单词 <ArrowRight className="inline h-3 w-3" />
            </div>
          </motion.div>

          <motion.div
            variants={staggerChildVariants}
            className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-6 lg:w-[320px]"
          >
            <span className="text-xs font-semibold uppercase tracking-wider text-teal-600">
              词汇三本
            </span>
            {VOCAB.map(({ title, desc, Icon, iconColor, bg }) => (
              <div
                key={title}
                className="flex items-center gap-3 rounded-lg border border-slate-100 bg-slate-50 px-3 py-2.5"
              >
                <div
                  className={`flex h-9 w-9 items-center justify-center rounded-lg ${bg}`}
                >
                  <Icon className="h-4 w-4" style={{ color: iconColor }} />
                </div>
                <div className="flex flex-col">
                  <span className="text-sm font-semibold text-slate-900">
                    {title}
                  </span>
                  <span className="text-xs text-slate-500">{desc}</span>
                </div>
              </div>
            ))}
          </motion.div>
        </motion.div>

        <motion.p
          className="max-w-[640px] text-center text-[15px] text-slate-500"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          艾宾浩斯遗忘曲线智能推送复习，你只需要玩。
        </motion.p>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/learning-loop-section.tsx
git commit -m "feat(web/home): add learning-loop section merging course + vocab"
```

---

## Task 12: Build `community-section.tsx` (renamed + streak card)

**Files:**
- Create: `dx-web/src/features/web/home/components/community-section.tsx`

- [ ] **Step 1: Write the section**

```tsx
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

interface Card {
  Icon: LucideIcon;
  iconColor: string;
  bgClassName: string;
  title: string;
  description: string;
  extra?: React.ReactNode;
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

function StreakHeatmap() {
  const cells = Array.from({ length: 7 });
  return (
    <motion.div
      className="mt-2 flex gap-1"
      initial="hidden"
      whileInView="show"
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
            hidden: { backgroundColor: "#f1f5f9" },
            show: {
              backgroundColor: i < 6 ? "#0d9488" : "#99f6e4",
            },
          }}
          transition={{ duration: 0.3 }}
        />
      ))}
    </motion.div>
  );
}

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
```

- [ ] **Step 2: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/community-section.tsx
git commit -m "feat(web/home): add community section with streak heatmap card"
```

---

## Task 13: Build `membership-section.tsx` (tier cards + referral band)

**Files:**
- Create: `dx-web/src/features/web/home/components/membership-section.tsx`

- [ ] **Step 1: Write the section**

```tsx
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
    badge: {
      label: "推荐",
      className: "bg-teal-600 text-white",
    },
    cardClassName: "bg-white ring-2 ring-teal-500",
    ctaClassName: "bg-teal-600 text-white hover:bg-teal-700",
    pulse: true,
  },
  {
    grade: USER_GRADES.LIFETIME,
    period: "",
    features: ["一次付费，终身生效", "最高能量豆赠送", "邀请好友首充享 30% 返佣"],
    badge: {
      label: "最超值",
      className: "bg-violet-600 text-white",
    },
    cardClassName:
      "bg-gradient-to-br from-violet-600 to-teal-600 text-white ring-2 ring-violet-500",
    ctaClassName: "bg-white text-teal-700 hover:bg-white/90",
  },
];

function TierCard({ tier }: { tier: TierConfig }) {
  const price = USER_GRADE_PRICES[tier.grade];
  const name = USER_GRADE_LABELS[tier.grade];
  const isDark = tier.grade === USER_GRADES.LIFETIME;

  return (
    <motion.div
      variants={staggerChildVariants}
      className={`relative flex flex-col gap-4 rounded-2xl p-6 shadow-[0_4px_16px_rgba(15,23,42,0.05)] ${tier.cardClassName}`}
      animate={tier.pulse ? { y: [0, -4, 0] } : undefined}
      transition={
        tier.pulse
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
            <TierCard key={t.grade} tier={t} />
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
```

- [ ] **Step 2: Type check + lint + browser check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass. Dev: scroll to membership section; four cards render with real prices (¥0 / ¥39 / ¥309 / ¥1999); referral band below. Click a tier CTA — routes to `/purchase/membership`.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/membership-section.tsx
git commit -m "feat(web/home): add membership & referral section using real pricing"
```

---

## Task 14: Build `faq-section.tsx`

**Files:**
- Create: `dx-web/src/features/web/home/components/faq-section.tsx`

The 9 FAQ items reference real doc slugs from `dx-web/src/features/web/docs/registry.ts`:

| # | Question | Target doc |
|---|---|---|
| 1 | `斗学适合什么水平的学习者？` | `/docs/getting-started/what-is-douxue` |
| 2 | `能量豆是什么？怎么获得和消耗？` | `/docs/membership/beans-monthly` |
| 3 | `免费用户能玩到哪些内容？` | `/docs/membership/tiers-compare` |
| 4 | `会员自动续费怎么关掉？` | `/docs/membership/purchase-flow` |
| 5 | `未成年人可以用吗？需要家长同意吗？` | `/docs/account/guardian-consent` |
| 6 | `支持微信支付和支付宝吗？` | `/docs/membership/purchase-flow` |
| 7 | `我的学习数据会怎么保护？` | `/docs/account/privacy-policy` |
| 8 | `邀请返佣什么时候到账？` | `/docs/invites/referral-program` |
| 9 | `如何提交反馈或报告游戏问题？` | `/docs/account/feedback` |

Short answers (1–2 sentences) are grounded in these docs. If the engineer finds an answer contradicts the doc during review, update it before commit.

- [ ] **Step 1: Write the section**

```tsx
// dx-web/src/features/web/home/components/faq-section.tsx
"use client";

import Link from "next/link";
import { motion } from "motion/react";
import { ArrowRight } from "lucide-react";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { revealVariants } from "@/features/web/home/helpers/motion-presets";

interface Faq {
  q: string;
  a: string;
  href: string;
}

const FAQS: Faq[] = [
  {
    q: "斗学适合什么水平的学习者？",
    a: "从零基础到雅思/托福备考都可以：课程按 CEFR 难度分级，多种玩法覆盖听说读写，每天 10 分钟就能坚持。",
    href: "/docs/getting-started/what-is-douxue",
  },
  {
    q: "能量豆是什么？怎么获得和消耗？",
    a: "能量豆用于使用 AI 随心学等高级功能。注册赠送、每日登录、购买、或会员月度赠送都可获得；AI 生成失败会全额退还。",
    href: "/docs/membership/beans-monthly",
  },
  {
    q: "免费用户能玩到哪些内容？",
    a: "免费用户可以玩每个课程的首关、使用基础游戏模式、加入少量学习群。付费后解锁全部关卡与 PK、小组、AI 等能力。",
    href: "/docs/membership/tiers-compare",
  },
  {
    q: "会员自动续费怎么关掉？",
    a: "在购买页或订单中心随时关闭自动续费，当前周期到期后不再扣款，已享权益保留到期。",
    href: "/docs/membership/purchase-flow",
  },
  {
    q: "未成年人可以用吗？需要家长同意吗？",
    a: "8 周岁以上可以使用；涉及付费功能需要监护人阅读并同意《监护人同意书》。我们对未成年人有专门的隐私与内容保护规则。",
    href: "/docs/account/guardian-consent",
  },
  {
    q: "支持微信支付和支付宝吗？",
    a: "支持。订单创建后会跳转到选定的支付方式，30 分钟内未支付订单自动失效。",
    href: "/docs/membership/purchase-flow",
  },
  {
    q: "我的学习数据会怎么保护？",
    a: "数据存储在中国境内，采用传输与存储加密。你可以随时查询、更正、删除或注销账号。完整条款见《隐私政策》。",
    href: "/docs/account/privacy-policy",
  },
  {
    q: "邀请返佣什么时候到账？",
    a: "永久会员邀请的新用户首次付费后，按 30% 比例计入你的返佣余额；完成攻略期与结算后按规则发放。",
    href: "/docs/invites/referral-program",
  },
  {
    q: "如何提交反馈或报告游戏问题？",
    a: "在「我的 → 提交反馈」页提交，支持多种类型；游戏内关卡问题可以直接从结算页上报。",
    href: "/docs/account/feedback",
  },
];

export function FaqSection() {
  return (
    <section className="w-full bg-gradient-to-b from-white to-slate-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-12 lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-violet-500">
            常见问题
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            开始之前，你可能想问
          </h2>
        </motion.div>

        <motion.div
          className="w-full max-w-[880px]"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.15 }}
        >
          <Accordion type="single" collapsible defaultValue="faq-0">
            {FAQS.map((f, i) => (
              <AccordionItem key={f.q} value={`faq-${i}`}>
                <AccordionTrigger className="text-left text-[15px] font-semibold text-slate-900">
                  {f.q}
                </AccordionTrigger>
                <AccordionContent className="flex flex-col gap-3 text-sm leading-relaxed text-slate-600">
                  <p>{f.a}</p>
                  <Link
                    href={f.href}
                    className="inline-flex items-center gap-1 text-sm font-semibold text-teal-600 hover:text-teal-700"
                  >
                    查看完整说明
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </motion.div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Type check + lint + browser check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass. Dev: FAQ accordion expands/collapses; first item pre-open; every `查看完整说明` links to a real doc route that renders (manually spot-check at least 2-3).

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/faq-section.tsx
git commit -m "feat(web/home): add FAQ section linking to real docs"
```

---

## Task 15: Tighten `final-cta-section.tsx`

**Files:**
- Modify: `dx-web/src/features/web/home/components/final-cta-section.tsx`

- [ ] **Step 1: Replace file contents**

```tsx
// dx-web/src/features/web/home/components/final-cta-section.tsx
"use client";

import Link from "next/link";
import { motion } from "motion/react";
import { Rocket, ArrowRight } from "lucide-react";
import { usePrefersReducedMotion } from "@/features/web/home/hooks/use-in-view";

interface FinalCtaSectionProps {
  isLoggedIn: boolean;
}

export function FinalCtaSection({ isLoggedIn }: FinalCtaSectionProps) {
  const reduced = usePrefersReducedMotion();
  const primaryHref = isLoggedIn ? "/hall" : "/auth/signup";

  return (
    <section className="w-full bg-gradient-to-b from-slate-50 to-teal-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-8 px-5 md:px-10 lg:px-[120px]">
        <div className="h-1 w-full max-w-[600px] rounded-full bg-gradient-to-r from-teal-400/0 via-teal-400 via-30% to-violet-500/0" />
        <motion.h2
          className="text-center text-4xl font-extrabold tracking-[-1px] text-slate-900 md:text-5xl lg:text-[52px] lg:tracking-[-2px]"
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.4 }}
          transition={{ duration: 0.5 }}
        >
          准备好把英语打通关了吗？
        </motion.h2>
        <p className="max-w-[600px] text-center text-base leading-[1.6] text-slate-500 md:text-lg">
          30 秒注册，今晚就能陪你玩到手滑。
        </p>
        <div className="flex flex-col items-center gap-4 md:flex-row">
          <Link
            href={primaryHref}
            className="group flex items-center gap-3 rounded-[14px] bg-teal-600 px-10 py-[18px] shadow-[0_8px_40px_rgba(13,148,136,0.27)] transition-colors hover:bg-teal-700"
          >
            <motion.span
              animate={reduced ? undefined : { y: [0, -2, 0] }}
              transition={reduced ? undefined : { duration: 1.6, repeat: Infinity, ease: "easeInOut" }}
              className="flex items-center"
            >
              <Rocket className="h-[22px] w-[22px] text-white" />
            </motion.span>
            <span className="text-[17px] font-bold text-white">
              开始你的斗学冒险
            </span>
          </Link>
          <Link
            href="/docs"
            className="flex items-center gap-2.5 rounded-[14px] border-[1.5px] border-slate-200 bg-white/70 px-8 py-[18px] transition-colors hover:bg-white"
          >
            <span className="text-[15px] font-medium text-slate-900">
              查看使用文档
            </span>
            <ArrowRight className="h-[18px] w-[18px] text-slate-900" />
          </Link>
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Type check + lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git add src/features/web/home/components/final-cta-section.tsx
git commit -m "feat(web/home): tighten final-cta with rocket idle and doc CTA"
```

---

## Task 16: Rewrite `page.tsx` composition and delete obsolete sections

**Files:**
- Modify: `dx-web/src/app/(web)/(home)/page.tsx`
- Delete:
  - `dx-web/src/features/web/home/components/stats-section.tsx`
  - `dx-web/src/features/web/home/components/testimonials-section.tsx`
  - `dx-web/src/features/web/home/components/smart-vocabulary-section.tsx`
  - `dx-web/src/features/web/home/components/course-platform-section.tsx`
  - `dx-web/src/features/web/home/components/social-community-section.tsx`

- [ ] **Step 1: Replace `page.tsx` contents**

```tsx
// dx-web/src/app/(web)/(home)/page.tsx
import { cookies } from "next/headers";
import { StickyHeader } from "@/components/in/sticky-header";
import { HeroSection } from "@/features/web/home/components/hero-section";
import { WhyDifferentSection } from "@/features/web/home/components/why-different-section";
import { FeaturesSection } from "@/features/web/home/components/features-section";
import { AiFeaturesSection } from "@/features/web/home/components/ai-features-section";
import { LearningLoopSection } from "@/features/web/home/components/learning-loop-section";
import { CommunitySection } from "@/features/web/home/components/community-section";
import { MembershipSection } from "@/features/web/home/components/membership-section";
import { FaqSection } from "@/features/web/home/components/faq-section";
import { FinalCtaSection } from "@/features/web/home/components/final-cta-section";
import { Footer } from "@/components/in/footer";

export default async function HomePage() {
  const cookieStore = await cookies();
  const isLoggedIn = !!cookieStore.get("dx_token")?.value;

  return (
    <div className="flex min-h-screen w-full flex-col bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white bg-[length:100%_620px] bg-top bg-no-repeat">
      <StickyHeader isLoggedIn={isLoggedIn} transparent />
      <HeroSection isLoggedIn={isLoggedIn} />
      <WhyDifferentSection />
      <FeaturesSection />
      <AiFeaturesSection />
      <LearningLoopSection />
      <CommunitySection />
      <MembershipSection />
      <FaqSection />
      <FinalCtaSection isLoggedIn={isLoggedIn} />
      <Footer />
    </div>
  );
}
```

- [ ] **Step 2: Verify deleted files are not referenced anywhere else**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
```

Use Grep to search for each of the five soon-deleted filenames across `src/`. Expected: only the home `page.tsx` / `features-content.tsx` may reference them. `features-content.tsx` is used by `/features` page and does not import any of the five; confirm.

- [ ] **Step 3: Delete obsolete files**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
git rm src/features/web/home/components/stats-section.tsx \
       src/features/web/home/components/testimonials-section.tsx \
       src/features/web/home/components/smart-vocabulary-section.tsx \
       src/features/web/home/components/course-platform-section.tsx \
       src/features/web/home/components/social-community-section.tsx
```

- [ ] **Step 4: Type check + lint + build**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
npm run build
```

Expected: all three pass. Build output should succeed without missing-module errors.

- [ ] **Step 5: Browser check (full landing)**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run dev
```

Open `http://localhost:3000/`. Walk through every section top to bottom:
- Hero: headline, sub, both CTAs, live demo with both scenes.
- Why-different: 3 rows reveal with stagger.
- 4 game modes: card hover plays mini preview.
- AI 随心学: streaming typewriter runs.
- Learning loop: dot travels across arrow; vocab three books appear.
- Community: 4 cards incl. streak heatmap.
- Membership: tier cards with real ¥ prices, referral band.
- FAQ: first item auto-open; each 查看完整说明 link works.
- Final CTA: rocket idle-floats; both CTAs route correctly.

Also verify:
- Header transparent above 40px scroll, solid below.
- Clicking `了解更多` scrolls to `#features`.
- Clicking `开始斗学之旅` while logged out goes to `/auth/signup`; set a fake `dx_token` cookie in DevTools Application panel and reload to verify logged-in CTA goes to `/hall`.

Kill dev server.

- [ ] **Step 6: Commit**

```bash
git add src/app/\(web\)/\(home\)/page.tsx
git commit -m "feat(web/home): rewrite / composition and drop obsolete fake sections"
```

---

## Task 17: Responsive & reduced-motion pass

**Files:** none (verification only — fix inline if issues found)

- [ ] **Step 1: Resize sweep**

`npm run dev`. For each width — 1440, 1280, 1024, 768, 640, 480, 375, 320 — scroll the full page. Check no horizontal overflow, no overlapping text, typography scales cleanly, grids collapse as expected.

If any section overflows, fix the offending component's classes (typical fixes: add `break-words`, tighten `max-w-*`, convert `px-[120px]` to the `px-5 md:px-10 lg:px-[120px]` pattern, reduce `text-[XX]px` on the smallest breakpoint). Commit any fixes as `fix(web/home): responsive polish — <section>`.

- [ ] **Step 2: Reduced-motion check**

In DevTools → Rendering → "Emulate CSS media feature prefers-reduced-motion" → `reduce`.

Reload `/`. Confirm:
- Hero demo shows end frames only (no tweening).
- Section reveals still transition but without large y-offset animation.
- Rocket stops idle-floating.
- Membership recommended card no longer pulses.
- Streak heatmap cells appear at their final colors.
- Typewriter in AI section renders all lines at once.

If any infinite loop or intra-scene animation still runs under reduced-motion, find the offending `animate` / `transition` in the component and guard it with the `reduced` flag from `usePrefersReducedMotion()`. Commit.

- [ ] **Step 3: Build sanity**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run build
```

Expected: passes. Note the `/` route's reported JS size — sanity check it's within ~50 KB of the previous build size + `motion`'s ~20 KB delta.

- [ ] **Step 4: Commit any polish changes** (if nothing was changed, skip this step).

---

## Task 18: Full verification + unrelated-route sanity

**Files:** none (verification only)

- [ ] **Step 1: Unrelated routes still work**

`npm run dev`. Manually open each of:
- `/features`
- `/docs`
- `/docs/account/faq`
- `/docs/invites/referral-program`
- `/docs/membership/tiers-compare`
- `/purchase/membership`
- `/auth/signup`
- `/auth/signin`
- `/hall` (with and without a cookie — unauth should redirect per existing guard)

Each should render without runtime errors. If any throws, investigate — the redesign should not have touched anything under those routes.

- [ ] **Step 2: Link audit on `/`**

On `/`, open every link/button and confirm the destination exists (not a 404):
- `/auth/signup`, `/auth/signin` (via header), `/hall`, `#features`, `/purchase/membership`, `/docs`, `/docs/invites/referral-program`, plus all 9 FAQ `查看完整说明` targets.

- [ ] **Step 3: Lighthouse**

In DevTools → Lighthouse → Desktop → Performance + Accessibility + Best Practices + SEO → Analyze.

Expected: Performance ≥90, Accessibility ≥95, Best Practices ≥95, SEO ≥95. If Accessibility is lower, check image alt / color contrast / landmarks. If Performance is lower, check bundle size and oversized images.

Attach the Lighthouse numbers to the final commit message.

- [ ] **Step 4: Final `git status` clean**

```bash
git status
```

Expected: clean tree.

- [ ] **Step 5: Final commit if any fixes came out of verification**

If fixes were made, each fix should already be its own commit from Task 17. If Step 1/2/3 surfaced issues and you fixed them here, commit as:

```bash
git add .
git commit -m "fix(web/home): verification polish — <what>"
```

If nothing needed fixing, skip.

---

## Post-plan notes

- No backend or docker changes. Pure frontend redesign of `/`.
- `motion` (successor to framer-motion) lives in `dx-web/package.json` only.
- All pricing is sourced from `@/consts/user-grade`; any upstream change flows through automatically.
- All FAQ answers link to existing real docs; update the FAQ answer text if the linked doc text drifts.
- If a future task adds real user testimonials or real usage stats, the slot between `community` and `membership` is a natural place to reintroduce them.
