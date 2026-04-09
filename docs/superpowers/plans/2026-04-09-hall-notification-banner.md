# Hall Notification Banner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a rotating notification banner above the main 3-card row on `/hall` that cycles through the latest 3 notices, opens a ShadCN dialog on click, and move the existing `StatsRow` down above `LearningHeatmap`.

**Architecture:** Two new React client components (banner container + presentational dialog) in `features/web/hall/components/`. Banner fetches `/api/notices?limit=3` on mount, auto-rotates every 5 seconds with pause-on-hover, opens a controlled Dialog on click. Integrate by editing the hall home page. No backend, database, or deploy changes.

**Tech Stack:** Next.js 16 (App Router, React 19), TypeScript strict, Tailwind CSS 4, `tw-animate-css`, `radix-ui` (via shadcn Dialog), `lucide-react` icons, `apiClient` helper.

**Spec:** `docs/superpowers/specs/2026-04-09-hall-notification-banner-design.md`

**Testing note:** dx-web has no React unit test infrastructure (no vitest/jest in `package.json`). Each task's "verification" is `npm run lint` + `npm run build`, with final manual browser testing in Task 4 against the spec's 15-point manual test plan.

---

### File Map

| Action | File | Purpose |
|--------|------|---------|
| Create | `dx-web/src/features/web/hall/components/notification-banner-dialog.tsx` | Presentational ShadCN dialog displaying a single notice (header / scrollable body / bottom-right footer) |
| Create | `dx-web/src/features/web/hall/components/notification-banner.tsx` | Container: fetches notices, manages rotation state, renders ticker, opens dialog |
| Modify | `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` | Add `NotificationBanner` import; insert `<NotificationBanner />` after `<AdCardsRow />`; move `<StatsRow />` to sit directly above `LearningHeatmap` |

No other files touched. dx-api is unchanged.

---

### Task 1: Create `NotificationBannerDialog` (leaf component, no deps)

**Files:**
- Create: `dx-web/src/features/web/hall/components/notification-banner-dialog.tsx`

- [ ] **Step 1: Write the full component file**

Create `dx-web/src/features/web/hall/components/notification-banner-dialog.tsx` with this exact content:

```tsx
"use client";

import { createElement } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { NoticeItem } from "@/features/web/notice/actions/notice.action";
import { resolveNoticeIcon } from "@/features/web/notice/helpers/notice-icon";
import { formatRelativeTime } from "@/features/web/notice/helpers/notice-time";

interface NotificationBannerDialogProps {
  notice: NoticeItem | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/** ShadCN dialog showing a single notice, opened from NotificationBanner */
export function NotificationBannerDialog({
  notice,
  open,
  onOpenChange,
}: NotificationBannerDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="gap-0 overflow-hidden p-0 sm:max-w-lg">
        <DialogHeader className="border-b border-border bg-slate-50/60 px-6 py-4">
          <DialogTitle className="text-base font-bold text-foreground">
            斗学消息通知
          </DialogTitle>
          <DialogDescription className="sr-only">
            查看消息通知内容
          </DialogDescription>
        </DialogHeader>

        <div className="flex max-h-[60vh] flex-col gap-3 overflow-y-auto px-6 py-5">
          {notice && (
            <>
              <div className="flex items-start gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[10px] bg-teal-50">
                  {createElement(resolveNoticeIcon(notice.icon), {
                    className: "h-[18px] w-[18px] text-teal-600",
                  })}
                </div>
                <h3 className="pt-1.5 text-[15px] font-semibold text-foreground">
                  {notice.title}
                </h3>
              </div>
              {notice.content && (
                <p className="text-sm leading-relaxed whitespace-pre-wrap text-muted-foreground">
                  {notice.content}
                </p>
              )}
            </>
          )}
        </div>

        <div className="flex items-center justify-end border-t border-border px-6 py-3">
          <span className="text-xs text-muted-foreground">
            {notice ? formatRelativeTime(notice.createdAt) : ""}
          </span>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Run ESLint on the new file**

Run from repo root:
```bash
cd dx-web && npm run lint
```
Expected: exits 0 with no errors or warnings pointing at the new file. The existing codebase's lint status is the baseline — this step must not introduce any new warnings.

- [ ] **Step 3: Run Next.js build to verify types**

Run:
```bash
cd dx-web && npm run build
```
Expected: build succeeds ("Compiled successfully"), zero TypeScript errors. The build must reach the "Linting and checking validity of types" phase and pass. If it fails with "Cannot find module '@/features/web/notice/helpers/notice-icon'" or similar, check the file paths in the imports.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/hall/components/notification-banner-dialog.tsx
git commit -m "$(cat <<'EOF'
feat(hall): add NotificationBannerDialog component

Presentational shadcn Dialog for displaying a single notice. Header
shows the constant title 斗学消息通知, body shows the notice icon +
title + content (scrollable if long), footer pins the relative timestamp
bottom-right. Accepts a nullable notice prop so the last-displayed notice
stays visible during the dialog's exit animation without flicker.

Refs: docs/superpowers/specs/2026-04-09-hall-notification-banner-design.md
EOF
)"
```

