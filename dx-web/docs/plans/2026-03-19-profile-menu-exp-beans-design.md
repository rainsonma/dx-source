# Profile Menu: EXP & Beans Display Items

## Summary

Replace the first two menu items in `UserProfileMenu` dropdown:
- 个人中心 → 经验值 (display-only, with number badge)
- 应用设置 → 能量豆 (display-only, with number badge)

## Changes

### 1. Add `beans` to `UserProfile` type and query

- `src/models/user/user.query.ts` — add `beans: true` to `getUserProfile()` select, include in return
- `src/features/web/auth/types/user.types.ts` — add `beans: number` to `UserProfile`

### 2. Update menu component

**File:** `src/features/web/auth/components/user-profile-menu.tsx`

Replace the first group's static items with two display-only items:

| Item | Icon | Badge Style |
|------|------|-------------|
| 经验值 | `Star` (lucide) | indigo — `bg-indigo-100 text-indigo-600` (matches existing Lv. badge) |
| 能量豆 | `Coins` (lucide, consistent with me-hero.tsx) | amber — `bg-amber-100 text-amber-600` |

- Items are non-clickable (no `onClick`, no `href`)
- Number rendered as a right-aligned colored badge pill
- Remove `个人中心` and `应用设置` from the menu items array

### 3. Visual layout

```
┌──────────────────────────┐
│  [Avatar] Name  Lv.5     │
│  @username               │
├──────────────────────────┤
│  ⭐ 经验值         [1250] │  ← indigo badge, no link
│  🪙 能量豆        [10000] │  ← amber badge, no link
├──────────────────────────┤
│  👑 升级会员              │
│  🎫 兑换会员              │
│  🎁 推广邀请              │
├──────────────────────────┤
│  📄 帮助文档              │
├──────────────────────────┤
│  🚪 安全退出              │
└──────────────────────────┘
```

## Files touched

1. `src/models/user/user.query.ts`
2. `src/features/web/auth/types/user.types.ts`
3. `src/features/web/auth/components/user-profile-menu.tsx`
