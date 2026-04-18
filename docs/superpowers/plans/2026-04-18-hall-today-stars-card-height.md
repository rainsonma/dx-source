# Hall Today-Stars Card Height Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `TodayStarsCard` match `GameProgressCard`'s default height on desktop by fixing CSS Grid equalization and adding a flex-1 scrollable body structure.

**Architecture:** Three targeted CSS class changes across three files. No logic changes. Grid wrapper divs get `lg:h-full` to propagate the grid row track height to nested `h-full` cards. `TodayStarsCard` gains a `flex-1 min-h-0` body wrapper so its intrinsic height contribution to the grid row shrinks to header-only (~68px), letting `GameProgressCard` drive the row. `TodayStarsList` gets `lg:max-h-none lg:h-full` so it fills the list section on desktop while retaining its mobile 280px cap.

**Tech Stack:** Next.js 16, TailwindCSS v4, React

---

### Task 1: Add `lg:h-full` to column wrapper divs in `page.tsx`

**Files:**
- Modify: `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx:82-97`

- [ ] **Step 1: Apply the class changes**

  Open `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` and update the three column wrapper divs inside the main content row (lines 84, 89, 94):

  ```tsx
  {/* Main content row */}
  <div className="flex flex-col gap-5 lg:grid lg:grid-cols-3">
    {/* Right column - daily check-in (shows first on mobile) */}
    <div className="order-first lg:order-last lg:h-full">
      <DailyChallengeCard />
    </div>

    {/* Left column - game progress */}
    <div className="order-2 lg:order-first lg:h-full">
      <GameProgressCard sessions={data?.sessions ?? []} />
    </div>

    {/* Center column - today's stars */}
    <div className="order-3 lg:order-2 lg:h-full">
      <TodayStarsCard />
    </div>
  </div>
  ```

  Each wrapper div now has an explicit `height: 100%` on desktop, resolving to the grid row track size. This gives nested `h-full` cards a definite parent height to reference — required for cross-browser reliability.

- [ ] **Step 2: Run lint**

  ```bash
  cd dx-web && npm run lint
  ```

  Expected: no errors.

- [ ] **Step 3: Commit**

  ```bash
  git add dx-web/src/app/\(web\)/hall/\(main\)/\(home\)/page.tsx
  git commit -m "fix(web): add lg:h-full to hall column wrappers for grid height propagation"
  ```

---

### Task 2: Restructure `TodayStarsCard` with a `flex-1 min-h-0` body

**Files:**
- Modify: `dx-web/src/features/web/hall/components/today-stars-card.tsx`

- [ ] **Step 1: Replace the file content**

  Full new content for `dx-web/src/features/web/hall/components/today-stars-card.tsx`:

  ```tsx
  "use client";

  import Link from "next/link";
  import { ArrowRight, Loader2, Trophy } from "lucide-react";
  import { useTodayStars } from "@/features/web/hall/hooks/use-today-stars";
  import { TodayStarsPodium } from "./today-stars-podium";
  import { TodayStarsList } from "./today-stars-list";

  /** 今日明星榜 — Today's star leaderboard for the hall dashboard */
  export function TodayStarsCard() {
    const { data, isLoading } = useTodayStars();

    const podiumEntries = data.entries.slice(0, 3);
    const listEntries = data.entries.slice(3);

    return (
      <div className="flex h-full w-full flex-col gap-4 rounded-[14px] border border-border bg-card p-5">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-amber-50">
              <Trophy className="h-5 w-5 text-amber-500" />
            </div>
            <h3 className="text-base font-bold text-foreground">今日明星榜</h3>
          </div>
          <Link
            href="/hall/leaderboard"
            className="flex items-center gap-1 text-[13px] font-semibold text-teal-600 hover:text-teal-700"
          >
            查看全部
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        {/* Body */}
        {isLoading ? (
          <div className="flex flex-1 items-center justify-center">
            <Loader2 className="h-5 w-5 animate-spin text-teal-600" />
          </div>
        ) : data.entries.length === 0 ? (
          <div className="flex flex-1 items-center justify-center rounded-lg border border-border text-sm text-muted-foreground">
            暂无排名数据
          </div>
        ) : (
          <div className="flex min-h-0 flex-1 flex-col gap-4">
            {podiumEntries.length > 0 && (
              <TodayStarsPodium entries={podiumEntries} type="playtime" />
            )}
            {listEntries.length > 0 && (
              <div className="min-h-0 flex-1 overflow-hidden rounded-lg border border-border">
                <TodayStarsList entries={listEntries} type="playtime" />
              </div>
            )}
          </div>
        )}
      </div>
    );
  }
  ```

  Key changes from the original:
  - Loading state: removed `py-12`, added `flex-1` — fills body height, centres the spinner.
  - Empty state: removed `py-12`, added `flex-1` — fills body height, centres the text.
  - Data state: `<>` fragment replaced with `div.flex.min-h-0.flex-1.flex-col.gap-4` — the `min-h-0` allows the body to shrink for grid row sizing; `flex-1` fills allocated height.
  - List section: added `min-h-0 flex-1` to the `overflow-hidden rounded-lg border border-border` wrapper — absorbs remaining height after the podium.