---

### Task 2: Create `NotificationBanner` (container)

**Files:**
- Create: `dx-web/src/features/web/hall/components/notification-banner.tsx`

- [ ] **Step 1: Write the full component file**

Create `dx-web/src/features/web/hall/components/notification-banner.tsx` with this exact content:

```tsx
"use client";

import { createElement, useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import type { NoticeItem } from "@/features/web/notice/actions/notice.action";
import { resolveNoticeIcon } from "@/features/web/notice/helpers/notice-icon";
import { formatRelativeTime } from "@/features/web/notice/helpers/notice-time";
import { NotificationBannerDialog } from "./notification-banner-dialog";

const ROTATION_INTERVAL_MS = 5000;

type NoticeListResponse = {
  items: NoticeItem[];
  nextCursor: string;
  hasMore: boolean;
};

/** Rotating notification ticker above the /hall 3-card row */
export function NotificationBanner() {
  const [notices, setNotices] = useState<NoticeItem[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [hovered, setHovered] = useState(false);
  const [dialogNotice, setDialogNotice] = useState<NoticeItem | null>(null);

  const paused = hovered || dialogNotice !== null;

  // Fetch latest 3 notices once on mount. Silent failure: non-zero code or
  // thrown error leaves `notices` empty, which causes the render to return null.
  useEffect(() => {
    let cancelled = false;
    apiClient
      .get<NoticeListResponse>("/api/notices?limit=3")
      .then((res) => {
        if (cancelled) return;
        if (res.code === 0) setNotices(res.data.items ?? []);
        setLoaded(true);
      })
      .catch(() => {
        if (!cancelled) setLoaded(true);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // Auto-rotate every ROTATION_INTERVAL_MS while unpaused and more than 1 notice.
  useEffect(() => {
    if (notices.length < 2 || paused) return;
    const id = setInterval(() => {
      setCurrentIndex((i) => (i + 1) % notices.length);
    }, ROTATION_INTERVAL_MS);
    return () => clearInterval(id);
  }, [notices.length, paused]);

  if (!loaded || notices.length === 0) return null;

  const current = notices[currentIndex];
  const Icon = resolveNoticeIcon(current.icon);

  return (
    <>
      <button
        type="button"
        onClick={() => setDialogNotice(current)}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
        onFocus={() => setHovered(true)}
        onBlur={() => setHovered(false)}
        className="group flex w-full items-center gap-3 rounded-[14px] border border-border bg-card px-4 py-2.5 text-left transition-colors hover:border-teal-200 hover:bg-teal-50/30 focus-visible:ring-2 focus-visible:ring-teal-500/50 focus-visible:ring-offset-2 focus-visible:outline-hidden sm:px-5"
        aria-label={`查看消息通知:${current.title}`}
      >
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] bg-teal-50 transition-colors group-hover:bg-teal-100">
          {createElement(Icon, {
            className: "h-[18px] w-[18px] text-teal-600",
          })}
        </div>

        <div
          key={currentIndex}
          className="animate-in fade-in-0 slide-in-from-bottom-1 flex min-w-0 flex-1 items-center gap-2.5 duration-300 sm:gap-3"
        >
          <span className="shrink-0 text-sm font-semibold text-foreground">
            {current.title}
          </span>
          {current.content && (
            <span className="hidden truncate text-sm text-muted-foreground sm:inline">
              {current.content}
            </span>
          )}
        </div>

        {notices.length > 1 && (
          <div className="hidden shrink-0 items-center gap-1 lg:flex">
            {notices.map((n, i) => (
              <span
                key={n.id}
                aria-hidden="true"
                className={cn(
                  "h-1.5 rounded-full transition-all",
                  i === currentIndex ? "w-3.5 bg-teal-500" : "w-1.5 bg-slate-300",
                )}
              />
            ))}
          </div>
        )}

        <span className="shrink-0 text-xs text-muted-foreground">
          {formatRelativeTime(current.createdAt)}
        </span>
      </button>

      <NotificationBannerDialog
        notice={dialogNotice}
        open={dialogNotice !== null}
        onOpenChange={(open) => {
          if (!open) setDialogNotice(null);
        }}
      />
    </>
  );
}
```

