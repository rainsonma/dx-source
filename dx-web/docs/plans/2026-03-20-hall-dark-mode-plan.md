# Hall Dark Mode Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add light/dark mode toggle to all hall pages via the Sun icon button, using `next-themes` and semantic color tokens.

**Architecture:** Hall-scoped ThemeProvider wraps the hall layout. A client-side toggle button switches between light/dark themes stored in localStorage. All hardcoded light-mode color classes are replaced with semantic CSS variable-based tokens that respond to the `.dark` class already defined in `globals.css`.

**Tech Stack:** next-themes, TailwindCSS v4 with oklch theme, shadcn/ui semantic tokens

---

## Color Mapping Reference

Use this mapping consistently across ALL tasks:

| Hardcoded | Semantic Replacement |
|-----------|---------------------|
| `bg-white` | `bg-card` |
| `bg-slate-50` | `bg-muted` |
| `bg-slate-100` | `bg-muted` |
| `bg-slate-200` (dividers) | `bg-border` |
| `bg-slate-300` | `bg-border` |
| `text-slate-900` | `text-foreground` |
| `text-slate-800` | `text-foreground` |
| `text-slate-700` | `text-foreground` |
| `text-slate-600` | `text-muted-foreground` |
| `text-slate-500` | `text-muted-foreground` |
| `text-slate-400` | `text-muted-foreground` |
| `text-slate-300` | `text-muted-foreground` |
| `border-slate-200` | `border-border` |
| `border-slate-100` | `border-border` |
| `border-slate-300` | `border-border` |
| `hover:bg-slate-50` | `hover:bg-accent` |
| `hover:bg-slate-100` | `hover:bg-accent` |
| `hover:bg-slate-300` | `hover:bg-accent` |
| `placeholder:text-slate-400` | `placeholder:text-muted-foreground` |
| `focus:border-slate-*` | `focus:border-ring` |

**Keep unchanged:**
- `bg-slate-900/50`, `bg-slate-900/[0.38]` (modal overlays — work in both modes)
- `bg-black/50` (overlays)
- `text-white` (on colored backgrounds)
- All brand accents: teal-*, blue-*, emerald-*, violet-*, amber-*, orange-*, red-*, indigo-*

---

### Task 1: ThemeProvider + Toggle Button (Infrastructure)

**Files:**
- Create: `src/features/web/hall/components/hall-theme-provider.tsx`
- Create: `src/features/web/hall/components/theme-toggle-button.tsx`
- Modify: `src/app/(web)/hall/layout.tsx`
- Modify: `src/features/web/hall/components/top-actions.tsx`

**Step 1: Create HallThemeProvider**

Create `src/features/web/hall/components/hall-theme-provider.tsx`:

```tsx
"use client";

import { ThemeProvider } from "next-themes";

/** Scoped theme provider for all hall pages */
export function HallThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="light" storageKey="hall-theme">
      {children}
    </ThemeProvider>
  );
}
```

**Step 2: Create ThemeToggleButton**

Create `src/features/web/hall/components/theme-toggle-button.tsx`:

```tsx
"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";

/** Toggle button that switches between light and dark mode */
export function ThemeToggleButton() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  return (
    <button
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      className="flex h-10 w-10 items-center justify-center rounded-[10px] border border-border bg-card text-muted-foreground hover:bg-accent"
    >
      {mounted && theme === "dark" ? (
        <Moon className="h-[18px] w-[18px]" />
      ) : (
        <Sun className="h-[18px] w-[18px]" />
      )}
    </button>
  );
}
```

**Step 3: Wrap hall layout with ThemeProvider**

Modify `src/app/(web)/hall/layout.tsx` — wrap `children` with `HallThemeProvider`:

```tsx
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { HallThemeProvider } from "@/features/web/hall/components/hall-theme-provider";

export default async function HallLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  await fetchUserProfile();

  return <HallThemeProvider>{children}</HallThemeProvider>;
}
```

**Step 4: Replace static Sun button with ThemeToggleButton**

In `src/features/web/hall/components/top-actions.tsx`:
- Import `ThemeToggleButton`
- Replace the static `<button>` containing `<Sun>` with `<ThemeToggleButton />`
- Also migrate the Bell link and its hardcoded colors to semantic tokens

**Step 5: Verify build passes**

Run: `npm run build`
Expected: Build succeeds with no errors

**Step 6: Verify in browser**

Run: `npm run dev`
1. Navigate to `/hall`
2. Click the Sun/Moon button — page should toggle between light/dark
3. Refresh — theme should persist
4. Check other pages aren't affected (e.g., `/`, `/auth/login`)

**Step 7: Commit**

