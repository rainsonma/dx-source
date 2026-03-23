# Hall Dashboard Responsive Design

## Overview

Add responsive design to the hall dashboard page (`src/app/(web)/hall/(main)/(home)/page.tsx`) and its child components. The page currently has no responsive breakpoints and is desktop-only.

## Breakpoints

| Name    | Range        | Tailwind prefix |
|---------|-------------|-----------------|
| Mobile  | <768px      | (default)       |
| Tablet  | 768–1024px  | md:             |
| Desktop | 1024px+     | lg:             |

The shell (`HallMainShell`) already handles sidebar/mobile-header toggle at `md:`. Our layout changes use `lg:` for the desktop switch.

## Changes by Component

### 1. Page container (`page.tsx`)

- Padding: `px-4 pt-5 pb-12 lg:px-8 lg:pt-7 lg:pb-16`
- Gap: `gap-5 lg:gap-6`
- Main content row: stack vertically by default, side-by-side on `lg:`
- Right column: full width by default, fixed `w-80` on `lg:`

### 2. AdCardsRow (`ad-cards-row.tsx`)

- Stack vertically by default, horizontal on `lg:`
- `flex w-full flex-col gap-4 lg:flex-row`

### 3. StatsRow (`stats-row.tsx`)

- 2×2 grid at all sizes
- `grid w-full grid-cols-2 gap-4`

### 4. LearningHeatmap (`learning-heatmap.tsx`)

- Body: stack vertically by default, side-by-side on `lg:`
- Right sidebar: full width by default, fixed `w-44` on `lg:`
- Heatmap grid wrapper: `overflow-x-auto` for horizontal scroll on small screens

### 5. No changes

- GreetingTopBar — already works with flex justify-between
- TopActions — already has its own responsive hiding at `lg:` and `xl:`
- GameProgressCard — full width in all layouts
- DailyChallengeCard — adapts to container width

## Constraints

- No new files or components
- No structural/logic changes — CSS classes only
- Must not break existing functionality
- Follow existing responsive patterns (e.g., leaderboard page uses `px-4 py-7 md:px-8`)
