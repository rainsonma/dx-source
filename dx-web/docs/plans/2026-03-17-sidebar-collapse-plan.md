# Sidebar Collapse/Expand Toggle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a smooth 260px ↔ 130px collapse toggle to the hall sidebar via the PanelLeftClose button.

**Architecture:** Local `useState` in `HallSidebar`, passed to `SidebarContent`. Width animated with Tailwind transition classes. Mobile sidebar unaffected.

**Tech Stack:** React useState, Tailwind CSS transitions, Lucide icons (PanelLeftClose/PanelLeftOpen)

---

### Task 1: Add collapse state and toggle handler

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx`

**Step 1: Add PanelLeftOpen import**

Add `PanelLeftOpen` to the lucide-react import on line 7:

```tsx
import {
  GraduationCap,
  PanelLeftClose,
  PanelLeftOpen,
  // ... rest unchanged
} from "lucide-react";
```

**Step 2: Add `collapsed` prop to `SidebarContent`**

Update `SidebarContent` to accept and use a `collapsed` prop and `onToggle` callback:

```tsx
function SidebarContent({
  onNavigate,
  collapsed,
  onToggle,
}: {
  onNavigate?: () => void;
  collapsed?: boolean;
  onToggle?: () => void;
}) {
```

**Step 3: Wire the toggle button**

Replace the PanelLeftClose button (lines 185-190) with:

```tsx
{onToggle && (
  <button
    type="button"
    onClick={onToggle}
    className="hidden md:flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-50"
  >
    {collapsed ? (
      <PanelLeftOpen className="h-[18px] w-[18px]" />
    ) : (
      <PanelLeftClose className="h-[18px] w-[18px]" />
    )}
  </button>
)}
```

**Step 4: Add state to `HallSidebar` and animate width**

Update `HallSidebar` to:

```tsx
export function HallSidebar() {
  const [collapsed, setCollapsed] = useState(false);

  return (
    <aside
      className={`hidden md:flex h-full shrink-0 flex-col border-r border-slate-200 bg-white px-5 py-6 overflow-hidden transition-all duration-300 ease-in-out ${
        collapsed ? "w-[130px]" : "w-[260px]"
      }`}
    >
      <SidebarContent
        collapsed={collapsed}
        onToggle={() => setCollapsed((prev) => !prev)}
      />
    </aside>
  );
}
```

Add `useState` import at top of file:

```tsx
import { useState } from "react";
```

**Step 5: Verify**

Run: `npm run build`
Expected: Build succeeds with no errors.

Manual check:
1. Open `/hall` in browser at desktop width
2. Click PanelLeftClose icon → sidebar smoothly shrinks to 130px, icon becomes PanelLeftOpen
3. Click PanelLeftOpen icon → sidebar smoothly expands to 260px, icon becomes PanelLeftClose
4. All content visible at both widths (text wraps/truncates naturally)
5. Mobile Sheet sidebar unchanged

**Step 6: Commit**

```bash
git add src/features/web/hall/components/hall-sidebar.tsx
git commit -m "feat: add sidebar collapse/expand toggle with smooth animation"
```
