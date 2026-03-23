# AI Custom Pages Responsive Breakpoints

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Align AI custom pages responsive breakpoints with the games pages pattern (mobile <768px, tablet 768-1024px, desktop 1024px+).

**Architecture:** CSS-only changes — shift major layout breakpoints from `md:` to `lg:`, align page wrappers and grid columns with the established games pattern. Dialog internals remain unchanged.

**Tech Stack:** TailwindCSS v4, Next.js App Router

---

## Reference Pattern (from games pages)

- Page wrapper: `flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16`
- Grid: `grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5`
- Hero inner: `flex w-full flex-col gap-5 p-4 lg:flex-row lg:gap-7 lg:p-6`
- Hero image: `lg:h-[280px] lg:w-[280px]`
- Detail layout: `flex flex-1 flex-col gap-5 lg:flex-row`
- Sidebar: `lg:w-80 lg:shrink-0 lg:p-6`

---

### Task 1: Page Wrappers (3 files)

**Files:**
- Modify: `src/app/(web)/hall/(main)/ai-custom/page.tsx:24`
- Modify: `src/app/(web)/hall/(main)/ai-custom/[id]/page.tsx:11`
- Modify: `src/app/(web)/hall/(main)/ai-custom/[id]/[levelId]/page.tsx:27`

Change all three from:
```
flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7
```
To:
```
flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16
```

### Task 2: AI Custom Grid

**File:** `src/features/web/ai-custom/components/ai-custom-grid.tsx`

- Banner (line 58): `p-5 pb-6 md:p-7 md:pb-8` → `p-5 pb-6 lg:p-7 lg:pb-8`
- Banner title (line 63): `md:text-[28px]` → `lg:text-[28px]`
- Game grid (line 101): `grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5` → `grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5`

### Task 3: Game Hero Card

**File:** `src/features/web/ai-custom/components/game-hero-card.tsx`

- Inner flex (line 116): `flex flex-col items-center gap-7 p-4 md:flex-row md:p-6` → `flex w-full flex-col gap-5 p-4 lg:flex-row lg:gap-7 lg:p-6`
- Image with cover (line 122): `md:h-[280px] md:w-[280px]` → `lg:h-[280px] lg:w-[280px]`
- Image fallback (line 126): `md:h-[280px] md:w-[280px]` → `lg:h-[280px] lg:w-[280px]`
- Title (line 137): `md:text-2xl` → `lg:text-2xl`

### Task 4: Game Info Card

**File:** `src/features/web/ai-custom/components/game-info-card.tsx`

- Container (line 58): `p-4 md:p-6 lg:w-80` → `p-4 lg:w-80 lg:p-6`

### Task 5: Game Levels Card

**File:** `src/features/web/ai-custom/components/game-levels-card.tsx`

- Container (line 66): `p-4 md:p-6` → `p-4 lg:p-6`

### Task 6: Level Units Panel

**File:** `src/features/web/ai-custom/components/level-units-panel.tsx`

- Two-column layout (line 528): `md:flex-row` → `lg:flex-row`
- Left panel padding (line 548): `p-4 pb-0 md:p-5 md:pb-0` → `p-4 pb-0 lg:p-5 lg:pb-0`
- Left scroll area (line 608): `px-4 pb-4 md:px-5 md:pb-5` → `px-4 pb-4 lg:px-5 lg:pb-5`
- Right panel (line 649): `p-4 md:p-5` → `p-4 lg:p-5`

### Task 7: Sortable Content Item

**File:** `src/features/web/ai-custom/components/sortable-content-item.tsx`

- Title bar (line 100): `px-2 pt-2 md:px-2` → `px-2 pt-2` (remove no-op)
- Content body (line 174): `px-3 py-2.5 md:px-4` → `px-3 py-2.5 lg:px-4`

### Task 8: Verify

Run `npm run build` to confirm no breakage.

### NOT Changed

- CreateCourseForm — dialog internal padding (`md:px-6`, `md:px-7`)
- EditGameDialog — dialog context
- AddLevelDialog — dialog context
- AddMetadataDialog — dialog context
- GameCardItem — no responsive classes to change
- SortableMetaItem — no responsive classes to change
- ProcessingOverlay — fixed overlay, no responsive classes
