# Home Landing Page Redesign — Design Spec

**Date:** 2026-04-17
**Scope:** `dx-web` only — route `/` (file: `src/app/(web)/(home)/page.tsx`) and everything it composes
**Status:** Approved by user for implementation planning

---

## 1. Goal & Positioning

Replace the current mocked landing page with an industry-grade, motion-rich home landing that faithfully advertises the **real** Douxue product using the existing teal theme.

**Primary audience:** Chinese-speaking English learners, ages **16–40**.
**Tone:** Energetic, fun, competitive; playful but not juvenile.
**Hook:** *"别「学」英语了，玩着玩着就会了"* (kept from current page).
**Brand spine:** teal-600 primary (`#0d9488`), violet-500 accent, soft multi-hue top gradient kept subtle.

**Non-goals**
- No redesign of header, footer, or any route other than `/`.
- No i18n or dark-mode variant for this page.
- No API, model, or design-token changes. `globals.css` is untouched.
- No fake data anywhere. Every number, tier, benefit, FAQ answer must come from the real product.

---

## 2. Information Architecture

Desktop-first. `max-w-[1280px]`, gutters `px-[120px]` desktop / `px-10` tablet / `px-5` mobile, `py-[100px]` sections on desktop (reduced on mobile).

| # | Section ID | Purpose | Delta vs. current |
|---|---|---|---|
| 1 | `hero` | 30-second elevator pitch + live game demo | Refreshed (adds demo) |
| 2 | `why-different` | Traditional cramming vs. 斗学 contrast band | **New** (replaces fake stats) |
| 3 | `features` | 4 game modes with per-card mini animated preview | Refreshed |
| 4 | `ai-features` | **AI 随心学 only** (no AI 随心练) | Refreshed |
| 5 | `learning-loop` | Course progression + vocab三本 (生词/复习/已掌握) merged | **Merged** from 2 sections |
| 6 | `community` | PK · 学习群 · 斗学社 · 排行榜 · 连续打卡 streak | Refreshed (adds streak card) |
| 7 | `membership` | 4 tier cards + referral 30% band | **New** |
| 8 | `faq` | 9 real answers from existing docs | **New** (replaces fake testimonials) |
| 9 | `final-cta` | `准备好把英语打通关了吗？` | Kept + tightened |

Footer unchanged (already holds legal links, ICP, 公安备案).

**Composition in `page.tsx`:**
`StickyHeader → Hero → WhyDifferent → Features → AiFeatures → LearningLoop → Community → Membership → Faq → FinalCta → Footer`

**Deleted files** (no dead code kept):
- `features/web/home/components/stats-section.tsx`
- `features/web/home/components/testimonials-section.tsx`
- `features/web/home/components/smart-vocabulary-section.tsx`
- `features/web/home/components/course-platform-section.tsx`
- `features/web/home/components/social-community-section.tsx`

---

## 3. Hero Section

### Layout
Above-the-fold block, centered column, `py-20` top / `pb-[100px]` bottom on desktop.

```
[•] 斗学，让你玩着玩着就学会英语
        别「学」英语了
      玩着玩着就会了      ← teal-400→violet-500 gradient
多种游戏模式 · AI 定制内容 · 和朋友一起闯关
每天 10 分钟，英语悄悄就流利了
[ 开始斗学之旅 ]   [ 了解更多 → ]
┌──────── HeroGameDemo (≈ 900×420) ────────┐
│  [ 连词成句 | 词汇对轰 ]  tabs pill       │
│  <animated scene>                         │
└───────────────────────────────────────────┘
```

### Copy (final)
- **Eyebrow:** `斗学，让你玩着玩着就学会英语`
- **H1a** (`text-slate-900`): `别「学」英语了`
- **H1b** (gradient text-transparent, `from-teal-400 to-violet-500`): `玩着玩着就会了`
- **Sub** (`text-slate-500`): `多种游戏模式 · AI 定制内容 · 和朋友一起闯关 — 每天 10 分钟，英语悄悄就流利了`
- **Primary CTA:** `开始斗学之旅` — href `/auth/signup` when logged-out, `/hall` when logged-in. Teal-600 with teal-shadow `shadow-[0_4px_30px_rgba(13,148,136,0.27)]`.
- **Secondary CTA:** `了解更多 →` — anchor `#features`.

