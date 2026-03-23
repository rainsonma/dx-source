# AI Energy Bean Consumption Design

## Overview

Add energy bean consumption to the 5 AI operations in the ai-custom module. Users must have sufficient beans before each request. Beans are deducted upfront; on AI failure, a refund entry is created. An insufficient-beans alert dialog guides users to a new `/recharge` page.

## Operations & Costs

| # | Operation | Slug (consume) | Slug (refund) | Cost |
|---|-----------|---------------|---------------|------|
| 1 | AI 生成 | `ai-generate-consume` | `ai-generate-refund` | Fixed 5 beans |
| 2 | 语句检查并格式化 | `ai-format-sentence-consume` | `ai-format-sentence-refund` | 1 bean/word in input |
| 3 | 词汇检查并格式化 | `ai-format-vocab-consume` | `ai-format-vocab-refund` | 1 bean/word in input |
| 4 | 分解 | `ai-break-consume` | `ai-break-refund` | 1 bean/word in pending metas' `sourceData` |
| 5 | 生成 | `ai-gen-items-consume` | `ai-gen-items-refund` | 1 bean/word in pending items' `content` |

### Reasons (Chinese)

| Slug | Reason |
|------|--------|
| `ai-generate-consume` | AI 生成消耗 |
| `ai-generate-refund` | AI 生成失败退还 |
| `ai-format-sentence-consume` | 语句格式化消耗 |
| `ai-format-sentence-refund` | 语句格式化失败退还 |
| `ai-format-vocab-consume` | 词汇格式化消耗 |
| `ai-format-vocab-refund` | 词汇格式化失败退还 |
| `ai-break-consume` | 分解消耗 |
| `ai-break-refund` | 分解失败退还 |
| `ai-gen-items-consume` | 生成消耗 |
| `ai-gen-items-refund` | 生成失败退还 |

## Word Counting

`countWords(text)` — split on whitespace, count non-empty tokens.

- **Format operations** (ops 2-3): count words in user's input `content` string
- **分解** (op 4): count words in `sourceData` of metas fetched by `getContentMetasForBreak()` (only `isBreakDone=false`)
- **生成** (op 5): fetch all content items for metas from `getContentMetasForGenerate()` upfront, count words in each item's `content` — this is the exact data sent to DeepSeek, no duplication

## Server-Side Flow

### Single-request operations (ops 1-3)

```
1. Parse & validate input
2. Calculate cost (fixed 5 or countWords)
3. Read User.beans → if < cost → return INSUFFICIENT_BEANS error
4. consumeBeans(userId, cost, slug, reason, data)
5. Call DeepSeek API
6. On failure → refundBeans(userId, cost, refundSlug, refundReason, data)
7. On success → return result as normal
```

### Batch SSE operations (ops 4-5)

```
1. Fetch all pending data upfront (metas for break, content items for generate)
2. Calculate total cost from fetched data
3. Read User.beans → if < cost → return INSUFFICIENT_BEANS error (JSON, before SSE starts)
4. consumeBeans(userId, totalCost, slug, reason, data)
5. Run batch with SSE progress (existing logic unchanged)
6. After batch completes, if failed > 0:
   → sum word counts of failed items' data
   → refundBeans(userId, failedCost, refundSlug, refundReason, data)
```

### Bean helpers

New file: `src/features/web/ai-custom/helpers/bean-consume.ts` (server-only)

- `consumeBeans(userId, amount, slug, reason, data?)` — wraps `createBeanEntry` with negative beans and FIFO grantedDelta
- `refundBeans(userId, amount, slug, reason, data?)` — wraps `createBeanEntry` with positive beans

Both use `createBeanEntry` from `src/models/user-bean/user-bean.mutation.ts` — existing FIFO logic preserved.

## Client-Side Flow

### Error response format

```json
{ "error": "能量豆不足", "code": "INSUFFICIENT_BEANS", "required": 45, "available": 20 }
```

### Client helper changes

Extend `ok: false` return type with optional fields (non-breaking):

```typescript
{ ok: false; message: string; code?: string; required?: number; available?: number }
```

Updated helpers:
- `format-api.ts` — extract `code/required/available` from error JSON
- `generate-api.ts` — same
- `stream-progress.ts` — same (INSUFFICIENT_BEANS returned as JSON before SSE stream, handled by existing `!res.ok` branch)

### Insufficient beans dialog

New file: `src/components/in/insufficient-beans-dialog.tsx`

AlertDialog showing:
- Title: 能量豆不足
- Body: 本次操作需要 **{required}** 能量豆，当前余额 **{available}**，还差 **{required - available}**
- Cancel button
- Action button: 去充值 → navigates to `/recharge`

Rendered in two calling components:
- `add-metadata-dialog.tsx` — handles ops 1-3
- `level-units-panel.tsx` — handles ops 4-5

Each component manages its own dialog state. No global context needed.

## New /recharge Route

- Path: `src/app/(web)/recharge/page.tsx`
- Auth: page-level `auth()` check, redirect to `/auth/signin` if unauthenticated
- Content: empty placeholder for now

## Files Changed/Added

| File | Action |
|------|--------|
| `src/app/(web)/recharge/page.tsx` | New — empty auth-protected page |
| `src/consts/bean-slug.ts` | Edit — add 10 slugs |
| `src/consts/bean-reason.ts` | Edit — add 10 reasons |
| `src/features/web/ai-custom/helpers/count-words.ts` | New — word count utility |
| `src/features/web/ai-custom/helpers/bean-consume.ts` | New — consumeBeans/refundBeans |
| `src/features/web/ai-custom/services/generate-metadata.service.ts` | Edit — add bean logic |
| `src/features/web/ai-custom/services/format-metadata.service.ts` | Edit — add bean logic |
| `src/features/web/ai-custom/services/break-metadata.service.ts` | Edit — add bean logic |
| `src/features/web/ai-custom/services/generate-content-items.service.ts` | Edit — add bean logic + fetch items upfront |
| `src/features/web/ai-custom/helpers/format-api.ts` | Edit — extend error type |
| `src/features/web/ai-custom/helpers/generate-api.ts` | Edit — extend error type |
| `src/features/web/ai-custom/helpers/stream-progress.ts` | Edit — extend error type |
| `src/components/in/insufficient-beans-dialog.tsx` | New — shared alert dialog |
| `src/features/web/ai-custom/components/add-metadata-dialog.tsx` | Edit — handle INSUFFICIENT_BEANS |
| `src/features/web/ai-custom/components/level-units-panel.tsx` | Edit — handle INSUFFICIENT_BEANS |
| `rules/EnergyBeanRule.md` | Edit — add AI consumption section |

## What stays unchanged

- All existing bean grant/reset/FIFO logic
- DeepSeek API calls, prompts, response parsing
- SSE streaming, progress tracking, abort handling
- All existing UI behavior when user has sufficient beans
- Database schema — no migrations needed
