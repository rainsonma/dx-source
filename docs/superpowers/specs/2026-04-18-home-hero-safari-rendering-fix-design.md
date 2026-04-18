# Home Page Hero Safari Rendering Fix

## Problem

On the home landing page (`dx-web/src/app/(web)/(home)/page.tsx`), Safari 26.1 renders the hero area with two visible defects that Chrome does not:

1. **Background degenerates to a flat color** instead of the multi-color gradient envelope (teal → blue → violet → pink → white) that Chrome shows correctly.
2. **Hero content stretches edge-to-edge** instead of being constrained to `max-w-[1280px]` and centered horizontally.

Sections further down the page (WhyDifferent, Features, Membership, FaqSection, FinalCta), and the auth/purchase pages, all render correctly on Safari 26.1. The defect is localized to the home page wrapper.

## Root Cause

The home page wrapper combines two unusual CSS choices that Safari mishandles when stacked:

```jsx
// dx-web/src/app/(web)/(home)/page.tsx:19
<div className="flex min-h-screen w-full flex-col
                bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white
                bg-[length:100%_620px] bg-top bg-no-repeat">
```

**Cause 1 — gradient drops to a single color.**
Tailwind v4 compiles `bg-gradient-to-b` to `linear-gradient(to bottom in oklab, var(--tw-gradient-stops))`. The `in oklab` color-interpolation method is well-supported in Safari 26 on its own (the auth page proves it: same gradient classes, no defect). But when the same element also carries `bg-[length:100%_620px]` (`background-size: 100% 620px`) plus `bg-no-repeat`, Safari 26 collapses the multi-stop gradient to a flat color. Auth/purchase layouts use the identical gradient utilities without `bg-[length:…]` / `bg-no-repeat`, and they render fine.

**Cause 2 — content stretches past `max-w-[1280px]`.**
The auth/purchase wrappers use `items-center` (`align-items: center`); their flex children take their declared `max-width` and are centered. The home wrapper omits `items-center`, leaving the default `align-items: stretch`. Chrome resolves `align-items: stretch` + `width: 100%` + `max-width: 1280px` + `margin-inline: auto` by capping at the max-width and distributing the leftover space to the auto margins (visually centered). Safari, in this specific stacked configuration, resolves stretch first and ignores the auto margins, leaving children edge-to-edge at the container width.

The `bg-[length:100%_620px] bg-top bg-no-repeat` triple was added so the gradient appears only behind the top 620 px of the page (a "hero envelope" effect). The intent is correct; the implementation puts conflicting background shorthands on the same element as the multi-stop gradient.

## Scope

- **In scope:** the single home page wrapper in `dx-web/src/app/(web)/(home)/page.tsx`. Both symptoms collapse to this one element.
- **Out of scope:** auth layout, purchase layout, all section components below the hero (each has its own self-contained gradient that already works), StickyHeader, HeroSection internals, motion/animations, the Footer, the global theme, and the Tailwind/PostCSS toolchain.

## Design

### Wrapper restructure

Replace the single backgrounded `<div>` with two siblings: a positioning/layout wrapper and a dedicated absolutely-positioned gradient layer.

**Before:**

```jsx
<div className="flex min-h-screen w-full flex-col
                bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white
                bg-[length:100%_620px] bg-top bg-no-repeat">
  <StickyHeader isLoggedIn={isLoggedIn} transparent />
  <HeroSection isLoggedIn={isLoggedIn} />
  …
</div>
```

**After:**

```jsx
<div className="relative isolate flex min-h-screen w-full flex-col items-center">
  <div
    aria-hidden="true"
    className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[620px]
               bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white"
  />
  <StickyHeader isLoggedIn={isLoggedIn} transparent />
  <HeroSection isLoggedIn={isLoggedIn} />
  …
</div>
```

### Why each change earns its place