### HeroGameDemo component
**File:** `features/web/home/components/hero-game-demo.tsx` (client; dynamic import with `ssr: false`; mounted only after hero enters viewport via `useInView`).

- Outer frame: white card, `rounded-2xl`, `border border-slate-200`, `shadow-[0_12px_40px_rgba(13,148,136,0.12)]`, soft radial teal glow behind it.
- Top-left tab pill lets user manually toggle scenes: `连词成句 | 词汇对轰`.
- Scene auto-rotation: **6s per scene** with a 400ms crossfade. Rotation **pauses** when:
  - User hovers over the card, OR
  - The demo IntersectionObserver reports <20% visibility, OR
  - `prefers-reduced-motion: reduce` is set (in that case, render the end-state frame only and still rotate, but without intra-scene motion).

**Scene A — 连词成句 (~5s loop):**
1. `t=0–720ms`: Six colored word chips drop in (stagger 120ms, spring). Words: `I / eat / fresh / apples / every / morning`. Each chip uses an existing tailwind palette pair (teal/violet/pink/amber/rose/blue @ `-100` bg, `-700` text).
2. `t=720–1200ms`: Chips briefly appear in a shuffled order.
3. `t=1200–2400ms`: Chips reorder to the correct sentence via spring transition; green underline sweeps L→R.
4. `t=2400–3000ms`: `+128 ★ · combo x2` badge pops with a small bounce; score bar fills teal→violet.
5. `t=3000–5000ms`: Hold the final frame; fade out into crossfade.

**Scene B — 词汇对轰 (~5s loop):**
1. `t=0–600ms`: Opponent avatar slides in from right; a Chinese prompt chip (e.g. `勇气`) appears in the cannon barrel.
2. `t=600–1600ms`: User side shows an input row; letters typewriter-appear spelling `courage` (100ms per letter).
3. `t=1600–2400ms`: A teal projectile flies user→opponent along a slight arc; on impact, `shake` keyframe on opponent HP bar; `-15 HP` floats up.
4. `t=2400–3000ms`: Combo counter flashes `combo x3`.
5. `t=3000–5000ms`: Hold + crossfade.

**Motion library:** `motion` (npm package — imperative `animate()` + `stagger()` + `spring`; React helper `<motion.*>` used for reveal-on-scroll elsewhere).

### Accessibility
- `HeroGameDemo` wrapper has `aria-hidden="true"`; demo is purely decorative.
- A visually-hidden `<p>` inside hero text describes the demo for screen readers: *"示例游戏循环演示连词成句和词汇对轰两种玩法。"*
- Headline/sub text carry the real semantic content.

### Page top background
Keep the existing multi-hue top gradient but tone it down:
- Height: `620px` (was 720px)
- Stops: `teal-100 → blue-100 → violet-100 → pink-100 → white`
- Behind the demo card, add a `bg-[radial-gradient(ellipse_at_center,rgba(94,234,212,0.35),transparent_70%)]` halo.

---

## 4. Why-Different Section

Replaces the fake stats section. Grounds the hook with an honest contrast.

- Eyebrow: `为什么是斗学` (`text-teal-600`)
- H2: `不再死记硬背，让大脑爱上英语`
- Three horizontal rows (stacked vertically on mobile):

| 以前 (muted slate) | 斗学 (teal/violet accented) |
|---|---|
| 背了就忘，靠意志力硬撑 | 游戏化循环，大脑自发想再玩一局 |
| 学的和用的两张皮 | 连词成句、对话、对战都是真实语料 |
| 一个人孤独地学 | 好友开黑 · 学习群 · 排行榜 · 每日连胜 |

