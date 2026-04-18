# Home Page Hero Safari Rendering Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the home landing page so Safari 26.1 renders the hero envelope gradient and centers the hero content correctly, matching the existing Chrome behavior.

**Architecture:** Frontend-only single-file change. The home page wrapper is restructured so (a) the multi-stop gradient lives on its own absolutely-positioned decorative `<div>` with no conflicting `background-size` / `background-repeat` shorthands, and (b) the wrapper itself uses `items-center` (matching the working auth/purchase layouts) so flex children honor `max-w-[1280px]` on Safari.

**Tech Stack:** Next.js 16 App Router (server component), React 19, TailwindCSS v4. No test framework configured for dx-web — verification is `npm run lint`, `npm run build`, and manual checks in Safari 26.1 + Chrome.

**Spec:** `docs/superpowers/specs/2026-04-18-home-hero-safari-rendering-fix-design.md`

---

## File Structure

One file changes. No new files, no deletions, no other files touched.

Frontend (dx-web):
- `src/app/(web)/(home)/page.tsx` — the home page wrapper `<div>` className changes; one decorative `<div aria-hidden>` is added as the first child of the wrapper.

No backend files change.

---

## Task 1: Restructure the home page wrapper

**Files:**
- Modify: `dx-web/src/app/(web)/(home)/page.tsx` (lines 18-32)

- [ ] **Step 1: Replace the wrapper `<div>` and add the decorative gradient layer**

Open `dx-web/src/app/(web)/(home)/page.tsx`. The current `return` block (lines 18-32) is:

```tsx
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
```

Replace it with:

```tsx
  return (
    <div className="relative isolate flex min-h-screen w-full flex-col items-center">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[620px] bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white"
      />
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
```

Notes on each className delta — these are intentional and required:
- `relative` (kept) — anchors the absolutely-positioned gradient layer to the wrapper.
- `isolate` (added) — `isolation: isolate`. Creates a new stacking context so `-z-10` on the gradient layer cannot bleed behind the document body or sibling stacking contexts.
- `items-center` (added) — `align-items: center`. Restores `max-w-[1280px]` honoring on Safari for all child sections that use `mx-auto max-w-[1280px]`.
- `bg-gradient-to-b … to-white bg-[length:100%_620px] bg-top bg-no-repeat` (removed from wrapper) — these moved to the new dedicated layer (minus the `bg-[length:…]` / `bg-top` / `bg-no-repeat`, which are now expressed as `h-[620px]` on the layer).
- The decorative layer uses `pointer-events-none` so it never intercepts clicks (the StickyHeader sits at the same top region) and `aria-hidden="true"` so screen readers ignore it.

- [ ] **Step 2: Run lint to verify the change compiles cleanly**

Run from `dx-web/`:

```bash
npm run lint
```

Expected: exits 0 with no new errors or warnings attributed to `src/app/(web)/(home)/page.tsx`.

- [ ] **Step 3: Run build to type-check the whole dx-web project**

Run from `dx-web/`:

```bash
npm run build
```

Expected: build succeeds with no TypeScript errors. (This catches any prop or JSX issue that `eslint` alone would miss.)

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/app/\(web\)/\(home\)/page.tsx
git commit -m "fix(web): restructure home wrapper so Safari renders hero gradient and max-width"
```

---

## Task 2: Manual browser verification

**Files:** none modified. This task confirms the visual fix works in Safari and that Chrome and other pages are unchanged.

- [ ] **Step 1: Start dx-api and dx-web**

In one shell:

```bash
cd dx-api && go run .
```

In another:

```bash
cd dx-web && npm run dev
```

Expected: api listens on `:3001`, web on `:3000`.

- [ ] **Step 2: Verify the home page in Safari 26.1 at a wide viewport (≥ 1440 px)**

Open `http://localhost:3000/` in Safari with the window at ≥ 1440 px wide.