| Change | Reason |
|---|---|
| `items-center` added | Switches `align-items` from `stretch` to `center`. Children with `mx-auto max-w-[1280px]` are now centered correctly on Safari. Mirrors the working auth/purchase layouts. |
| `relative` retained | Anchors the absolutely-positioned gradient layer. |
| `isolate` added | Establishes a new stacking context so `-z-10` on the gradient layer is contained inside the wrapper — it cannot disappear behind body / sibling stacking contexts. |
| Gradient moved to its own `<div>` | The gradient is now alone on its element, with no `background-size`/`background-repeat` shorthands stacked on top of `linear-gradient(in oklab, …)`. Safari renders the multi-stop gradient correctly. |
| `h-[620px]` on the gradient layer | Replaces the old `bg-[length:100%_620px] bg-top bg-no-repeat` trick. Same visual: a 620 px tall colorful envelope at the top, transparent below. |
| `pointer-events-none` | The decorative layer must not intercept clicks (e.g., on the StickyHeader, which sits at the same top region). |
| `aria-hidden="true"` | Decorative layer is hidden from assistive tech. |

### What stays the same

- `StickyHeader` (`sticky top-0 z-50 w-full`) — sticky positioning still works inside a flex column with `items-center`; `w-full` keeps it edge-to-edge; `z-50` stays above the gradient's `-z-10`.
- `HeroSection` markup — unchanged. Its internal `mx-auto flex w-full max-w-[1280px] flex-col items-center …` now centers correctly because the parent allows it.
- All section components below the hero (`WhyDifferentSection` … `FinalCtaSection`) — unchanged; each is `w-full` with its own internal centering.
- `Footer` — unchanged; `w-full` so `items-center` is a no-op for it.
- Animations (`motion/react` in HeroSection's child demos, WhyDifferent rows, FeatureCards, FinalCta CTA) — unchanged; all motion is inside section components, decoupled from the wrapper.
- The exact gradient colors and the 620 px height — unchanged.

### What is intentionally not done

- **No global CSS, no custom utility class.** Tailwind utilities cover the fix; no `globals.css` change is needed. (Considered and rejected as unnecessary scope.)
- **No `style={{ backgroundImage: … }}` inline override.** Tailwind utilities work once the conflicting background shorthands are off the element.
- **No change to the auth or purchase layouts.** They already work; touching them risks regressions.
- **No change to the gradient color stops.** The Tailwind quirk where multiple `via-*` classes overwrite the same `--tw-gradient-via` variable (last one wins) is preserved exactly as today, so the rendered Chrome appearance does not change.

## Visual / Layout Verification

```
Before (Safari):                 After (Safari, matches Chrome):

┌─────────────────────────┐      ┌─────────────────────────┐
│ ░░░░░ flat color ░░░░░ │      │ ▓▓▓▓ teal→…→white ▓▓▓▓ │  ← 620px gradient envelope
│ [hero content full-w] │      │      [hero content     │
│                         │      │       max-w-1280, ctr] │
├─────────────────────────┤      ├─────────────────────────┤
│ ▒▒ section 2 (works) ▒▒ │      │ ▒▒ section 2 (works) ▒▒ │
│   …                     │      │   …                     │
└─────────────────────────┘      └─────────────────────────┘
```

## Testing

- `npm run lint` clean.
- `npm run build` clean (TypeScript / Next.js).

Manual verification (dev server, both browsers):

1. **Safari 26.1, viewport ≥ 1440 px** — open `/`. Hero gradient renders teal→blue→violet→pink→white from top, fading out around 620 px. Hero content (badge, h1 lines, CTAs, demo card) is centered with visible whitespace on both sides. Compare side-by-side with Chrome at the same viewport — they match.
2. **Chrome, viewport ≥ 1440 px** — open `/`. Visually identical to before this change (regression check).
3. **Safari and Chrome, narrow viewport (≤ 768 px)** — hero content fills available width with the existing `px-5` / `md:px-10` padding (max-width is not yet hit). Gradient envelope still visible at top.
4. **Click the StickyHeader logo and nav links** — confirm they remain clickable (the `pointer-events-none` decorative layer must not block them).
5. **Scroll the page** — StickyHeader transitions to the white/blur state at scrollY > 40 (existing behavior). HeroGameDemo cycles through its 4 scenes (existing motion).
6. **Auth pages (`/auth/signin`, `/auth/signup`)** — render unchanged (regression check; we did not touch them).
7. **Purchase pages (`/purchase/membership`, `/purchase/beans`)** — render unchanged (regression check).

## Files Touched

- `dx-web/src/app/(web)/(home)/page.tsx` — wrapper className changes; one `<div aria-hidden>` added as the first child.

No other files change.
