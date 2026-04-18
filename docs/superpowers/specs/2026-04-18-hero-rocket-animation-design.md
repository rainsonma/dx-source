# Design: Hero Button Rocket Icon + Float Animation

**Date:** 2026-04-18  
**Scope:** `dx-web` frontend only — home landing page hero section

---

## Change

Replace the static `Gamepad2` icon on the 开启斗学之旅 button in `hero-section.tsx` with an animated `Rocket` icon matching the float animation on the 开始你的斗学冒险 button in `final-cta-section.tsx`.

---

## File Changed

**`dx-web/src/features/web/home/components/hero-section.tsx`**

| | Before | After |
|---|---|---|
| Directive | _(server component)_ | `"use client"` |
| Icon import | `Gamepad2` | `Rocket` |
| New imports | — | `motion` from `motion/react`; `usePrefersReducedMotion` from `@/features/web/home/hooks/use-in-view` |
| Hook | — | `const reduced = usePrefersReducedMotion()` |
| Icon JSX | `<Gamepad2 className="h-5 w-5 text-white" />` | `<motion.span animate … ><Rocket className="h-5 w-5 text-white" /></motion.span>` |

**Animation spec** (identical to `FinalCtaSection`):

```tsx
<motion.span
  animate={reduced ? undefined : { y: [0, -2, 0] }}
  transition={reduced ? undefined : { duration: 1.6, repeat: Infinity, ease: "easeInOut" }}
  className="flex items-center"
>
  <Rocket className="h-5 w-5 text-white" />
</motion.span>
```

- Gentle 2px vertical float, 1.6 s loop, `easeInOut` curve.
- `usePrefersReducedMotion` disables animation when the OS accessibility setting is on.
- Icon size `h-5 w-5` is unchanged from the original `Gamepad2`.

---

## Non-Goals

- No changes to button styling, text, or href.
- No changes to `final-cta-section.tsx` or any other file.