```
feat: add dark mode toggle infrastructure for hall pages
```

---

### Task 2: Hall Shell + Sidebar

**Files:**
- Modify: `src/features/web/hall/components/hall-main-shell.tsx`
- Modify: `src/features/web/hall/components/hall-sidebar.tsx`

Apply the color mapping to both files. Key replacements:
- `bg-slate-50` → `bg-muted` (shell background)
- `bg-white` → `bg-card` (mobile header, sidebar)
- `border-slate-200` → `border-border`
- `text-slate-900` → `text-foreground`
- `text-slate-500` → `text-muted-foreground`
- `hover:bg-slate-50` → `hover:bg-accent`
- `bg-slate-200` (dividers) → `bg-border`
- `text-slate-700`, `text-slate-400`, `text-slate-600` → `text-foreground` or `text-muted-foreground`

Keep unchanged: teal, orange, amber, violet CTA gradients, `bg-red-500` notification dot.

**Verify:** Build passes. Sidebar and shell look correct in both light and dark mode.

**Commit:** `refactor: migrate hall shell and sidebar to semantic color tokens`

---

### Task 3: Hall Top Bar + Action Buttons

**Files:**
- Modify: `src/features/web/hall/components/greeting-top-bar.tsx`
- Modify: `src/features/web/hall/components/page-top-bar.tsx`
- Modify: `src/features/web/hall/components/breadcrumb-top-bar.tsx`
- Modify: `src/features/web/hall/components/feedback-button.tsx`
- Modify: `src/features/web/hall/components/content-seek-button.tsx`
- Modify: `src/features/web/hall/components/game-search-trigger.tsx`

Apply the color mapping. These are all icon buttons and top bars sharing the same pattern: `border-slate-200 bg-white text-slate-500 hover:bg-slate-50`.

**Verify:** Build passes. All top bar variants and buttons render correctly in both modes.

**Commit:** `refactor: migrate hall top bars and action buttons to semantic tokens`

---

### Task 4: Hall Dashboard Components

**Files:**
- Modify: `src/features/web/hall/components/stats-row.tsx`
- Modify: `src/features/web/hall/components/game-progress-card.tsx`
- Modify: `src/features/web/hall/components/learning-heatmap.tsx`
- Modify: `src/features/web/hall/components/heatmap-grid.tsx`
- Modify: `src/features/web/hall/components/heatmap-summary.tsx`
- Modify: `src/features/web/hall/components/favorite-card.tsx`
- Modify: `src/features/web/hall/components/played-game-card.tsx`

Apply the color mapping. Keep teal-50 accent backgrounds unchanged.

**Verify:** Build passes. Dashboard cards, stats, heatmap render correctly in both modes.

**Commit:** `refactor: migrate hall dashboard components to semantic tokens`

---

### Task 5: Hall Modals

**Files:**
- Modify: `src/features/web/hall/components/content-seek-modal.tsx`
- Modify: `src/features/web/hall/components/feedback-modal.tsx`

Apply the color mapping. Modal backgrounds `bg-white` → `bg-card`, inputs and borders to semantic tokens.

**Verify:** Build passes. Open each modal in both modes — content is readable, inputs are visible.

**Commit:** `refactor: migrate hall modals to semantic tokens`

---

### Task 6: Games Feature Components

**Files:**
- Modify: `src/features/web/games/components/filter-section.tsx`
- Modify: `src/features/web/games/components/game-card.tsx`
- Modify: `src/features/web/games/components/games-page-content.tsx`
- Modify: `src/features/web/games/components/hero-card.tsx`
- Modify: `src/features/web/games/components/level-grid.tsx`
- Modify: `src/features/web/games/components/my-stats-card.tsx`
- Modify: `src/features/web/games/components/rules-card.tsx`

Apply the color mapping. ~38 instances.

**Verify:** Build passes. Games listing, game detail, level grid all render correctly.

**Commit:** `refactor: migrate games components to semantic tokens`

---

### Task 7: Play Feature Components

**Files:**
- Modify: `src/features/web/play/components/game-play-shell.tsx`
- Modify: `src/features/web/play/components/game-top-bar.tsx`
- Modify: `src/features/web/play/components/game-mode-card.tsx`
- Modify: `src/features/web/play/components/game-result-card.tsx`
- Modify: `src/features/web/play/components/game-settings-modal.tsx`
- Modify: `src/features/web/play/components/game-exit-modal.tsx`
- Modify: `src/features/web/play/components/game-pause-overlay.tsx`
- Modify: `src/features/web/play/components/game-reset-modal.tsx`
- Modify: `src/features/web/play/components/game-report-modal.tsx`
- Modify: `src/features/web/play/components/game-loading-screen.tsx`
- Modify: `src/features/web/play/components/progress-ring.tsx`
- Modify: `src/features/web/play/components/stat-row.tsx`

