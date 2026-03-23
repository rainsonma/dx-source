# Energy Bean Rules

## Overview

Energy Beans are a virtual currency granted to paid members monthly. They follow a ledger pattern with FIFO consumption — granted beans are consumed first.

## User Fields

| Field | Type | Description |
|-------|------|-------------|
| `beans` | Int | Total bean balance (granted + bought) |
| `grantedBeans` | Int | Remaining beans from the latest grant, consumed first |

Both fields change together in every transaction. When `grantedBeans` changes, `beans` changes by the same amount.

## Grant Amounts

| Membership Grade | Monthly Grant |
|------------------|---------------|
| free | 0 (not eligible) |
| month | 10,000 |
| season | 10,000 |
| year | 10,000 |
| lifetime | 15,000 |

## Ledger Model (UserBean)

Every bean operation creates a `UserBean` record capturing:

| Field | Description |
|-------|-------------|
| `beans` | Signed amount (+10000 or -10000) |
| `origin` | User's `beans` balance before this transaction |
| `result` | User's `beans` balance after (`origin + beans`) |
| `slug` | Operation type constant |
| `reason` | Human-readable Chinese description |
| `data` | JSON metadata (grade, vipDueAt, etc.) |

### Slugs

| Slug | Reason       | Description |
|------|--------------|-------------|
| `membership-grant` | 购买会员赠送能量豆    | Initial grant on membership purchase/redeem |
| `monthly-reset-debit` | 月度未消耗赠送能量豆清零 | Disable remaining granted beans on monthly reset |
| `monthly-reset-credit` | 月度赠送能量豆续发    | New monthly grant for active members |
| `ai-generate-consume` | AI 生成消耗      | Deduct beans for AI story generation |
| `ai-generate-refund` | AI 生成失败退还    | Refund on AI generation failure |
| `ai-format-sentence-consume` | AI 语句格式化消耗   | Deduct beans for sentence formatting |
| `ai-format-sentence-refund` | AI 语句格式化失败退还 | Refund on sentence format failure |
| `ai-format-vocab-consume` | AI 词汇格式化消耗   | Deduct beans for vocabulary formatting |
| `ai-format-vocab-refund` | AI 词汇格式化失败退还 | Refund on vocabulary format failure |
| `ai-break-consume` | AI 分解消耗      | Deduct beans for content decomposition |
| `ai-break-refund` | AI 分解失败退还    | Refund on decomposition failure |
| `ai-gen-items-consume` | AI 生成消耗      | Deduct beans for word-level generation |
| `ai-gen-items-refund` | AI 生成失败退还    | Refund on word-level generation failure |

## Consistency Rules

1. Every `UserBean` insert must atomically update `User.beans` and `User.grantedBeans` in the same transaction
2. `User.beans` must always equal the sum of all `UserBean.beans` for that user
3. `User.grantedBeans` tracks the remaining portion from the latest grant, never exceeds `beans`

## FIFO Consumption

When a user spends beans:
- `beans -= spentAmount` (always)
- `grantedBeans = max(0, grantedBeans - spentAmount)` (granted consumed first)

### Example

```
State: beans=15000, grantedBeans=10000 (bought 5000 extra)
Spend 3000:
  beans = 15000 - 3000 = 12000
  grantedBeans = max(0, 10000 - 3000) = 7000

Spend 9000 more:
  beans = 12000 - 9000 = 3000
  grantedBeans = max(0, 7000 - 9000) = 0
  (7000 from grant + 2000 from bought consumed)
```

## Grant Trigger

Energy beans are granted when a user **redeems** a membership code (purchase flow to be added later). The grant happens inside the existing redeem transaction.

## Monthly Reset (Cron)

Runs daily at **1 AM** (`0 1 * * *`).

### Reset Day

Determined by the day-of-month of the user's first `membership-grant` UserBean record. Edge case: if reset day is 29/30/31 and the current month has fewer days, trigger on the last day of the month.

### Reset Logic

| Membership Status | grantedBeans | Action |
|-------------------|--------------|--------|
| Expired (`vipDueAt < now`, not lifetime) | > 0 | Debit `grantedBeans` only |
| Expired | = 0 | Skip (nothing to do) |
| Active | > 0 | Debit `grantedBeans`, then credit new grant |
| Active | = 0 | Credit new grant only |

### Reset Example (Season Member, granted 10000)

```
Month 1: Redeem season membership
  → +10000 (membership-grant)
  → beans=10000, grantedBeans=10000

Month 2 (reset day, used 3000):
  → -7000 (monthly-reset-debit)    beans=3000, grantedBeans=0
  → +10000 (monthly-reset-credit)  beans=13000, grantedBeans=10000

Month 3 (reset day, used all):
  → grantedBeans=0, skip debit
  → +10000 (monthly-reset-credit)  beans=10000+bought, grantedBeans=10000

Month 4 (membership expired, 4000 unused):
  → -4000 (monthly-reset-debit)    beans-=4000, grantedBeans=0
  → no credit (expired)
```

## AI Consumption

Energy beans are consumed when users trigger AI operations in the content creation module.

### Operations & Costs

| Operation | Slug | Cost | Word Source |
|-----------|------|------|-------------|
| AI 生成 | `ai-generate-consume` | Fixed 5 beans | — |
| 语句格式化 | `ai-format-sentence-consume` | 1 bean/word | Words in input text |
| 词汇格式化 | `ai-format-vocab-consume` | 1 bean/word | Words in input text |
| 分解 | `ai-break-consume` | 1 bean/word | Words in `sourceData` of pending metas |
| 生成 | `ai-gen-items-consume` | 1 bean/word | Words in `content` of pending items |

### Consumption Flow

1. Calculate cost from the exact request data
2. Check balance — if insufficient, return error (no beans deducted)
3. Deduct beans atomically (FIFO: grantedBeans consumed first)
4. Call AI API
5. On failure: create refund entry (slug: `*-refund`, reason: `*失败退还`)

For batch operations (分解, 生成): total cost is calculated upfront from all pending items. If some items in the batch fail, only the failed items' word count is refunded.

### Refund Slugs

| Slug | Reason |
|------|--------|
| `ai-generate-refund` | AI 生成失败退还 |
| `ai-format-sentence-refund` | 语句格式化失败退还 |
| `ai-format-vocab-refund` | 词汇格式化失败退还 |
| `ai-break-refund` | 分解失败退还 |
| `ai-gen-items-refund` | 生成失败退还 |
