# Profile Menu: EXP & Beans Display Items — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace 个人中心/应用设置 menu items with display-only 经验值/能量豆 items showing colored number badges.

**Architecture:** Add `beans` to the lightweight `UserProfile` query/type, then update the dropdown menu component to render two non-clickable display items with colored badge pills instead of navigable links.

**Tech Stack:** Next.js, React, TypeScript, Prisma, lucide-react, shadcn/ui DropdownMenu

---

### Task 1: Add `beans` to UserProfile query

**Files:**
- Modify: `src/models/user/user.query.ts:50-83`

**Step 1: Add `beans` to select and return**

In `getUserProfile()`, add `beans: true` to the select object (after `exp: true` on line 58), and add `beans: user.beans` to the return object (after `exp: user.exp` on line 77):

```typescript
// In select (line 58, after exp: true):
beans: true,

// In return (line 77, after exp: user.exp):
beans: user.beans,
```

**Step 2: Verify build**

Run: `npx next build` or `npx tsc --noEmit`
Expected: PASS (type will be inferred, but UserProfile type needs updating next)

---

### Task 2: Add `beans` to UserProfile type

**Files:**
- Modify: `src/features/web/auth/types/user.types.ts:1-11`

**Step 1: Add beans field**

Add `beans: number` after the `exp` field (line 9):

```typescript
export type UserProfile = {
  id: string;
  username: string;
  nickname: string | null;
  email: string | null;
  grade: UserGrade;
  exp: number;
  beans: number;
  avatarUrl: string | null;
};
```

---

### Task 3: Update menu component

**Files:**
- Modify: `src/features/web/auth/components/user-profile-menu.tsx`

**Step 1: Update icon imports**

Replace `Settings` with `Coins` and add `Star` to the lucide imports (line 5-14):

```typescript
import {
  User,
  Star,
  Coins,
  Crown,
  Ticket,
  Gift,
  FileText,
  LogOut,
  ChevronDown,
} from "lucide-react";
```

**Step 2: Remove 个人中心 and 应用设置 from menuItems**

Remove the first group from `menuItems` (lines 45-48). The array becomes:

```typescript
const menuItems = [
  {
    group: [
      { label: "升级会员", icon: Crown, href: "/auth/membership" },
      { label: "兑换会员", icon: Ticket, href: "/hall/redeem" },
      { label: "推广邀请", icon: Gift, href: "/hall/invite" },
    ],
  },
  {
    group: [{ label: "帮助文档", icon: FileText, href: "/docs" }],
  },
];
```

**Step 3: Add display-only EXP and Beans items**

Insert a new section between `<DropdownMenuLabel>` and the `menuItems.map(...)` block (after line 125, before the `{menuItems.map(...)}` block). These are non-clickable display items:

```tsx
<DropdownMenuSeparator />
<DropdownMenuGroup>
  <DropdownMenuItem className="flex items-center justify-between" onSelect={(e) => e.preventDefault()}>
    <span className="flex items-center gap-2">
      <Star className="h-4 w-4" />
      经验值
    </span>
    <span className="rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-semibold text-indigo-600">
      {profile.exp.toLocaleString()}
    </span>
  </DropdownMenuItem>
  <DropdownMenuItem className="flex items-center justify-between" onSelect={(e) => e.preventDefault()}>
    <span className="flex items-center gap-2">
      <Coins className="h-4 w-4" />
      能量豆
    </span>
    <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-600">
      {profile.beans.toLocaleString()}
    </span>
  </DropdownMenuItem>
</DropdownMenuGroup>
```

**Step 4: Verify build**

Run: `npm run build`
Expected: PASS

**Step 5: Visual check**

Run: `npm run dev`, navigate to hall, open profile dropdown menu.
Expected: First group shows 经验值 with indigo badge and 能量豆 with amber badge. Not clickable.

---

### Task 4: Commit

```bash
git add src/models/user/user.query.ts src/features/web/auth/types/user.types.ts src/features/web/auth/components/user-profile-menu.tsx
git commit -m "feat: show EXP and beans with badges in profile menu"
```
