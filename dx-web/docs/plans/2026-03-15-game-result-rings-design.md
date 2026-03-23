# Game Result Card — Progress Ring Redesign

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the top stat blocks in `game-result-card.tsx` with circular SVG progress rings for 4 key percentage KPIs, keeping remaining raw stats as blocks below.

**Architecture:** New `ProgressRing` component renders an SVG donut chart with animated fill. The result card is restructured: header/score at top, then a ring row for percentages, then stat blocks for raw data, with dividers between sections.

**Tech Stack:** React, SVG, Tailwind CSS (no new dependencies)

---

### Task 1: Create ProgressRing component

**Files:**
- Create: `src/features/web/play/components/progress-ring.tsx`

**Step 1: Create the component file**

```tsx
"use client";

import { useEffect, useState } from "react";

interface ProgressRingProps {
  percent: number;
  color: string;
  trackColor?: string;
  size?: number;
  strokeWidth?: number;
  label: string;
}

/** Circular SVG progress ring with animated fill and center percentage */
export function ProgressRing({
  percent,
  color,
  trackColor = "stroke-slate-200",
  size = 90,
  strokeWidth = 8,
  label,
}: ProgressRingProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // Delay to trigger CSS transition from 0 → target
    const id = requestAnimationFrame(() => setMounted(true));
    return () => cancelAnimationFrame(id);
  }, []);

  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const clamped = Math.max(0, Math.min(percent, 100));
  const offset = circumference * (1 - (mounted ? clamped : 0) / 100);

  return (
    <div className="flex flex-col items-center gap-1.5">
      <div className="relative" style={{ width: size, height: size }}>
        <svg
          width={size}
          height={size}
          className="-rotate-90"
        >
          {/* Background track */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            strokeWidth={strokeWidth}
            className={trackColor}
          />
          {/* Foreground arc */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            strokeWidth={strokeWidth}
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            className={color}
            style={{ transition: "stroke-dashoffset 0.8s ease-out" }}
          />
        </svg>
        {/* Center percentage text */}
        <span className="absolute inset-0 flex items-center justify-center text-lg font-extrabold text-slate-900">
          {Math.round(clamped)}%
        </span>
      </div>
      <span className="text-xs text-slate-400">{label}</span>
    </div>
  );
}
```

**Step 2: Verify no build errors**

Run: `npm run build`
Expected: build succeeds (component is not imported yet, but file must parse)

**Step 3: Commit**

```bash
git add src/features/web/play/components/progress-ring.tsx
git commit -m "feat: add ProgressRing SVG component for game result"
```

---

### Task 2: Update GameResultCard layout

**Files:**
- Modify: `src/features/web/play/components/game-result-card.tsx`

**Step 1: Update imports**

Remove unused icons: `Target` (accuracy was in statsRow1, now a ring).

Add import for the new component:
```tsx
import { ProgressRing } from "@/features/web/play/components/progress-ring";
```

**Step 2: Add ring metric calculations**

After the existing `accuracyPercent` calculation (line ~70), add:

```tsx
const completionPercent = totalItems > 0
  ? Math.round(((totalItems - skipCount) / totalItems) * 100)
  : 0;
const scorePercent = totalItems > 0
  ? Math.round(Math.min((score / (totalItems * SCORING.CORRECT_ANSWER)) * 100, 100))
  : 0;
const comboPercent = totalItems > 0
  ? Math.round((combo.maxCombo / totalItems) * 100)
  : 0;
```

**Step 3: Replace stats grid with ring row + restructured blocks**

Remove old `statsRow1`, `statsRow2`, `statsRow3` arrays. Replace with:

```tsx
const statsRow1: StatItem[] = [
  { icon: CircleCheck, iconColor: "text-emerald-500", value: `${correctCount}/${totalItems}`, label: "正确题数" },
  { icon: CircleX, iconColor: "text-rose-400", value: `${wrongCount}`, label: "错误" },
  { icon: SkipForward, iconColor: "text-slate-400", value: `${skipCount}`, label: "跳过" },
];

const statsRow2: StatItem[] = [
  { icon: Zap, iconColor: "text-red-500", value: `${combo.maxCombo}连击`, label: "最高连击" },
  { icon: Timer, iconColor: "text-amber-500", value: elapsedTime, label: "用时" },
  { icon: Star, iconColor: "text-amber-500", value: `+${expEarned}`, label: "经验值" },
];
```

**Step 4: Update JSX — replace stats grid section**

Replace the stats grid JSX (between the first divider and the `completing` indicator) with:

```tsx
{/* Ring row — key percentage KPIs */}
<div className="grid w-full grid-cols-4 gap-2">
  <ProgressRing percent={accuracyPercent} color="stroke-amber-500" label="正确率" />
  <ProgressRing percent={completionPercent} color="stroke-amber-500" label="完成率" />
  <ProgressRing percent={scorePercent} color="stroke-teal-500" label="得分率" />
  <ProgressRing percent={comboPercent} color="stroke-teal-500" label="连击率" />
</div>

<div className="h-px w-full bg-slate-100" />

{/* Stats blocks — raw data */}
<div className="flex w-full flex-col gap-3">
  <div className="grid grid-cols-3 gap-3">
    {statsRow1.map((stat) => (
      <StatBlock key={stat.label} {...stat} />
    ))}
  </div>
  <div className="grid grid-cols-3 gap-3">
    {statsRow2.map((stat) => (
      <StatBlock key={stat.label} {...stat} />
    ))}
  </div>
</div>

<div className="h-px w-full bg-slate-100" />
```

**Step 5: Clean up unused imports**

Remove `Target` from the lucide-react import (no longer used).

**Step 6: Verify build**

Run: `npm run build`
Expected: build succeeds

**Step 7: Commit**

```bash
git add src/features/web/play/components/game-result-card.tsx
git commit -m "feat: add progress rings and restructure game result card layout"
```

---

### Task 3: Visual verification

**Step 1: Run dev server and test**

Run: `npm run dev`

Manually play through a game level and verify the result screen shows:
- 4 animated progress rings in a row (amber + teal colors)
- Rings animate from 0 to their target value on mount
- Labels below each ring
- Divider between rings and stat blocks
- 2 rows of 3 stat blocks below
- Divider between stat blocks and action buttons
- All action buttons still work

**Step 2: Test edge cases**
- 0% accuracy (all rings should handle 0 gracefully)
- 100% accuracy (rings fully filled)
- Skip-heavy session (completion rate < 100%)
- High combo (combo rate ring fills proportionally)
