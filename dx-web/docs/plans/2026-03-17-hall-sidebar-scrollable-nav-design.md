# Hall Sidebar Scrollable Nav — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the hall sidebar's nav section scrollable while pinning the logo at top and CTAs at bottom.

**Architecture:** Restructure `SidebarContent` from a 2-child flex layout (`justify-between`) to a 3-child flex column: pinned header, scrollable nav (`flex-1 min-h-0 overflow-y-auto`), pinned CTAs.

**Tech Stack:** TailwindCSS, React

---

## Design

When viewport height is limited, the sidebar's nav items overflow into the bottom CTA cards. `SidebarContent` uses `flex justify-between` with 2 children — the top section grows unbounded with no scroll handling.

**Fix:** Split into 3 flex children so the nav scrolls independently.

```
flex h-full flex-col
  ├── Header (shrink-0)                        — pinned top
  ├── Nav (flex-1 min-h-0 overflow-y-auto)     — scrollable
  └── CTAs (shrink-0)                          — pinned bottom
```

---

### Task 1: Restructure SidebarContent into 3 flex sections

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx:173-223`

**Step 1: Edit the SidebarContent JSX**

Replace the current `SidebarContent` return block (lines 176-223) with:

```tsx
<div className="flex h-full flex-col">
  {/* Header — pinned top */}
  <div className="shrink-0">
    <div className="flex w-full items-center justify-between">
      <Link href="/" className="flex items-center gap-2.5">
        <GraduationCap className="h-7 w-7 text-teal-600" />
        <span className="text-lg font-extrabold text-slate-900">斗学</span>
      </Link>
      <button
        type="button"
        className="hidden md:flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-50"
      >
        <PanelLeftClose className="h-[18px] w-[18px]" />
      </button>
    </div>
  </div>

  {/* Nav — scrollable middle */}
  <nav className="mt-8 flex-1 overflow-y-auto min-h-0 flex flex-col gap-1">
    {navSections.map((section) => (
      <div key={section.items[0].label} className="flex flex-col gap-1">
        {section !== navSections[0] && (
          <div className="my-1 h-px w-full bg-slate-200" />
        )}
        {section.items.map((item) => (
          <NavItem
            key={item.label}
            icon={item.icon}
            label={item.label}
            href={item.href}
            active={pathname === item.href}
            onClick={onNavigate}
          />
        ))}
      </div>
    ))}
  </nav>

  {/* Bottom CTAs — pinned bottom */}
  <div className="shrink-0 mt-4 flex flex-col gap-2">
    {ctaItems.map((item) => (
      <CtaCard key={item.label} {...item} onClick={onNavigate} />
    ))}
  </div>
</div>
```

Key changes:
- Removed `justify-between` from outer div
- Header wrapped in `shrink-0` div (pinned top)
- Nav gets `mt-8 flex-1 overflow-y-auto min-h-0` (scrollable middle)
- CTAs get `shrink-0 mt-4` (pinned bottom)

**Step 2: Verify visually**

Run: `npm run dev`

Check:
1. Normal height: sidebar looks identical to before (no visible scrollbar)
2. Short viewport (resize browser to ~500px height): nav scrolls, logo stays at top, CTAs stay at bottom
3. Mobile: open sheet drawer, same scroll behavior applies

**Step 3: Run lint**

Run: `npm run lint`
Expected: No new errors

**Step 4: Commit**

```bash
git add src/features/web/hall/components/hall-sidebar.tsx
git commit -m "fix: make hall sidebar nav scrollable with pinned header and CTAs"
```
