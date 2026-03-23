# Energy Beans Feature Design

## Summary

Add a virtual currency system (Energy Beans) to the platform. Paid members receive a monthly grant of beans (10,000 standard, 15,000 lifetime). Unused granted beans are disabled on each monthly anniversary and replaced with a fresh grant, as long as the membership is active.

## Schema Changes

### User Model

- Rename `stars` (unused, default 0) → `beans`
- Add `grantedBeans Int @default(0)` after `beans`

### New UserBean Model

```prisma
model UserBean {
  id        String   @id @db.Char(26)
  userId    String   @map("user_id") @db.Char(26)
  beans     Int                        // signed: +10000 or -10000
  origin    Int      @default(0)       // balance before transaction
  result    Int      @default(0)       // balance after (origin + beans)
  slug      String                     // operation type
  reason    String                     // Chinese description
  data      Json?    @db.JsonB         // flexible metadata
  createdAt DateTime @map("created_at")
  updatedAt DateTime @map("updated_at")

  @@map("user_bean")
}
```

No DB-level FK constraints (code-level `assertFK()` pattern).

## Constants

### `src/consts/bean-slug.ts`

- `MEMBERSHIP_GRANT` → `"membership-grant"`
- `MONTHLY_RESET_DEBIT` → `"monthly-reset-debit"`
- `MONTHLY_RESET_CREDIT` → `"monthly-reset-credit"`

### `src/consts/bean-reason.ts`

- `MEMBERSHIP_GRANT` → `"会员购买赠送"`
- `MONTHLY_RESET_DEBIT` → `"月度未使用能量豆清零"`
- `MONTHLY_RESET_CREDIT` → `"月度能量豆续发"`

## Models Layer

### `src/models/user-bean/user-bean.mutation.ts`

Core function `createBeanEntry(userId, beans, slug, reason, data?, tx?)`:
1. `assertFK(client, [{ table: "user", id: userId }])`
2. Read current `User.beans` → set as `origin`
3. Compute `result = origin + beans`
4. Insert `UserBean` record
5. Update `User.beans = result` and `User.grantedBeans` accordingly

All within the same transaction. Callers wrap in `db.$transaction()` when not already inside one.

### `src/models/user-bean/user-bean.query.ts`

- `findBeansByUserId(userId, page)` — paginated ledger
- `sumBeansByUserId(userId)` — for reconciliation

## Redeem Integration

Extend `redeemCode()` in `src/features/web/redeem/services/redeem.service.ts`:
- Inside existing transaction, after updating grade + vipDueAt
- Call `createBeanEntry(userId, amount, MEMBERSHIP_GRANT, ...)` with `tx`
- Amount: 15,000 for lifetime, 10,000 otherwise
- `data`: `{ gradeAtTime, vipDueAt }`

## Cron Script

### `scripts/cron/reset-energy-beans.ts` — daily at 1 AM

1. Find users whose monthly anniversary matches today
2. Reset day = day-of-month from first `membership-grant` UserBean
3. Edge case: day 29/30/31 → use last day of current month
4. Per user, in a transaction:
   - Expired + grantedBeans > 0 → debit only
   - Expired + grantedBeans = 0 → skip
   - Active + grantedBeans > 0 → debit + credit
   - Active + grantedBeans = 0 → credit only
5. Uses standalone `createDb()` from `scripts/lib/db.ts`

## FIFO Consumption

Granted beans consumed first when spending:
- `beans -= spent`
- `grantedBeans = max(0, grantedBeans - spent)`

## Consistency Guarantees

- Every UserBean insert atomically updates User.beans and User.grantedBeans
- All operations in transactions
- `assertFK()` for code-level FK validation
- `User.beans` always equals sum of all UserBean.beans for that user

## Rules Reference

See `rules/EnergyBeanRule.md` for detailed rules, examples, and edge cases.