- [ ] **Step 2: Run ESLint**

Run:
```bash
cd dx-web && npm run lint
```
Expected: exits 0. The most likely warning categories to watch for:
- `react-hooks/exhaustive-deps` — the empty dep array on the fetch effect is correct (mount-once); if lint flags it, add `// eslint-disable-next-line react-hooks/exhaustive-deps` on the closing `}, []);` line with a comment explaining why
- `@typescript-eslint/no-unused-vars` — confirm `createElement` is actually used (it is, inside the icon render)

- [ ] **Step 3: Run Next.js build**

Run:
```bash
cd dx-web && npm run build
```
Expected: build succeeds, zero TypeScript errors. Verify the new file is included in the compiled output (check the build log for "Compiled successfully").

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/hall/components/notification-banner.tsx
git commit -m "$(cat <<'EOF'
feat(hall): add NotificationBanner rotating ticker

Client component that fetches the latest 3 notices from /api/notices?limit=3,
auto-rotates every 5 seconds with slide-in animation, pauses on hover or
while the dialog is open, and opens NotificationBannerDialog on click.
Silent-hides on empty/error state so it never disrupts page layout.

Renders inside a real <button type="button"> for free keyboard a11y.
Responsive: mobile shows icon + title + time; tablet adds content snippet;
desktop adds dot indicators.

Refs: docs/superpowers/specs/2026-04-09-hall-notification-banner-design.md
EOF
)"
```

---

### Task 3: Integrate into `/hall` page (move StatsRow + mount NotificationBanner)

**Files:**
- Modify: `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx`

- [ ] **Step 1: Add the import**

In `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx`, add one import line after the `LearningHeatmap` import (around line 11):

**Before:**
```tsx
import { LearningHeatmap } from "@/features/web/hall/components/learning-heatmap";
```

**After:**
```tsx
import { LearningHeatmap } from "@/features/web/hall/components/learning-heatmap";
import { NotificationBanner } from "@/features/web/hall/components/notification-banner";
```

- [ ] **Step 2: Replace the JSX return block**

Find the JSX `return ( ... );` block (starts around line 61, ends around line 103). Replace the entire return statement with this exact content:

```tsx
  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <GreetingTopBar
        title={`早上好，${displayName} 👋`}
        subtitle="继续你的学习之旅，今天也要加油！"
      />
      <AdCardsRow />
      <NotificationBanner />

      {/* Main content row */}
      <div className="flex flex-col gap-5 lg:grid lg:grid-cols-3">
        {/* Right column - daily check-in (shows first on mobile) */}
        <div className="order-first lg:order-last">
          <DailyChallengeCard />
        </div>

        {/* Left column - game progress */}
        <div className="order-2 lg:order-first">
          <GameProgressCard sessions={data?.sessions ?? []} />
        </div>

        {/* Center column - today's stars */}
        <div className="order-3 lg:order-2">
          <TodayStarsCard />
        </div>
      </div>

      <StatsRow
        exp={data?.profile.exp ?? 0}
        currentPlayStreak={data?.profile.currentPlayStreak ?? 0}
        masteredTotal={data?.masterStats.total ?? 0}
        masteredThisWeek={data?.masterStats.thisWeek ?? 0}
        reviewPending={data?.reviewStats.pending ?? 0}
      />

      {/* Learning heatmap */}
      {heatmap && (
        <LearningHeatmap
          initialYear={heatmap.year}
          initialDays={heatmap.days}
          accountYear={heatmap.accountYear}
        />
      )}
    </div>
  );
