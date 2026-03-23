# Game Progress Card Polish

## Goal

Optimize the "我的游戏进度" card on the hall dashboard with visual consistency and stable layout.

## Changes

All changes in `src/features/web/hall/components/game-progress-card.tsx`.

### 1. Add icon to header

Add a `ListChecks` icon matching the 学习热力图 pattern:
- `h-9 w-9` rounded-lg container with `bg-teal-50`
- `h-5 w-5` icon in `text-teal-600`

### 2. Fixed list height

Add `min-h-[288px]` on the progress list container so the card maintains consistent height when fewer than 5 items are on the current page.

Calculation: 5 items × ~48px + 4 gaps × 12px = 288px.

### 3. Page size

Stays at 5 (no change).