Motion: when section reaches 35% in-view, each row reveals left→right with a 60ms stagger; right cells slide in from +16px with a soft teal bloom behind them.

---

## 5. Game Modes Section (refreshed `features-section.tsx`)

- Eyebrow: `核心玩法 · 4 种模式，覆盖听说读写` (`text-violet-500`)
- H2: `挑一个就能上手，全都玩转就起飞`
- 2×2 grid on desktop (2-col tablet, 1-col mobile).

Each card keeps its existing icon+palette pairing:
| Mode | Icon | Palette |
|---|---|---|
| 连词成句 | Keyboard | violet-600 on violet-100 |
| 词汇配对 | Swords | teal-600 on blue-100 |
| 词汇消消乐 | Shuffle | emerald-600 on pink-100 |
| 词汇对轰 | Crosshair | red-500 on red-100 |

Card structure: icon chip → title → one-sentence benefit → **mini preview strip** (height ~80px) at card bottom.

**`GameModePreview` mini loops** (new component, one variant per mode):
- **连词成句:** three tiny word chips shuffle then snap to correct order.
- **词汇配对:** two columns of 3 chips each; matching pairs highlight and fade.
- **词汇消消乐:** a 3×3 grid where two adjacent cells pulse and pop.
- **词汇对轰:** a small projectile flies across the strip; opponent HP bar ticks down.

Previews: **hover-play on desktop**, **autoplay-when-visible on mobile** (no hover available).

Motion: cards reveal with 60ms stagger; on hover, card lifts `translate-y-[-3px]`, ring glows teal-100/50, preview strip plays. Non-hovered cards render a single static end-frame.

---

## 6. AI 随心学 Section (refreshed `ai-features-section.tsx`)

AI 随心练 is removed entirely from the landing.

### Layout
Two-column split (stacks on mobile):

- **Left column**: eyebrow `AI 驱动 · 专属于你的学习` → H2 `AI 帮你定制课程，你想学什么都可以` → 3 feature bullets:
  - `输入任意主题或场景，AI 按你的水平生成课程`
  - `CEFR A1–C2 全覆盖，难度智能匹配`
  - `内容沉淀进你的词汇系统，复习自动推送`
- **Right column**: a mock "create-with-AI" card showing:
  - Input field value: `职场面试高频词`
  - CEFR level pill: `B1–B2` (teal)
  - Streaming output: 5 lines of vocab+sentence typed out one by one (180ms per character, 300ms between lines) using a typewriter effect that mirrors the real SSE generation flow.
  - Bottom chip: `消耗 5 能量豆` with a small bean icon — honest about the real cost mechanic.

Motion: when entering view, the level pill does a small spring-pop; the typewriter loops on-screen (resets after a 2s hold) while the section is visible.

---

## 7. Learning Loop Section (new, merges course-platform + smart-vocabulary)

- Eyebrow: `学习闭环 · 从陌生到掌握` (`text-violet-500`)
- H2: `一套系统，追踪你每一个单词的命运`

### Layout
Single horizontal diagram row (stacks vertically on mobile):

```
[ Lv.01 ✓ ]       →  每次游戏自动沉淀单词  →       [ 生词本  ]
[ Lv.02 ✓ ]                                        [ 复习本  ]
[ Lv.03 • ]                                        [ 已掌握  ]
```

- Left stack: three level cards with a teal progress line connecting them; current level pulses.
- Middle arrow: a thin teal line with a travelling dot.
- Right stack: three mini cards using the existing vocab icons/colors (BookOpen/pink, RefreshCw/violet, CircleCheck/teal), each with a small count chip + a tiny filled bar.

Below the diagram, one line: `艾宾浩斯遗忘曲线智能推送复习，你只需要玩。`

Motion: on scroll-in, the left stack fades in first; the travelling dot animates from left to right along the arrow; as it reaches the right, each vocab card flip-reveals (120ms stagger).