```

**Note:** The Chinese strings in the block above use full-width punctuation (`，` U+FF0C and `！` U+FF01) matching the existing file exactly. Copy-paste the block verbatim — do not substitute half-width `,` or `!`, which would change the rendered greeting text.

The three structural changes compared to the original:
1. **Added:** `<NotificationBanner />` immediately after `<AdCardsRow />`
2. **Removed:** `<StatsRow .../>` from its original position (between `<AdCardsRow />` and the main content row)
3. **Added:** `<StatsRow .../>` below the main content row, immediately above the `{heatmap && ...}` block

- [ ] **Step 3: Verify the file**

Re-read the modified file and confirm:
- Line count is ~103 (original was 104; net change: +1 import line, +1 `<NotificationBanner />` line, StatsRow's 7 lines moved not added = net +2)
- `NotificationBanner` import is present and unused-import warnings are absent
- `StatsRow` import is still present (it's still used, just relocated)
- `<NotificationBanner />` appears exactly once
- `<StatsRow .../>` appears exactly once
- The order is: GreetingTopBar → AdCardsRow → NotificationBanner → main content grid → StatsRow → LearningHeatmap
- Chinese strings preserved with full-width punctuation

- [ ] **Step 4: Run ESLint**

Run:
```bash
cd dx-web && npm run lint
```
Expected: exits 0 with no errors or warnings on `page.tsx`.

- [ ] **Step 5: Run Next.js build**

Run:
```bash
cd dx-web && npm run build
```
Expected: build succeeds, zero TypeScript errors. Pay attention to the build output — if `NotificationBanner` or `NotificationBannerDialog` are listed as a "new" chunk, that confirms they're wired in.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/app/\(web\)/hall/\(main\)/\(home\)/page.tsx
git commit -m "$(cat <<'EOF'
feat(hall): mount NotificationBanner and relocate StatsRow

Insert <NotificationBanner /> between AdCardsRow and the main 3-card row
on the hall dashboard. Move <StatsRow /> from its previous position
(directly below AdCardsRow) to sit directly above LearningHeatmap.

No prop changes to StatsRow. Vertical rhythm is preserved by the parent
container's gap-5 lg:gap-6 classes, so no spacing tweaks are needed.

Refs: docs/superpowers/specs/2026-04-09-hall-notification-banner-design.md
EOF
)"
```

---

### Task 4: Manual browser verification

No code changes in this task — this is a verification checkpoint that exercises all 15 test cases from the spec against the running dev server.

**Files:** (none modified)

- [ ] **Step 1: Start the dev server**

Run in a dedicated terminal:
```bash
cd dx-web && npm run dev
```
Expected: "Ready in ..." followed by "Local: http://localhost:3000". Keep this running for the remaining steps.

- [ ] **Step 2: Start the Go backend**

In a second terminal:
```bash
cd dx-api && go run .
```
Expected: server listens on `:3001`. Keep this running.

- [ ] **Step 3: Ensure test notices exist**

Option A — already have 3+ active notices: proceed to Step 4.

Option B — need to seed notices: log in as the admin user, navigate to `/hall/notices`, click the publish button, and create 3 notices with distinct titles like:
1. "系统升级公告" / "新版本上线,新增 AI 对战模式" / icon: `megaphone`
2. "活动提醒" / "本周末有积分双倍活动,千万不要错过" / icon: `gift`
3. "使用小贴士" / "每天坚持打卡,连续学习奖励更丰厚" / icon: `sparkles`

- [ ] **Step 4: Execute the 15-point manual test plan**

Open http://localhost:3000, sign in, navigate to `/hall`. Verify each case passes by checking the box:

- [ ] **4.1** Banner appears between `AdCardsRow` and the 3-card row
- [ ] **4.2** Banner rotates every 5 seconds between the 3 notices
- [ ] **4.3** Dot indicators animate (active dot elongates to `w-3.5`, inactive stays `w-1.5`)
- [ ] **4.4** Hover on banner pauses rotation; mouse-leave resumes it
- [ ] **4.5** Click the banner opens dialog with header `斗学消息通知`
- [ ] **4.6** Dialog body shows notice icon + title + content matching the clicked notice
- [ ] **4.7** Dialog footer shows relative timestamp pinned bottom-right
- [ ] **4.8** Close dialog via the X button, overlay click, or ESC → rotation resumes from the displayed index
- [ ] **4.9** In admin, soft-delete all notices (set `is_active=false`) → refresh `/hall` → banner is absent, rest of page renders normally (no gap, no error)
- [ ] **4.10** Create exactly 1 active notice → banner shows it with no dots and no rotation
- [ ] **4.11** Create exactly 2 active notices → banner rotates between them with 2 dots visible at `lg+` viewport
- [ ] **4.12** Resize to mobile viewport (`< 640px`, ~375px wide) → content snippet hidden, dots hidden, icon + title + timestamp visible, layout stays clean
- [ ] **4.13** Resize to tablet viewport (640–1023px, ~768px wide) → content snippet visible, dots hidden
- [ ] **4.14** Resize to desktop viewport (`≥ 1024px`, ~1440px wide) → all elements visible
- [ ] **4.15** Simulate API error: open DevTools → Network → right-click `/api/notices` → "Block request URL" → refresh `/hall` → banner is absent, rest of hall page renders normally

Additional checks not in the spec's 15 but needed for final sign-off:

- [ ] **4.16** Verify `StatsRow` now sits directly above `LearningHeatmap`, with the same `gap-5 lg:gap-6` spacing as other sections
- [ ] **4.17** Verify no console errors or warnings. Specifically:
  - No radix-ui warning about missing `DialogDescription` (we provide `sr-only`)
  - No React key warnings from the dot map (we use `n.id` as key)
  - No hydration warnings
- [ ] **4.18** Verify Tab key navigation: Tab onto the banner → focus-visible ring appears (`ring-2 ring-teal-500/50`); press Enter → dialog opens; Tab inside dialog → closes focus ring loops inside dialog
- [ ] **4.19** Create a notice with > 500 characters of content → click to open → dialog body scrolls internally while footer and header stay fixed
- [ ] **4.20** Verify the spec's "no breaking changes" guarantee: click through all other hall sections (GameProgressCard has navigation, DailyChallengeCard has a button, TodayStarsCard has "查看全部", LearningHeatmap renders) → all work unchanged

- [ ] **Step 5: Document any defects**

If any check above fails, stop and report the failing check number with a description. Do NOT proceed to Task 5 until all checks pass. Defects should be logged as follow-up tasks, not ignored.

- [ ] **Step 6: No commit for this task**

Manual verification doesn't produce code changes. Move to Task 5.

---

### Task 5: Final static checks

A last full-pass to confirm the repo is PR-ready.

**Files:** (none modified)

- [ ] **Step 1: Run ESLint across the whole project**

```bash
cd dx-web && npm run lint
```
Expected: exits 0 with no errors or warnings. The lint output should be empty or only show "✔ No ESLint warnings or errors".

- [ ] **Step 2: Run the full production build**

```bash
cd dx-web && npm run build
```
Expected: build completes, shows "Compiled successfully", prints the route map, prints the static pages summary. No TypeScript errors.

- [ ] **Step 3: Review the git log for the 3 commits**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git log --oneline -5
```
Expected: the top 3 commits should be (newest first):
1. `feat(hall): mount NotificationBanner and relocate StatsRow`
2. `feat(hall): add NotificationBanner rotating ticker`
3. `feat(hall): add NotificationBannerDialog component`

- [ ] **Step 4: Review the cumulative diff**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git diff HEAD~3 HEAD --stat
```
Expected: shows 3 files changed:
- `dx-web/src/app/(web)/hall/(main)/(home)/page.tsx` — ~10 insertions, ~7 deletions (net +3 lines)
- `dx-web/src/features/web/hall/components/notification-banner.tsx` — ~130 insertions, 0 deletions
- `dx-web/src/features/web/hall/components/notification-banner-dialog.tsx` — ~65 insertions, 0 deletions