- [ ] **Step 2: Run lint**

  ```bash
  cd dx-web && npm run lint
  ```

  Expected: no errors.

- [ ] **Step 3: Commit**

  ```bash
  git add dx-web/src/features/web/hall/components/today-stars-card.tsx
  git commit -m "fix(web): restructure TodayStarsCard body with flex-1 min-h-0 for height constraint"
  ```

---

### Task 3: Update `TodayStarsList` responsive max-height

**Files:**
- Modify: `dx-web/src/features/web/hall/components/today-stars-list.tsx:19`

- [ ] **Step 1: Apply the class change**

  In `dx-web/src/features/web/hall/components/today-stars-list.tsx`, update line 19 — the outer `<div>` of the returned JSX:

  ```tsx
  // Before
  <div className="max-h-[280px] overflow-y-auto">

  // After
  <div className="max-h-[280px] overflow-y-auto lg:max-h-none lg:h-full">
  ```

  No other changes. `max-h-[280px]` stays active on mobile (< `lg`). On desktop, `lg:max-h-none` removes the cap and `lg:h-full` fills the `flex-1 min-h-0` list-section container from Task 2, with `overflow-y-auto` handling the scroll.

- [ ] **Step 2: Run lint**

  ```bash
  cd dx-web && npm run lint
  ```

  Expected: no errors.

- [ ] **Step 3: Commit**

  ```bash
  git add dx-web/src/features/web/hall/components/today-stars-list.tsx
  git commit -m "fix(web): add lg:h-full lg:max-h-none to TodayStarsList for desktop height fill"
  ```

---

### Task 4: Build verification and visual check

**Files:** none (verification only)

- [ ] **Step 1: Production build**

  ```bash
  cd dx-web && npm run build
  ```

  Expected: exits 0 with no TypeScript or build errors.

- [ ] **Step 2: Start dev server and visually verify**

  ```bash
  cd dx-web && npm run dev
  ```

  Open `http://localhost:3000/hall` in a desktop-width browser (≥ 1024px) and verify:

  1. `今日明星榜` card height matches `我的学习进度` card height side by side.
  2. The podium (top 3 users) is visible and does not scroll — only the rank list (rank 4+) scrolls within the card.
  3. Loading state (briefly on first load): spinner is vertically centred in the card body.
  4. If you can force an empty state: "暂无排名数据" text is vertically centred in the card body.
  5. Resize to mobile width (< 1024px): `今日明星榜` card is content-driven height; rank list is capped at 280px with its own scroll.
  6. All other hall sections (stats row, heatmap, ad cards) are unchanged.
