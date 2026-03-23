# Sidebar Collapse/Expand Toggle Design

## Summary

Add a toggle to the hall sidebar's PanelLeftClose button that collapses the sidebar from 260px to 130px with a smooth animation. Clicking again restores the full width.

## Requirements

- Toggle sidebar between 260px (expanded) and 130px (collapsed)
- Smooth animated transition (~300ms)
- All content remains visible — text wraps/truncates naturally at narrower width
- Toggle button always visible, icon swaps between PanelLeftClose and PanelLeftOpen
- Mobile sidebar (Sheet) completely unaffected
- No changes to layout.tsx — main content uses flex-1 and adapts automatically

## Approach

Local state in `HallSidebar` component (Approach A — self-contained, no shared state needed).

## Implementation Details

### State

- `useState<boolean>(false)` in `HallSidebar`
- Passed to `SidebarContent` via a `collapsed` prop

### Sidebar `<aside>`

- Width class toggles: `w-[260px]` ↔ `w-[130px]`
- Add `transition-all duration-300 ease-in-out` for animation
- Add `overflow-hidden` to prevent content spill during transition

### Toggle Button

- Wire `onClick` to toggle `collapsed` state
- Icon: `PanelLeftClose` when expanded, `PanelLeftOpen` when collapsed
- Button stays in same header position, always visible

### Content at 130px

- No icon-only mode
- Text wraps/truncates naturally
- Nav items, CTAs, scroll behavior unchanged

### Unchanged

- Mobile sidebar (Sheet drawer)
- Nav items, CTA cards, styling, colors
- Hall layout structure
- All other components

## Files Modified

- `src/features/web/hall/components/hall-sidebar.tsx` — add state, toggle handler, width transition
