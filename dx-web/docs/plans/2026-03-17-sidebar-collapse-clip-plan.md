# Sidebar Collapse Clip Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the hall sidebar collapse by clipping content instead of reflowing it.

**Architecture:** Move padding from `<aside>` to a fixed-width inner div. The aside clips overflow when narrowed. Conditionally hide logo text when collapsed so the toggle button stays visible.

**Tech Stack:** React, TailwindCSS

---

### Task 1: Move padding and add fixed width to inner container

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx:248-250` (HallSidebar aside)
- Modify: `src/features/web/hall/components/hall-sidebar.tsx:187` (SidebarContent root div)

**Step 1: Remove padding from aside**

In `HallSidebar`, change the aside className from:
```tsx
className={`hidden md:flex h-full shrink-0 flex-col border-r border-slate-200 bg-white px-5 py-6 overflow-hidden transition-[width] duration-300 ease-in-out ${
  collapsed ? "w-[130px]" : "w-[260px]"
}`}
```
to:
```tsx
className={`hidden md:flex h-full shrink-0 flex-col border-r border-slate-200 bg-white overflow-hidden transition-[width] duration-300 ease-in-out ${
  collapsed ? "w-[130px]" : "w-[260px]"
}`}
```

**Step 2: Add fixed width and padding to SidebarContent root div**

In `SidebarContent`, change the root div from:
```tsx
<div className="flex h-full flex-col">
```
to:
```tsx
<div className="flex h-full w-[260px] shrink-0 flex-col px-5 py-6">
```

**Step 3: Verify in browser**

Run: `npm run dev`
- Expand sidebar: content displays normally at 260px
- Collapse sidebar: content stays in place, right portion clipped at 130px
- Animation smooth during transition

**Step 4: Commit**

```bash
git add src/features/web/hall/components/hall-sidebar.tsx
git commit -m "fix: clip sidebar content on collapse instead of reflowing"
```

---

### Task 2: Keep toggle button visible when collapsed

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx:191-193` (header section in SidebarContent)

**Step 1: Conditionally hide logo text when collapsed**

Change the header from:
```tsx
<Link href="/" className="flex items-center gap-2.5">
  <GraduationCap className="h-7 w-7 text-teal-600" />
  <span className="text-lg font-extrabold text-slate-900">斗学</span>
</Link>
```
to:
```tsx
<Link href="/" className="flex items-center gap-2.5">
  <GraduationCap className="h-7 w-7 shrink-0 text-teal-600" />
  {!collapsed && (
    <span className="text-lg font-extrabold text-slate-900">斗学</span>
  )}
</Link>
```

**Step 2: Verify in browser**

- Expanded: Logo icon + "斗学" text + toggle button all visible
- Collapsed: Logo icon + toggle button side by side, text hidden, both fit within 130px
- Toggle button clickable in both states

**Step 3: Commit**

```bash
git add src/features/web/hall/components/hall-sidebar.tsx
git commit -m "fix: hide logo text when sidebar collapsed to keep toggle visible"
```