---

## 8. Community Section (refreshed `social-community-section.tsx` → `community-section.tsx`)

- Eyebrow: `一起玩才好玩` (`text-orange-600`)
- H2: `和朋友开黑，排行榜上见`

Four cards (3-col on desktop with the 4th spanning the first row end, or a clean 4-col row; final call during implementation based on visual weight):

1. **排行榜** (Trophy / amber) — weekly & monthly;登上领奖台.
2. **斗学社** (MessageSquare / orange) — 分享心得，交流技巧.
3. **学习群** (Users / blue) — 组队闯关，群内发起课程游戏.
4. **连续打卡 🔥** (new) — 每日保持连胜，错过一天从 1 重来. Card renders a 7-cell mini-heatmap (reusing the hall heatmap color palette from `progress-color.helper.ts`).

Motion: streak card's 7 cells fill one at a time (80ms stagger) when visible; leaderboard card shows a small rank number counter that ticks `0 → 7` once on reveal.

**Truth note:** every card description must match what the feature actually does (no invented weekly prizes, etc.). Pull language from the docs registry.

---

## 9. Membership + Referral Section (new)

### Tier cards row
- Eyebrow: `会员计划`
- H2: `选一个你最舒服的节奏，越早开始越划算`
- Four cards (4-col desktop, 2×2 tablet, 1-col mobile):
  - **体验** — 免费
  - **月度会员**
  - **年度会员** — badge `推荐`, teal ring
  - **永久会员** — badge `最超值`, violet ring
- **All prices and feature bullets pulled from the same source the real `/purchase/membership` page uses.** Implementation rule: either import the shared data module, or — if it is server-fetched — mirror the current values and flag a single source-of-truth comment pointing at the real route. Do **not** hard-code new pricing by hand.
- Existing `features/web/purchase/components/pricing-card.tsx` should be reused/extended rather than duplicated.

### Referral callout (band immediately below tier cards)
- Full-width, violet→teal subtle gradient bg.
- Left: `成为永久会员，把斗学推荐给朋友`
- Right: `邀请好友首充，你拿 30% 终身返佣`
- Small link: `详细规则` → existing invite/referral doc route (exact path determined from `docs/registry.ts` during implementation).

Motion: recommended tier card floats `translate-y-[-4px]` with a slow teal glow pulse (2s ease-in-out loop); on scroll-in, cards fade-stagger-in; referral band slides up with a subtle left-to-right gradient sweep on reveal.

---

## 10. FAQ Section (new, replaces testimonials)

- Eyebrow: `常见问题`
- H2: `开始之前，你可能想问`
- Accordion using the existing `components/ui/accordion.tsx`.

Proposed 9 items (all sourced from real docs; final wording confirmed in implementation by cross-reading `docs/`):

1. `斗学适合什么水平的学习者？`
2. `能量豆是什么？怎么获得和消耗？`
3. `免费用户能玩到哪些内容？`
4. `会员自动续费怎么关掉？`
5. `未成年人可以用吗？需要家长同意吗？`
6. `支持微信支付和支付宝吗？`
7. `我的学习数据会怎么保护？`
8. `邀请返佣什么时候到账？`
9. `如何提交反馈或报告游戏问题？`

Each answer is 1–3 short sentences plus a `查看完整说明 →` link to the real doc. No invented policy text — everything traceable to `docs/` or the in-app legal agreements.

Motion: accordion uses shadcn's default height easing; on section reveal, the first item auto-opens once (nudge) to teach the pattern, then user interactions take over.

---

## 11. Final CTA Section (kept, tightened)

- Gradient divider line (kept).
- H2: `准备好把英语打通关了吗？`
- Sub: `30 秒注册，今晚就能陪你玩到手滑。`
- Primary CTA: `开始你的斗学冒险` — `/auth/signup` (logged-out) / `/hall` (logged-in). Rocket icon. Teal-600 with `shadow-[0_8px_40px_rgba(13,148,136,0.27)]`.
- Secondary CTA: `查看使用文档 →` → `/docs`.

