# Play Fullscreen Toggle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire up the existing fullscreen button in the game top bar to toggle browser fullscreen mode.

**Architecture:** A custom `useFullscreen` hook wraps the Browser Fullscreen API, exposing `isFullscreen` state and `toggleFullscreen()`. The shell passes these to the top bar, which swaps the icon and calls the handler.

**Tech Stack:** Browser Fullscreen API, React hooks, lucide-react icons

---

### Task 1: Create `useFullscreen` hook

**Files:**
- Create: `src/features/web/play/hooks/use-fullscreen.ts`

**Step 1: Write the hook**

```typescript
"use client";

import { useCallback, useEffect, useState } from "react";

/** Toggle browser fullscreen and track state via fullscreenchange event. */
export function useFullscreen() {
  const [isFullscreen, setIsFullscreen] = useState(false);

  useEffect(() => {
    /** Sync React state when fullscreen changes (includes ESC key exit). */
    const handleChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };

    document.addEventListener("fullscreenchange", handleChange);
    document.addEventListener("webkitfullscreenchange", handleChange);

    return () => {
      document.removeEventListener("fullscreenchange", handleChange);
      document.removeEventListener("webkitfullscreenchange", handleChange);
    };
  }, []);

  /** Enter or exit fullscreen on the root element. */
  const toggleFullscreen = useCallback(async () => {
    try {
      if (!document.fullscreenElement) {
        const el = document.documentElement as HTMLElement & {
          webkitRequestFullscreen?: () => Promise<void>;
        };
        if (el.requestFullscreen) {
          await el.requestFullscreen();
        } else if (el.webkitRequestFullscreen) {
          await el.webkitRequestFullscreen();
        }
      } else {
        const doc = document as Document & {
          webkitExitFullscreen?: () => Promise<void>;
        };
        if (doc.exitFullscreen) {
          await doc.exitFullscreen();
        } else if (doc.webkitExitFullscreen) {
          await doc.webkitExitFullscreen();
        }
      }
    } catch {
      // Fullscreen not supported or blocked by browser — silently ignore
    }
  }, []);

  return { isFullscreen, toggleFullscreen };
}
```

**Step 2: Verify no lint errors**

Run: `npx eslint src/features/web/play/hooks/use-fullscreen.ts`
Expected: No errors

---

### Task 2: Wire fullscreen into `GameTopBar`

**Files:**
- Modify: `src/features/web/play/components/game-top-bar.tsx`

**Step 1: Add props**

Add `onFullscreen` and `isFullscreen` to `GameTopBarProps`:

```typescript
interface GameTopBarProps {
  player: { nickname: string; avatarUrl: string | null };
  levelName: string;
  elapsedTime: string;
  isFullscreen: boolean;
  onExit: () => void;
  onReset: () => void;
  onSettings: () => void;
  onPause: () => void;
  onReport: () => void;
  onFullscreen: () => void;
}
```

**Step 2: Wire handler and swap icon**

1. Add `Minimize` to the lucide-react import
2. Add `fullscreen: onFullscreen` to `actionHandlers`
3. Change the `actionButtons` array to use a function for the fullscreen icon so it can swap based on `isFullscreen`:
   - When `isFullscreen` → show `Minimize` icon
   - When not → show `Maximize` icon

In the button render, override the icon for the fullscreen button:

```tsx
{actionButtons.map((btn) => {
  const Icon = btn.action === "fullscreen" && isFullscreen ? Minimize : btn.icon;
  return (
    <button
      key={btn.label}
      type="button"
      aria-label={btn.label}
      onClick={actionHandlers[btn.action]}
      className="flex h-8 w-8 items-center justify-center rounded-lg"
    >
      <Icon className="h-[18px] w-[18px] text-slate-500" />
    </button>
  );
})}
```

---

### Task 3: Connect hook in `GamePlayShell`

**Files:**
- Modify: `src/features/web/play/components/game-play-shell.tsx`

**Step 1: Import and use the hook**

```typescript
import { useFullscreen } from "@/features/web/play/hooks/use-fullscreen";
```

Inside `GamePlayShell`, before the return:

```typescript
const { isFullscreen, toggleFullscreen } = useFullscreen();
```

**Step 2: Pass props to GameTopBar**

Add to the `<GameTopBar>` JSX:

```tsx
isFullscreen={isFullscreen}
onFullscreen={toggleFullscreen}
```

---

### Task 4: Manual verification

**Step 1: Start dev server**

Run: `npm run dev`

**Step 2: Test in browser**

1. Navigate to a game play page
2. Click the fullscreen button (rightmost in top bar) → page enters fullscreen, icon changes to `Minimize`
3. Click the fullscreen button again → exits fullscreen, icon reverts to `Maximize`
4. Enter fullscreen again, press ESC → exits fullscreen, icon reverts

**Step 3: Commit**

```bash
git add src/features/web/play/hooks/use-fullscreen.ts \
       src/features/web/play/components/game-top-bar.tsx \
       src/features/web/play/components/game-play-shell.tsx
git commit -m "feat: wire up fullscreen toggle in game play top bar"
```