- [ ] **Step 5: Report completion**

Announce to the user: "Implementation complete. All 3 commits pushed to the current branch, all lint/build/manual checks pass. Ready for PR or manual review."

---

## Plan Self-Review

### Spec coverage

Walking through each spec section and confirming a task implements it:

| Spec section | Covered by |
|---|---|
| Summary (banner + StatsRow move) | Tasks 1–3 |
| Current State (reference only) | n/a |
| New Layout (vertical order change) | Task 3 |
| Visual Design (compact ticker, animation, interaction) | Task 2 |
| Visual Design (dialog structure) | Task 1 |
| Component Architecture — `notification-banner-dialog.tsx` | Task 1 |
| Component Architecture — `notification-banner.tsx` | Task 2 |
| Responsive behavior table | Task 2 (class usage) + Task 4 (manual test 4.12–4.14) |
| Data Strategy (fetch once, silent error, no mark-as-read) | Task 2 (fetch effect) + Task 4.15 (error test) |
| Edge Cases table | Task 2 (render decision tree) + Task 4.9–4.11, 4.15 (manual coverage) |
| File Changes — New Files | Tasks 1, 2 |
| File Changes — Modified Files | Task 3 |
| Reused imports | Tasks 1, 2 (explicit import statements) |
| No Backend Changes | n/a — enforced by plan scope |
| No Deploy Changes | n/a — enforced by plan scope |
| Styling (teal accents, card treatment, muted text) | Tasks 1, 2 (class strings) |
| Non-Goals | n/a — enforced by plan scope |
| Constraints (lint, types, no breaking changes) | Tasks 1.2, 1.3, 2.2, 2.3, 3.4, 3.5, 5.1, 5.2 |
| Verification — static checks | Task 5 |
| Verification — manual test plan (15 points) | Task 4 (all 15 covered as 4.1–4.15, plus 4.16–4.20 additions) |

No gaps.

### Placeholder scan

Searched the plan for red-flag patterns — none present:
- No "TBD", "TODO", "implement later", "fill in details"
- No "Add appropriate error handling" — error handling is specific (silent fallback to `setLoaded(true)`)
- No "Write tests for the above" — testing approach is explicit (lint + build + manual)
- No "Similar to Task N" — each task repeats its code verbatim
- No vague steps — every code step shows exact code, every command step shows the exact command

### Type consistency

- `NoticeItem` type imported identically in Tasks 1 and 2 from `@/features/web/notice/actions/notice.action`
- `NotificationBannerDialogProps` interface defined in Task 1, consumed by Task 2's `<NotificationBannerDialog notice={dialogNotice} open={...} onOpenChange={...} />` — prop names match exactly
- `resolveNoticeIcon` return type is `LucideIcon` (verified at `features/web/notice/helpers/notice-icon.ts:50`), passed to `createElement` in both tasks
- `formatRelativeTime` signature `(date: Date | string) => string` (verified at `features/web/notice/helpers/notice-time.ts:2`) — called with `notice.createdAt` (string) in both tasks
- Hook state types: `useState<NoticeItem[]>`, `useState<boolean>`, `useState<number>`, `useState<NoticeItem | null>` — all consistent with usage
- `apiClient.get<NoticeListResponse>` typed response — `res.code` and `res.data.items` access matches the envelope

No type drift.

### Execution notes

- Tasks 1 and 2 could theoretically run in parallel since Task 1 has no deps on Task 2, but Task 2 imports from Task 1. Run them sequentially in order.
- Task 4 requires both dev servers running and admin access to seed data; if the executor is a non-interactive subagent, it may need to stub the data or skip Task 4 and hand off to the user.
- Task 5 is a no-code checkpoint; it confirms the previous 3 tasks produced a clean, committed state.