Motion: rocket icon has a continuous 2px idle float (1.6s ease-in-out); on CTA hover, rocket translates up 4px with a short gradient-trail streak.

---

## 12. Motion System

### Dependency
- `npm i motion` in `dx-web/` (adds `motion` to `package.json`). `motion` is the modern successor to framer-motion; use the package's `animate()` imperative API for hero demo, and `<motion.*>` React components for declarative reveal-on-scroll.
- No other new runtime deps.

### Shared presets
Located at `features/web/home/helpers/motion-presets.ts`:
```ts
export const revealEase = { duration: 0.45, ease: [0.22, 1, 0.36, 1] } as const;
export const revealSpring = { type: "spring", stiffness: 160, damping: 24, mass: 0.6 } as const;
export const reveal = {
  initial: { opacity: 0, y: 24 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true, amount: 0.35 },
  transition: revealEase,
} as const;
export const staggerContainer = {
  whileInView: { transition: { staggerChildren: 0.06 } },
} as const;
```

### `useInView` hook
At `features/web/home/hooks/use-in-view.ts`. Thin wrapper over `IntersectionObserver` that:
- Exposes `(ref, inView: boolean)` plus `prefersReducedMotion: boolean`.
- When `prefersReducedMotion` is true, consumers must skip intra-scene animations and render end-state frames.

### `useRotatingScene` hook
At `features/web/home/hooks/use-rotating-scene.ts`. Drives the hero demo's 6-second scene cycle; respects visibility (pauses when hero leaves viewport) and reduced-motion.

### Scroll-reveal contract
- Each mid-page section wraps its content in a `<motion.div variants={reveal}>` or uses `whileInView` directly; `viewport.once = true` guarantees the animation fires once per load.
- Stagger is applied only where there's a collection (cards, rows).
- No parallax. No scroll scrubbing. No persistent background motion.

### Reduced-motion rule
For every `motion.*` element:
- If `prefersReducedMotion`, both `initial` and `whileInView` collapse to the end state (`opacity: 1, y: 0`).
- Hero demo renders end frames only; no tweening, but scene rotation still happens.
- No infinite loops (idle float, pulse, shimmer) when reduced motion is on.

---

## 13. Responsive Treatment

| Breakpoint | Gutters | Hero typography | Grid behavior |
|---|---|---|---|
| Desktop ≥1024px | `px-[120px]` | H1 `text-[72px]`, sub `text-lg` | Game modes 2×2, tiers 4-col, community 4-col |
| Tablet 640–1023px | `px-10` | H1 `text-6xl`, sub `text-base` | Game modes 2-col, tiers 2×2, community 2-col |
| Mobile <640px | `px-5` | H1 `text-5xl`, sub `text-sm` | All 1-col; demo `aspect-[5/3] w-full`; section padding `py-[64px]` |

Game-mode and hero-demo animations autoplay on mobile (no hover). Touch targets: tab pill in hero demo becomes a real shadcn `tabs` on mobile for fat-finger safety.

---

## 14. Accuracy & Truthfulness Guardrails

