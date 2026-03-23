# Sidebar CTA Enhancement Design

## Overview

Enhance the three bottom CTA cards in the hall sidebar (推广有奖, 兑换码, 续费升级) with subtitles, colored icon backgrounds, gradient badges, and proper linking. Also link the sidebar logo to the home landing page and auth-protect the membership page.

## Changes

### 1. Bottom CTA Cards — Enhanced Layout

Each CTA card gets:
- **Colored icon background**: icon wrapped in a rounded square with gradient background, white icon
- **Subtitle**: small descriptive text below the label
- **Badge** (where applicable): gradient pill next to the label

#### Card Details

| Card | Subtitle | Icon Gradient | Badge |
|------|----------|---------------|-------|
| 推广有奖 | 推广、邀请、赚佣金 | `from-orange-400 to-red-500` | HOT (orange→red gradient, white text) |
| 兑换码 | 兑换码兑换会员 | `from-violet-400 to-purple-600` | — |
| 续费升级 | 选择会员套餐 | `from-amber-300 to-yellow-500` | VIP (amber→yellow gradient, white text) |

#### Layout Structure

```
┌─────────────────────────────────────────┐
│  [icon]  Label  [BADGE]         ›       │
│  (bg)    subtitle text                  │
└─────────────────────────────────────────┘
```

- Icon container: `h-8 w-8 rounded-lg flex items-center justify-center bg-gradient-to-br`
- Icon: `h-4 w-4 text-white`
- Label: `text-[13px] text-slate-700 font-medium`
- Subtitle: `text-[11px] text-slate-400`
- Badge: `text-[10px] font-semibold px-1.5 py-0.5 rounded-full bg-gradient-to-r text-white`

### 2. 续费升级 Link

Change from `<button>` to `<Link href="/auth/membership">`. Navigates away from hall to the standalone membership page with auth layout.

### 3. Membership Page Auth Protection

Add `auth()` check to `src/app/(web)/auth/membership/page.tsx`. Redirect unauthenticated users to `/auth/signin`. Convert to async server component.

### 4. Sidebar Logo Link

Wrap the `GraduationCap` icon + "斗学" text in `<Link href="/">` so it navigates to the home landing page.

## Files Changed

| File | Change |
|------|--------|
| `src/features/web/hall/components/hall-sidebar.tsx` | Enhanced CTAs (subtitles, icon backgrounds, badges), logo → Link, 续费升级 → Link |
| `src/app/(web)/auth/membership/page.tsx` | Add auth guard, redirect if unauthenticated |
