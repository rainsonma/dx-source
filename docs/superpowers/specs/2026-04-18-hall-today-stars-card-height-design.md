# Design: Fix 今日明星榜 Card Height on /hall

**Date:** 2026-04-18  
**Scope:** `dx-web` frontend only — `/hall` dashboard page

---

## Problem

On the `/hall` dashboard, the 今日明星榜 (`TodayStarsCard`) grows taller than the 我的学习进度 (`GameProgressCard`) because its content (podium + rank list) exceeds `GameProgressCard`'s natural height. This drives the CSS Grid row to ~540px+, making the row taller than desired.

The root cause is two-fold:
1. The three column wrapper `<div>`s in `page.tsx` lack `h-full`, so the cards' `h-full` classes fall back to content-driven sizing rather than referencing the grid row track height.
2. `TodayStarsCard` has no flex body structure to constrain its intrinsic height contribution to the grid row.

---

## Desired Behaviour

- On desktop (`lg`): both cards occupy the same height — the natural height of `GameProgressCard` (~444px: header + `min-h-[288px]` content + pagination + `p-6` padding).
- `TodayStarsCard` body (podium + rank list) scrolls within that fixed height; the podium stays visually pinned at the top and only the rank list scrolls.
- On mobile: no change — cards remain content-driven height; rank list retains its existing `max-h-[280px]` scroll cap.

---

## Approach: CSS Grid Equalization + Flex-1 Body

No hard-coded heights. The grid row height is driven by `GameProgressCard`'s natural content; `TodayStarsCard` opts out of row-height contribution via `min-h-0` and fills the allocated height via `h-full`.

---

## Files Changed

### 1. `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx`

Add `lg:h-full` to all three column wrapper `<div>`s.

```tsx
// Before
<div className="order-first lg:order-last">
<div className="order-2 lg:order-first">
<div className="order-3 lg:order-2">

// After
<div className="order-first lg:order-last lg:h-full">
<div className="order-2 lg:order-first lg:h-full">
<div className="order-3 lg:order-2 lg:h-full">
```

This gives each card a definite parent height on desktop (= grid row track size), making nested `h-full` resolve reliably across all browsers.

---

### 2. `dx-web/src/features/web/hall/components/today-stars-card.tsx`

Insert a **body wrapper div** between the header and the existing content switch (loading / empty / data states). The outer card div is unchanged.

```
card div  (flex h-full flex-col gap-4 ... p-5)          ← unchanged
├── header div                                           ← unchanged
└── body div  (flex min-h-0 flex-1 flex-col gap-4)      ← NEW
      ├── loading  → div.flex.flex-1.items-center.justify-center
      ├── empty    → div.flex.flex-1.items-center.justify-center.rounded-lg.border…
      └── data     → React fragment
            ├── TodayStarsPodium   (unchanged)
            └── list section div  (min-h-0 flex-1 overflow-hidden rounded-lg border border-border)
                  └── TodayStarsList
```

Key properties:
- `flex-1 min-h-0` on the body div: fills all card height after the header; `min-h-0` allows the body to shrink to zero for grid track sizing, so `TodayStarsCard`'s intrinsic size contribution equals just its header (~68px).
- `flex-1 min-h-0 overflow-hidden` on the list section: fills body height after the podium; clips any overflow so `TodayStarsList`'s scroll bar is self-contained.
- Loading/empty states lose `py-12` (replaced by `flex-1 items-center justify-center`).

---

### 3. `dx-web/src/features/web/hall/components/today-stars-list.tsx`

Single-line change on the outer `<div>`:

```tsx
// Before
<div className="max-h-[280px] overflow-y-auto">

// After
<div className="max-h-[280px] overflow-y-auto lg:max-h-none lg:h-full">
```

- **Mobile**: retains `max-h-[280px]` cap and internal scroll (existing behaviour).
- **Desktop**: `lg:max-h-none` removes the cap; `lg:h-full` fills the `flex-1 min-h-0` list-section container; `overflow-y-auto` handles scroll. No nested scrollbars.

---

## Height Flow (Desktop)

```
grid row track height
  = max(GameProgressCard ~444px, DailyChallengeCard ~374px, TodayStarsCard intrinsic ~68px)
  = ~444px

wrapper div  (lg:h-full = 100% of grid area = 444px)
  └── TodayStarsCard  (h-full = 444px)
        ├── header  (~52px incl. gap)
        └── body  (flex-1 = ~352px)
              ├── TodayStarsPodium  (~150px)
              └── list section  (flex-1 min-h-0 = ~186px, scrollable)
```

---

## Non-Goals

- No changes to `GameProgressCard`, `DailyChallengeCard`, or any other component.
- No changes to the full leaderboard page.
- No mobile layout changes.

---

## Test Checklist

- [ ] Desktop: `TodayStarsCard` height matches `GameProgressCard` height when both have data.
- [ ] Desktop: rank list scrolls within the card; podium stays pinned.
- [ ] Desktop: loading and empty states are vertically centred within the card.
- [ ] Desktop: `DailyChallengeCard` height unchanged (stretches to grid row height as before).
- [ ] Mobile: cards are content-driven; rank list is capped at 280px with scroll.
- [ ] No lint errors.
- [ ] No regressions on other hall page sections.
