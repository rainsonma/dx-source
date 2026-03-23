# Game Progress Card Polish Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a ListChecks icon to the "我的游戏进度" header and stabilize the card height for 5 rows.

**Architecture:** Pure UI changes in one client component. Match the existing icon pattern from LearningHeatmap.

**Tech Stack:** React, Lucide icons, TailwindCSS

---

### Task 1: Add ListChecks icon and fix list height

**Files:**
- Modify: `src/features/web/hall/components/game-progress-card.tsx:1-98`

**Step 1: Update imports**

Replace:
```tsx
import { ArrowRight, Gamepad2 } from "lucide-react";
```
With:
```tsx
import { ArrowRight, Gamepad2, ListChecks } from "lucide-react";
```

**Step 2: Add icon to header**

Replace the header section (lines 43-52):
```tsx
<div className="flex w-full items-center justify-between">
  <h3 className="text-base font-bold text-slate-900">我的游戏进度</h3>
  <Link ...>
```
With:
```tsx
<div className="flex w-full items-center justify-between">
  <div className="flex items-center gap-3">
    <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-teal-50">
      <ListChecks className="h-5 w-5 text-teal-600" />
    </div>
    <h3 className="text-base font-bold text-slate-900">我的游戏进度</h3>
  </div>
  <Link ...>
```

**Step 3: Add min-height to progress list**

Replace:
```tsx
<div className="flex flex-col gap-3">
```
With:
```tsx
<div className="flex min-h-[288px] flex-col gap-3">
```

**Step 4: Verify visually**

Run: `npm run dev`
- Navigate to http://localhost:3000/hall
- Confirm the ListChecks icon appears in teal-50 container next to "我的游戏进度"
- Confirm the card height stays stable with fewer than 5 items
- Confirm pagination still works

**Step 5: Commit**

```bash
git add src/features/web/hall/components/game-progress-card.tsx
git commit -m "feat: add icon and fixed height to game progress card"
```
