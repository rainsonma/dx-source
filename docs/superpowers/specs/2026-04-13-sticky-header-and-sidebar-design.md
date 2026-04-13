# Sticky Header + Sticky Docs Sidebar — Design Spec

**Date:** 2026-04-13
**Scope:** `dx-web` frontend only (no backend changes)
**Target pages:** `/`, `/features`, `/docs/*`

## Purpose

Two coordinated layout changes on the marketing and docs routes:

1. **Sticky `LandingHeader`** on home, features, and docs pages so it stays visible while scrolling.
2. **Sticky docs sidebar** on `/docs/*` so the menu list stays pinned below the header and scrolls independently from the main content.

## Current State

- `LandingHeader` is a `h-20` (80px) `max-w-[1280px]` centered header. It is rendered inline inside each page and is **not** sticky — it scrolls off-screen.
- On the home page (`app/(web)/(home)/page.tsx`), `LandingHeader` is rendered **inside** a gradient wrapper div together with `HeroSection`. If we made it sticky in that position, `position: sticky` would unstick when the gradient wrapper scrolled out of view.
- In `docs/layout.tsx`, the sidebar `<aside>` is a normal flex-row child inside a `flex flex-1` container. Its height follows content, so once the topic page is long enough the sidebar scrolls away with the rest of the page.
- The mobile drawer strip in `docs/layout.tsx` is already `sticky top-0` on mobile viewports — it will need to move to `top-20` once the header sits sticky at `top-0`.

## Goals

- Sticky header on `/`, `/features`, `/docs/*` with a frosted-glass look (`bg-white/80 backdrop-blur-md border-b border-slate-200`) — visible on scroll across the full page width.
- Sticky docs sidebar pinned at `top-20` (below the 80px header), height `calc(100svh - 5rem)`, with its inner list `overflow-y-auto` so the 48-topic menu can scroll independently of the main content.
- Main content continues to scroll as part of the page body (no separate scroll area). Footer still appears at the end of every page.
- No backend changes. No new npm dependencies. No JS scroll listeners (pure CSS `position: sticky`).

## Non-Goals

- Sticky header for `/hall/*` or `/auth/*` or `/purchase/*` (out of scope — those routes use different layouts).
- "Transparent on top, solid on scroll" effect (requires a client scroll listener — deferred; see approaches below).
- Changes to the `HallSidebar` or any `/hall/*` layout.
- Footer changes.

## Design

### New component: `StickyHeader`

Thin wrapper around `LandingHeader`. Location: `src/components/in/sticky-header.tsx`.

```tsx
import { LandingHeader } from "@/components/in/landing-header";

interface StickyHeaderProps {
  isLoggedIn?: boolean;
}

export function StickyHeader({ isLoggedIn = false }: StickyHeaderProps) {
  return (
    <div className="sticky top-0 z-50 w-full border-b border-slate-200 bg-white/80 backdrop-blur-md">
      <LandingHeader isLoggedIn={isLoggedIn} />
    </div>
  );
}
```

Why a separate component instead of mutating `LandingHeader`:

- `LandingHeader` has `max-w-[1280px]` and is centered. If we made it itself sticky, its sticky background would also be constrained to 1280px — not full-width.
- Keeping `LandingHeader` unchanged means the inner layout (brand, nav links, auth buttons) stays untouched.
- `StickyHeader` only owns the sticky positioning and background treatment. Single responsibility.

### Edit: `app/(web)/(home)/page.tsx` (home landing)

**Change**: hoist the header out of the gradient wrapper.

Before:
```tsx
<div className="flex min-h-screen w-full flex-col">
  <div className="flex w-full flex-col bg-gradient-to-b from-teal-100 via-blue-100 ...">
    <LandingHeader isLoggedIn={isLoggedIn} />
    <HeroSection />
  </div>
  <FeaturesSection />
  ...
```

After:
```tsx
<div className="flex min-h-screen w-full flex-col">
  <StickyHeader isLoggedIn={isLoggedIn} />
  <div className="flex w-full flex-col bg-gradient-to-b from-teal-100 via-blue-100 ...">
    <HeroSection />
  </div>
  <FeaturesSection />
  ...
```

Result: gradient is applied only to the hero area. Sticky header sits above the gradient. Through `bg-white/80 backdrop-blur-md`, the gradient shows slightly through the header on the first screen.

### Edit: `app/(web)/(home)/features/page.tsx`

Single substitution: `<LandingHeader isLoggedIn={isLoggedIn} />` → `<StickyHeader isLoggedIn={isLoggedIn} />`.

### Edit: `app/(web)/(home)/docs/layout.tsx`

Four coordinated changes:

1. Swap `<LandingHeader>` → `<StickyHeader>`.
2. Delete the existing separator `<div className="h-px w-full bg-slate-200" />` — `StickyHeader` already has `border-b`.
3. Mobile drawer strip: change `sticky top-0 z-10` → `sticky top-20 z-40` so it stacks below the 80px sticky header on viewports `<lg`.
4. Desktop `<aside>` sidebar: add sticky positioning + independent scroll.

