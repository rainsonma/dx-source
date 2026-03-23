# Responsive Breakpoints Design

**Date:** 2026-03-20
**Scope:** Dashboard stats row + all games pages under `hall/(main)/games/`

## Breakpoints

| Name | Tailwind | Width |
|---|---|---|
| Mobile | (default) | <768px |
| Tablet | `md:` | 768–1024px |
| Desktop | `lg:` | 1024px+ |

These align with the existing dashboard page patterns (`lg:px-8`, `lg:pt-7`, etc.).

## Changes

### 1. `stats-row.tsx` — Dashboard stats cards

| Breakpoint | Layout |
|---|---|
| Mobile/Tablet | 2×2 grid (`grid-cols-2`) |
| Desktop | 4 in a row (`lg:grid-cols-4`) |

### 2. `games/page.tsx` — Games list page

Add responsive padding matching dashboard pattern:
`px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16`

### 3. `games-page-content.tsx` — Game cards grid

| Breakpoint | Columns |
|---|---|
| Mobile | 2 (`grid-cols-2`) |
| Tablet | 3 (`md:grid-cols-3`) |
| Desktop | 5 (`lg:grid-cols-5`) |

### 4. `games/mine/page.tsx` — My Games page

- Responsive padding (same as games list)
- Game cards grid: `grid-cols-2 md:grid-cols-3 lg:grid-cols-5`

### 5. `games/[id]/page.tsx` — Game detail page

Responsive padding (same pattern).

### 6. `hero-card.tsx` — Game detail hero

| Breakpoint | Layout |
|---|---|
| Mobile/Tablet | Vertical stack, cover full-width with reduced height |
| Desktop | Horizontal side-by-side, 280×280 cover |

### 7. `game-detail-content.tsx` — Level grid + sidebar

| Breakpoint | Layout |
|---|---|
| Mobile/Tablet | Stacked vertically, sidebar full-width below |
| Desktop | Side-by-side, sidebar `w-80` |

### 8. `level-grid.tsx` — Level cells grid

| Breakpoint | Columns |
|---|---|
| Mobile | 4 |
| Tablet | 7 |
| Desktop | 10 |

### 9. `filter-section.tsx` — Filter pills

Add `flex-wrap` to all filter rows so pills wrap on narrow screens.

## Principles

- Mobile-first: default styles target mobile, add `md:` and `lg:` overrides
- No functional changes — only layout/sizing adjustments
- Consistent padding pattern across all games pages