Apply the color mapping. ~80 instances. Keep overlay `bg-slate-900/50` unchanged.

**Verify:** Build passes. Start a game, check top bar, modals, result card, settings in both modes.

**Commit:** `refactor: migrate play components to semantic tokens`

---

### Task 8: AI Custom + AI Practice Components

**Files:**
- Modify all 16 files in `src/features/web/ai-custom/components/`
- Modify: `src/features/web/ai-practice/components/ai-chat-panel.tsx`
- Modify: `src/features/web/ai-practice/components/ai-topic-grid.tsx`

Apply the color mapping. ~191 instances total (largest batch).

**Verify:** Build passes. AI custom course creation, level editor, AI practice chat all render correctly.

**Commit:** `refactor: migrate ai-custom and ai-practice to semantic tokens`

---

### Task 9: Remaining Feature Components

**Files:**
- Modify: `src/features/web/leaderboard/components/leaderboard-content.tsx`
- Modify: `src/features/web/leaderboard/components/leaderboard-list.tsx`
- Modify: `src/features/web/leaderboard/components/leaderboard-my-rank.tsx`
- Modify: `src/features/web/leaderboard/components/leaderboard-podium.tsx`
- Modify: `src/features/web/community/components/community-feed.tsx`
- Modify: `src/features/web/invite/components/invite-content.tsx`
- Modify: `src/features/web/invite/components/invite-qr-card.tsx`
- Modify: `src/features/web/invite/components/invite-referral-table.tsx`
- Modify: `src/features/web/invite/components/share-snippets-modal.tsx`
- Modify: `src/features/web/groups/components/group-detail-content.tsx`
- Modify: `src/features/web/groups/components/group-list-content.tsx`

Apply the color mapping.

**Verify:** Build passes. Visit leaderboard, community, invite, groups pages in both modes.

**Commit:** `refactor: migrate leaderboard, community, invite, groups to semantic tokens`

---

### Task 10: Me + Redeem + Auth Components

**Files:**
- Modify all 10 files in `src/features/web/me/components/`
- Modify all 4 files in `src/features/web/redeem/components/`
- Modify: `src/features/web/auth/components/user-profile-menu.tsx`

Apply the color mapping.

**Verify:** Build passes. Me page, profile menu, redeem page all render correctly.

**Commit:** `refactor: migrate me, redeem, and auth profile to semantic tokens`

---

### Task 11: Shared Components + Route Pages

**Files:**
- Modify: `src/components/in/insufficient-beans-dialog.tsx`
- Modify: `src/components/in/data-table-pagination.tsx`
- Modify: `src/components/in/stat-card.tsx`
- Modify: `src/components/in/word-table.tsx`
- Modify: `src/components/in/tab-pill.tsx`
- Modify: `src/features/com/images/components/image-uploader.tsx`
- Modify: `src/app/(web)/hall/(main)/games/mine/page.tsx`
- Modify: `src/app/(web)/hall/(main)/favorites/page.tsx`
- Modify: `src/app/(web)/hall/(main)/ai-custom/[id]/[levelId]/page.tsx`

Apply the color mapping. Note: shared components in `components/in/` are also used outside hall — semantic tokens work in both light (default) and dark contexts, so this is safe.

**Verify:** Build passes. Check favorites, my games, and AI custom level pages in both modes.

**Commit:** `refactor: migrate shared components and route pages to semantic tokens`

---

### Task 12: Final Verification

**Step 1: Full build check**

Run: `npm run build`
Expected: Clean build with no errors

**Step 2: Full visual sweep in dark mode**

Visit every hall page in dark mode and check:
- [ ] Hall home dashboard
- [ ] Games listing
- [ ] Game detail page
- [ ] Game play (all modals: settings, pause, exit, reset, report, result)
- [ ] AI Custom (course list, course detail, level editor)
- [ ] AI Practice (topic grid, chat panel)
- [ ] Leaderboard
- [ ] Community feed
- [ ] Invite page
- [ ] Groups (list + detail)
- [ ] Me / Profile page
- [ ] Redeem page
- [ ] Notices page
- [ ] Favorites page
- [ ] Sidebar (desktop + mobile)

**Step 3: Verify light mode is not broken**

Repeat the same sweep in light mode — everything should look identical to before.

**Step 4: Check non-hall pages are unaffected**

- [ ] Public homepage (`/`)
- [ ] Auth pages (`/auth/login`)
- [ ] Admin pages (`/adm/`)

**Commit:** No commit needed unless fixes are required.