- **Pricing numbers:** single source with `/purchase/membership` page. Either shared import or explicit mirror + comment. Any change upstream must be reflected.
- **FAQ answers:** one-to-one traceable to a doc under `dx-web/src/features/web/docs/topics/*` or the legal agreements in `dx-source/docs/`.
- **Referral mechanics:** surface only what the service agreement documents; link to the agreement rather than paraphrase contract clauses.
- **Community / leaderboard descriptions:** match the actual product behaviour (e.g., weekly/monthly rankings if that's real; streak reset rules match `update-play-streaks` scheduled job).
- **No invented counts, users, reviews, logos.** If a stat/testimonial is added back in the future, it must come from real data.

---

## 15. File Structure

```
dx-web/src/features/web/home/
├── components/
│   ├── hero-section.tsx
│   ├── hero-game-demo.tsx                 # NEW, client-only, dynamic import
│   ├── hero-game-demo-word-sentence.tsx   # NEW, Scene A
│   ├── hero-game-demo-vocab-battle.tsx    # NEW, Scene B
│   ├── why-different-section.tsx          # NEW
│   ├── features-section.tsx
│   ├── game-mode-preview.tsx              # NEW, 4 variants keyed by mode
│   ├── ai-features-section.tsx            # refreshed (随心学 only)
│   ├── learning-loop-section.tsx          # NEW (merged)
│   ├── community-section.tsx              # renamed from social-community-section
│   ├── membership-section.tsx             # NEW
│   ├── faq-section.tsx                    # NEW
│   └── final-cta-section.tsx
├── hooks/
│   ├── use-in-view.ts                     # NEW
│   └── use-rotating-scene.ts              # NEW
└── helpers/
    └── motion-presets.ts                  # NEW
```

**Deleted files:** `stats-section.tsx`, `testimonials-section.tsx`, `smart-vocabulary-section.tsx`, `course-platform-section.tsx`, `social-community-section.tsx`, `features-content.tsx` if unreferenced after refresh (verify before delete).

**Edited outside `home/`:** none required. `StickyHeader`, `Footer`, `globals.css`, all other features remain untouched.

---

## 16. Verification Checklist (blocks "done")

1. `cd dx-web && npm i motion` — confirm `package.json` + lockfile updated; only `motion` added.
2. `npm run lint` — zero errors, zero warnings in all new/changed files. Fix anything ESLint surfaces.
3. `npm run build` — succeeds. Check bundle-size delta; `motion` should add ≲20 KB gzipped to the `/` route.
4. `npm run dev` manual pass:
   1. `/` logged-out — hero demo loops both scenes; scene switcher works; all reveals fire once; primary CTA → `/auth/signup`, secondary → `#features`.
   2. `/` logged-in — primary CTA → `/hall`; header shows `进入学习大厅`.
   3. Scroll top→bottom on a 4× CPU-throttled profile — no jank, no layout shift.
   4. DevTools → emulate `prefers-reduced-motion: reduce`: intra-scene animations skipped, content legible, no looping idle effects.
   5. Resize 1440 → 1024 → 768 → 375 → 320 px: no overflow, no overlap, typography scales cleanly.
   6. Click every CTA / anchor on `/`: route exists (`/auth/signup`, `/hall`, `/docs`, `/docs/*`, `/purchase/membership`, `/features`, `#features`).
   7. `/features`, `/hall`, `/docs`, `/purchase/membership`, `/auth/signup` still render — sanity check nothing unrelated broke.
5. Lighthouse on `/`: Performance ≥90 desktop, Accessibility ≥95, Best-Practices ≥95, SEO ≥95. Investigate any regression vs. current landing.
6. Confirm every hard number / benefit bullet is traceable to a real source (pricing module, real doc, real scheduled job, real service agreement).

---

## 17. Open Implementation Questions (to resolve in plan)

These are intentionally deferred to the implementation plan since they depend on reading specific files:

1. Exact file-path + shape of the pricing data source used by `/purchase/membership`. The plan should read that file first and decide whether to import or mirror.
2. Exact slug/paths of the real FAQ-backing docs in `dx-web/src/features/web/docs/topics/*` and `docs/`. The plan should list them explicitly per FAQ item.
3. Whether `features-content.tsx` is still used by `/features` after refresh — if not, delete; if yes, leave untouched.
4. Whether the community section best reads as a 4-col row or a 3+1 layout once the streak card is built — tiny visual decision made during implementation with a screenshot check.

---

## 18. Out of Scope (explicit)

- Header, footer, navigation behavior.
- Dark mode for `/`.
- i18n on `/`.
- Any route other than `/`.
- Any API, service, controller, model, DB, or Goravel change.
- Design-token changes in `globals.css`.
- Replacing shadcn primitives.