Sidebar before:
```tsx
<aside className="hidden w-[260px] shrink-0 border-r border-slate-200 bg-white px-5 py-6 lg:block">
  <DocsSidebar />
</aside>
```

Sidebar after:
```tsx
<aside className="sticky top-20 hidden h-[calc(100svh-5rem)] w-[260px] shrink-0 overflow-y-auto border-r border-slate-200 bg-white px-5 py-6 lg:block">
  <DocsSidebar />
</aside>
```

Added classes: `sticky top-20 h-[calc(100svh-5rem)] overflow-y-auto`.

- `sticky top-20` — pins below the 80px (`h-20`) header. `top-20` = `5rem`.
- `h-[calc(100svh-5rem)]` — sidebar fills exactly the remaining viewport height. `svh` (small viewport height) is more reliable than `vh` across mobile browsers even though this `<aside>` only renders at `lg` (≥1024px).
- `overflow-y-auto` — the aside becomes its own scroll container. When the 48-topic menu exceeds the viewport, its internal scrollbar lets users scroll the menu without affecting page scroll.

Main content `<main>`: no change. Body scrolls normally. When the user scrolls to the footer, the sticky sidebar reaches its parent flex-row boundary and scrolls away (native `position: sticky` behavior). This is the desired behavior — footer is cleanly below both columns, not cluttered by a lingering sticky sidebar.

### Z-index hierarchy

| Layer | z-index | Notes |
|---|---|---|
| `StickyHeader` | `z-50` | Top of everything |
| Mobile drawer strip (<lg) | `z-40` | Below header, above content |
| Desktop sticky sidebar | default (no explicit z) | Later-painted than header, naturally below header in stacking |
| Shadcn Sheet drawer (mobile) | shadcn-managed | Non-issue |

## Approaches Considered

**A) Always solid white header** (`bg-white border-b`) — zero visual overhead, but flat. Rejected in favor of B.

**B) Translucent + backdrop-blur** (chosen) — `bg-white/80 backdrop-blur-md border-b`. Zero JS, looks good over the gradient hero, indistinguishable from solid white on plain-white pages (features, docs).

**C) Transparent on top, solid on scroll** — requires a client-side scroll listener, a `useEffect`, and state management. Deferred: marginal polish gain, disproportionate complexity.

**Footer in docs**: kept at the bottom of the page. Alternative was hiding the footer on docs pages (common in Stripe/Tailwind docs), but that's a bigger editorial decision and out of scope.

**Main content scroll**: let body scroll handle it. Alternative was a second `overflow-y-auto` container wrapping `<main>`, creating two fully independent scroll areas (sidebar + main, neither affecting page scroll). Rejected — adds complexity, breaks natural footer placement, and doesn't match any standard docs-site pattern.

## Acceptance Criteria

1. `StickyHeader` component exists at `src/components/in/sticky-header.tsx` and renders `LandingHeader` inside a sticky wrapper.
2. Home page, features page, and docs layout all import and use `StickyHeader` (no direct `<LandingHeader>` usage).
3. On home page: gradient wrapper contains only `HeroSection`; `StickyHeader` is outside the gradient wrapper.
4. Scroll on any of the three pages — header stays pinned to `top: 0`.
5. On `/docs/*` at viewport ≥1024px: sidebar sticks to `top-20`, its inner list scrolls independently of the main content when the list overflows.
6. On `/docs/*` at viewport <1024px: `<lg` mobile drawer strip sits directly below the sticky header (no overlap, no gap).
7. Footer appears at the bottom of every page after scrolling past content.
8. `npm run build` succeeds with no errors.
9. `npm run lint` clean.
10. No new npm dependencies, no new shadcn installs.

## Verification Plan

1. `npm run build` — expect success, all routes compile.
2. `npm run lint` — expect zero errors.
3. `npm run dev` and manually visit in a desktop browser:
   - `/` — scroll; verify header pins and gradient bleeds through subtly.
   - `/features` — scroll; verify header pins.
   - `/docs` — scroll; verify header pins, sidebar stays pinned, category grid scrolls beneath.
   - `/docs/learning-modes/pk-mode` — scroll long content; verify main scrolls, sidebar stays.
4. Resize browser to <1024px width: verify mobile drawer strip sits below header (no overlap, no gap).
5. Scroll to the very bottom of `/docs/getting-started/what-is-douxue` — verify footer appears and sticky sidebar has scrolled away with its parent row.

## Files Touched

- **Create**: `dx-web/src/components/in/sticky-header.tsx`
- **Modify**: `dx-web/src/app/(web)/(home)/page.tsx`
- **Modify**: `dx-web/src/app/(web)/(home)/features/page.tsx`
- **Modify**: `dx-web/src/app/(web)/(home)/docs/layout.tsx`

Total: 1 new file, 3 modified files, ~15 lines of change across all four.
