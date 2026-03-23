# Redeem Page Design

## Overview

Full implementation of the redeem code system. Users can redeem codes to upgrade their membership grade. Admin (username=rainson) can generate codes in bulk and view all code records.

## Data Model

### UserRedeem (renamed from RedeemCode)

Table: `user_redeems` (was `redeem_codes`)

| Field | Type | Notes |
|-------|------|-------|
| id | Char(26) | ULID primary key |
| code | VarChar(19) | Unique. Format: `XXXX-XXXX-XXXX-XXXX` |
| grade | VarChar(20) | month, season, year, lifetime |
| userId | Char(26)? | Null until redeemed |
| redeemedAt | Timestamptz? | Null until redeemed |
| createdAt | Timestamptz | Auto |
| updatedAt | Timestamptz | Auto |

Indexes: `userId`, `createdAt`

### User Model Changes

- Rename `pay_due_at` → `vip_due_at`

## Code Generation

- Format: `XXXX-XXXX-XXXX-XXXX` (uppercase A-Z + digits 0-9)
- Generated with `crypto.getRandomValues()` — 36^16 combinations, enumeration-proof
- Bulk insert via `createMany()` with ULID ids
- Collision check on generation (regenerate if duplicate, extremely unlikely)
- Admin-only: server-side check `username === "rainson"`

## Redeem Flow

1. User enters code, clicks "立即兑换"
2. Server action validates input, fetches user profile
3. Service looks up code — reject if not found or already redeemed
4. Transaction: mark code redeemed + update user grade and vip_due_at

## VIP Due At Calculation

Grade durations: month=1, season=3, year=12, lifetime=null

**Lifetime:** Set `grade = "lifetime"`, `vip_due_at = null`

**Other grades — determine base date:**
- User is free OR vip_due_at is null OR expired → base = today
- vip_due_at is not expired → base = current vip_due_at

**Calculate new vip_due_at:**
- Add N months to base date
- Target = same day minus 1 in the target month
- If target month has no same day → last day of target month
- Example: March 17 + 1 month = April 16
- Example: Jan 31 + 1 month = Feb 28 (or 29 leap year)

**Grade update:** Always overwrite to the redeemed code's grade.

## Page Layout

```
PageTopBar: 兑换码 / 兑换码兑换会员

[生成兑换码 button] ← rainson only

Input Card:
  [请输入兑换码 input] [立即兑换 button]
  没有兑换码？购买会员

兑换记录 (user's own records):
  | 兑换码 | 兑换等级 | 兑换时间 |
  [pagination]

兑换码管理 (rainson only):
  | 兑换码 | 等级 | 状态 | 兑换用户 | 兑换时间 | 创建时间 |
  [pagination]
```

### Generate Modal

- 生成类型: Select (月度会员, 季度会员, 年度会员, 终身会员)
- 生成数量: Select (10, 50, 100, 500)
- 确认生成 button with loading state

## File Structure

```
prisma/schema/
  user-redeem.prisma

src/models/user-redeem/
  user-redeem.query.ts
  user-redeem.mutation.ts

src/features/web/redeem/
  actions/redeem.action.ts
  services/redeem.service.ts
  helpers/redeem-code.helper.ts
  helpers/vip-due-at.helper.ts
  schemas/redeem.schema.ts
  components/
    redeem-content.tsx
    redeem-input-card.tsx
    redeem-history-table.tsx
    redeem-admin-section.tsx
    generate-codes-modal.tsx
  hooks/
    use-redeem.ts
    use-redeem-history.ts
    use-redeem-admin.ts

src/models/user/
  user.mutation.ts  (add updateUserVip)
```

## Architecture

- Thin page → RedeemContent
- Server actions → services → models
- Transaction: redeem code + update user atomically
- Pagination: useTransition + server actions (invite referral pattern)
- Admin visibility: server-side check in actions, client-side via username prop
