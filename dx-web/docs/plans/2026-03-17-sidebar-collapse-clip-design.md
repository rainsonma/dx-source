# Sidebar Collapse — Clip Content Design

## Problem

When the hall sidebar collapses from 260px to 130px, the content reflows and squeezes messily. The desired behavior is a "sliding window" effect where content stays in place and the right portion is simply hidden.

## Solution

Fixed-width inner container with overflow clipping.

### Changes

**`<aside>` (HallSidebar):**
- Remove `px-5 py-6` padding from the aside — move to inner container
- Keep existing `overflow-hidden` and `transition-[width] duration-300 ease-in-out`

**Inner content div (SidebarContent root):**
- Add fixed `w-[260px] px-5 py-6`
- Content always renders at 260px regardless of sidebar width
- The aside clips the overflow when collapsed

**Header (collapsed state):**
- Hide the "斗学" text label when collapsed to fit logo icon + toggle button within visible area
- Logo icon and toggle button remain side by side with tighter spacing

## Behavior

- Expanded (260px): Full sidebar visible, all content shown normally
- Collapsed (130px): Left ~90px of content visible (130px minus padding), right portion clipped
- Toggle: Click PanelLeftOpen/PanelLeftClose button only — no hover reveal
- Mobile: Unaffected (uses separate Sheet component)