Expected:
- The top ~620 px shows a gradient envelope blending teal → blue → violet → pink → white from top to bottom (note: due to a Tailwind multi-`via-*` quirk that exists today, the effective stops are `teal-100 → violet-100 → white` — that is the pre-existing Chrome appearance and is preserved, not regressed).
- Below ~620 px the wrapper is transparent; each section's own background takes over as before.
- The hero "斗学，让你玩着玩着就会了" badge, the two h1 lines, the paragraph, the two CTA buttons, and the HeroGameDemo card are all centered with visible whitespace on both sides (max-width 1280 px is honored).
- The StickyHeader logo + nav stay clickable (the decorative gradient layer must not intercept clicks because it has `pointer-events-none`).

- [ ] **Step 3: Verify the home page in Chrome at a wide viewport (regression check)**

Open `http://localhost:3000/` in Chrome with the window at ≥ 1440 px wide.

Expected: visually identical to before this change. Same gradient envelope, same centered hero content, same StickyHeader behavior.

- [ ] **Step 4: Verify the home page at narrow viewport (≤ 768 px) in both browsers**

Resize each browser to ≤ 768 px wide (or open responsive devtools).

Expected:
- Hero content fills available width with the existing `px-5` / `md:px-10` paddings (max-width is not yet hit at this size).
- The 620 px gradient envelope is still visible at the top of the page.
- Layout is identical between Safari and Chrome.

- [ ] **Step 5: Verify scroll-driven behaviors still work**

In either browser on the home page:

1. Scroll down past 40 px.
2. Hover over the StickyHeader nav links.
3. Observe the HeroGameDemo card.

Expected:
- StickyHeader transitions to the white/blur background state once `scrollY > 40` (existing behavior).
- HeroGameDemo cycles through its four scenes (existing motion).
- All section motion (WhyDifferent rows, FeatureCards, FinalCta CTA) still triggers on scroll/in-view.

- [ ] **Step 6: Regression check — auth pages**

Open `http://localhost:3000/auth/signin` and `http://localhost:3000/auth/signup` in Safari.

Expected: render unchanged from before this change. Same gradient, same centered card. (We did not touch the auth layout.)

- [ ] **Step 7: Regression check — purchase pages**

Open `http://localhost:3000/purchase/membership` and `http://localhost:3000/purchase/beans` in Safari (sign in first if required).

Expected: render unchanged from before this change. Same gradient, same centered content. (We did not touch the purchase layout.)

- [ ] **Step 8: Final confirmation**

Report back: Safari hero gradient and max-width centering both render correctly; Chrome is visually unchanged; auth/purchase pages are unchanged; StickyHeader, HeroGameDemo, and section animations all still work; `npm run lint` and `npm run build` both passed in Task 1.

---

## Self-Review Notes

- **Spec coverage:** The spec's "Design → Wrapper restructure" maps to Task 1 Step 1 (the exact before/after blocks match the spec's blocks). The spec's "Testing" bullets map 1:1 to Task 2 Steps 2–7 (Safari wide, Chrome wide regression, narrow viewport, scroll behavior, click-through guard, auth regression, purchase regression). The spec's "Files Touched" lists exactly one file — that is the only file modified in Task 1.
- **No placeholders:** Every step has either the exact code (Task 1 Step 1) or the exact command + expected output (all other steps). No "TBD", no vague "verify it works".
- **Type consistency:** Only JSX is changed. All identifiers used in the new wrapper (`StickyHeader`, `HeroSection`, `WhyDifferentSection`, `FeaturesSection`, `AiFeaturesSection`, `LearningLoopSection`, `CommunitySection`, `MembershipSection`, `FaqSection`, `FinalCtaSection`, `Footer`) are already imported at the top of `page.tsx` and unchanged. No new imports required.
- **Why no automated test:** dx-web has no test runner configured (per `package.json` scripts). The bug is a visual rendering difference between Safari and Chrome — the only meaningful verification is opening both browsers, which Task 2 covers exhaustively.
