# Hall Pages Responsive Breakpoints

**Date:** 2026-03-20
**Scope:** notices, unknown, review, mastered pages

## Goal

Align responsive breakpoints for 4 hall pages to match the completed games page pattern, targeting three breakpoints: mobile (<768px), tablet (768px-1024px), desktop (1024px+).

## Changes

### Page containers (4 files)

`notices/page.tsx`, `unknown/page.tsx`, `review/page.tsx`, `mastered/page.tsx`

- From: `flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7`
- To: `flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16`

### StatCard grid (3 content components)

`unknown-content.tsx`, `review-content.tsx`, `master-content.tsx`

- From: `grid grid-cols-2 gap-4 md:grid-cols-3`
- To: `grid grid-cols-2 gap-4 lg:grid-cols-3`

### Shared StatCard (`src/components/in/stat-card.tsx`)

- Padding: `md:px-5 md:py-4` -> `lg:px-5 lg:py-4`
- Font size: `md:text-xl` -> `lg:text-xl`

### NoticeItem (`notice-item.tsx`)

- Padding: `md:px-5` -> `lg:px-5`

## Constraints

- Pure CSS class updates only
- No logic, data fetching, or component structure changes
