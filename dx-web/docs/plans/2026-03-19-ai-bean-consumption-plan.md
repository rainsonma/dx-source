# AI Energy Bean Consumption Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add energy bean consumption to all 5 AI operations with upfront deduction, failure refund, and an insufficient-beans dialog guiding users to a new `/recharge` page.

**Architecture:** Each AI service checks balance and deducts beans before calling DeepSeek. On failure, a refund entry is created. The client detects `INSUFFICIENT_BEANS` errors and shows a shared alert dialog. A new `consumeBeans`/`refundBeans` helper wraps the existing `createBeanEntry` with FIFO logic.

**Tech Stack:** Next.js 16, Prisma, shadcn/ui AlertDialog, TypeScript

**Design doc:** `docs/plans/2026-03-19-ai-bean-consumption-design.md`

---

### Task 1: Add bean slugs and reasons constants

**Files:**
- Modify: `src/consts/bean-slug.ts`
- Modify: `src/consts/bean-reason.ts`

**Step 1: Add 10 new slugs to `bean-slug.ts`**

Add to `BEAN_SLUGS`:
```typescript
AI_GENERATE_CONSUME: "ai-generate-consume",
AI_GENERATE_REFUND: "ai-generate-refund",
AI_FORMAT_SENTENCE_CONSUME: "ai-format-sentence-consume",
AI_FORMAT_SENTENCE_REFUND: "ai-format-sentence-refund",
AI_FORMAT_VOCAB_CONSUME: "ai-format-vocab-consume",
AI_FORMAT_VOCAB_REFUND: "ai-format-vocab-refund",
AI_BREAK_CONSUME: "ai-break-consume",
AI_BREAK_REFUND: "ai-break-refund",
AI_GEN_ITEMS_CONSUME: "ai-gen-items-consume",
AI_GEN_ITEMS_REFUND: "ai-gen-items-refund",
```

Add to `BEAN_SLUG_LABELS`:
```typescript
"ai-generate-consume": "AI 生成消耗",
"ai-generate-refund": "AI 生成失败退还",
"ai-format-sentence-consume": "语句格式化消耗",
"ai-format-sentence-refund": "语句格式化失败退还",
"ai-format-vocab-consume": "词汇格式化消耗",
"ai-format-vocab-refund": "词汇格式化失败退还",
"ai-break-consume": "分解消耗",
"ai-break-refund": "分解失败退还",
"ai-gen-items-consume": "生成消耗",
"ai-gen-items-refund": "生成失败退还",
```

**Step 2: Add 10 new reasons to `bean-reason.ts`**

Add to `BEAN_REASONS`:
```typescript
AI_GENERATE_CONSUME: "AI 生成消耗",
AI_GENERATE_REFUND: "AI 生成失败退还",
AI_FORMAT_SENTENCE_CONSUME: "语句格式化消耗",
AI_FORMAT_SENTENCE_REFUND: "语句格式化失败退还",
AI_FORMAT_VOCAB_CONSUME: "词汇格式化消耗",
AI_FORMAT_VOCAB_REFUND: "词汇格式化失败退还",
AI_BREAK_CONSUME: "分解消耗",
AI_BREAK_REFUND: "分解失败退还",
AI_GEN_ITEMS_CONSUME: "生成消耗",
AI_GEN_ITEMS_REFUND: "生成失败退还",
```

**Step 3: Verify build**

Run: `npm run build`
Expected: Build succeeds (types auto-derive from `as const`)

**Step 4: Commit**

```
feat: add AI bean consumption slugs and reasons
```

---

### Task 2: Create word count helper

**Files:**
- Create: `src/features/web/ai-custom/helpers/count-words.ts`

**Step 1: Create the helper**

```typescript
/** Count English words in text by splitting on whitespace */
export function countWords(text: string): number {
  return text.trim().split(/\s+/).filter(Boolean).length;
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```
feat: add word count helper for bean cost calculation
```

---

### Task 3: Create bean consume/refund helpers

**Files:**
- Create: `src/features/web/ai-custom/helpers/bean-consume.ts`

This is a server-only file. It wraps `createBeanEntry` from `src/models/user-bean/user-bean.mutation.ts`.

**Step 1: Create the helper**

```typescript
import "server-only";

import { db } from "@/lib/db";
import { createBeanEntry } from "@/models/user-bean/user-bean.mutation";
import type { BeanSlug } from "@/consts/bean-slug";
import type { BeanReason } from "@/consts/bean-reason";
import type { Prisma } from "@/lib/db";

/** Error thrown when user has insufficient beans for an operation */
export class InsufficientBeansError extends Error {
  constructor(
    public readonly required: number,
    public readonly available: number
  ) {
    super("Insufficient beans");
  }
}

/**
 * Check balance and deduct beans atomically.
 * Throws InsufficientBeansError if user does not have enough beans.
 * Uses FIFO: grantedBeans consumed first.
 */
export async function consumeBeans(
  userId: string,
  amount: number,
  slug: BeanSlug,
  reason: BeanReason,
  data?: Prisma.InputJsonValue
): Promise<void> {
  await db.$transaction(async (tx) => {
    const user = await tx.user.findUniqueOrThrow({
      where: { id: userId },
      select: { beans: true, grantedBeans: true },
    });

    if (user.beans < amount) {
      throw new InsufficientBeansError(amount, user.beans);
    }

    const grantedDelta = -Math.min(amount, user.grantedBeans);
    await createBeanEntry(userId, -amount, slug, reason, grantedDelta, data, tx);
  });
}

/**
 * Refund beans after a failed AI operation.
 * Credits beans back; refunded beans go to grantedBeans.
 */
export async function refundBeans(
  userId: string,
  amount: number,
  slug: BeanSlug,
  reason: BeanReason,
  data?: Prisma.InputJsonValue
): Promise<void> {
  await createBeanEntry(userId, amount, slug, reason, amount, data);
}
```

**Key decisions:**
- `consumeBeans` reads balance inside a transaction to prevent race conditions
- On insufficient balance, throws `InsufficientBeansError` (caught by calling services)
- `refundBeans` credits the amount back, adding to `grantedBeans` (same as a grant)
- Both delegate to existing `createBeanEntry` — no duplication of ledger logic

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```
feat: add consumeBeans and refundBeans helpers
```

---

### Task 4: Create /recharge page

**Files:**
- Create: `src/app/(web)/recharge/page.tsx`

**Step 1: Create the page**

```typescript
import { redirect } from "next/navigation";
import { auth } from "@/lib/auth";

export default async function RechargePage() {
  const session = await auth();
  if (!session?.user) redirect("/auth/signin");

  return (
    <div className="flex w-full max-w-[1200px] flex-col items-center gap-6 py-10">
      <h1 className="text-[32px] font-bold text-slate-900">充值</h1>
      <p className="text-sm text-slate-500">充值页面建设中</p>
    </div>
  );
}
```

**Step 2: Verify build and visit `http://localhost:3000/recharge`**

Run: `npm run build`

**Step 3: Commit**

```
feat: add empty /recharge page with auth protection
```

---

### Task 5: Create insufficient beans dialog component

**Files:**
- Create: `src/components/in/insufficient-beans-dialog.tsx`

**Step 1: Create the dialog**

Uses shadcn/ui AlertDialog (already in `src/components/ui/alert-dialog.tsx`). Uses `useRouter` from `next/navigation` for the recharge link.

```typescript
"use client";

import { useRouter } from "next/navigation";
import { BatteryWarning } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

type InsufficientBeansDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  required: number;
  available: number;
};

/** Alert dialog shown when user lacks enough energy beans for an AI operation */
export function InsufficientBeansDialog({
  open,
  onOpenChange,
  required,
  available,
}: InsufficientBeansDialogProps) {
  const router = useRouter();

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <BatteryWarning className="h-5 w-5 text-amber-500" />
            能量豆不足
          </AlertDialogTitle>
          <AlertDialogDescription>
            本次操作需要 <span className="font-semibold text-slate-900">{required}</span> 能量豆，当前余额{" "}
            <span className="font-semibold text-slate-900">{available}</span>，还差{" "}
            <span className="font-semibold text-red-600">{required - available}</span>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            className="bg-teal-600 hover:bg-teal-700"
            onClick={() => router.push("/recharge")}
          >
            去充值
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```
feat: add insufficient beans alert dialog component
```

---

### Task 6: Add bean consumption to generate-metadata service (op 1: AI 生成)

**Files:**
- Modify: `src/features/web/ai-custom/services/generate-metadata.service.ts`

**Step 1: Add bean logic to `handleGenerateMetadata`**

Add imports at top:
```typescript
import { consumeBeans, refundBeans, InsufficientBeansError } from "@/features/web/ai-custom/helpers/bean-consume";
import { BEAN_SLUGS } from "@/consts/bean-slug";
import { BEAN_REASONS } from "@/consts/bean-reason";
```

Insert bean logic after input validation (after `const { difficulty, keywords } = parsed.data;`), before the `try` block that calls DeepSeek:

```typescript
const AI_GENERATE_COST = 5;

try {
  await consumeBeans(
    session.user.id,
    AI_GENERATE_COST,
    BEAN_SLUGS.AI_GENERATE_CONSUME,
    BEAN_REASONS.AI_GENERATE_CONSUME,
    { operation: "ai-generate", keywords }
  );
} catch (err) {
  if (err instanceof InsufficientBeansError) {
    return NextResponse.json(
      { error: "能量豆不足", code: "INSUFFICIENT_BEANS", required: err.required, available: err.available },
      { status: 402 }
    );
  }
  throw err;
}
```

In the existing `catch` block for DeepSeek errors, add refund **before** returning the error response. Both the `DeepSeekError` branch and the generic fallback branch need refund:

```typescript
catch (err) {
  // Refund beans on any AI failure
  await refundBeans(
    session.user.id,
    AI_GENERATE_COST,
    BEAN_SLUGS.AI_GENERATE_REFUND,
    BEAN_REASONS.AI_GENERATE_REFUND,
    { operation: "ai-generate", keywords }
  );

  if (err instanceof DeepSeekError) {
    const mapped = mapDeepSeekError(err, "AI 服务");
    return NextResponse.json({ error: mapped.error }, { status: mapped.status });
  }
  return NextResponse.json({ error: "AI 服务暂时不可用" }, { status: 502 });
}
```

**Important:** The WARNING response (content moderation) is NOT a failure — the AI succeeded, just flagged content. Beans are consumed and not refunded.

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```
feat: add bean consumption to AI generate metadata service
```

---

### Task 7: Add bean consumption to format-metadata service (ops 2-3)

**Files:**
- Modify: `src/features/web/ai-custom/services/format-metadata.service.ts`

**Step 1: Add bean logic to `handleFormatMetadata`**

Add imports at top:
```typescript
import { consumeBeans, refundBeans, InsufficientBeansError } from "@/features/web/ai-custom/helpers/bean-consume";
import { BEAN_SLUGS } from "@/consts/bean-slug";
import { BEAN_REASONS } from "@/consts/bean-reason";
import { countWords } from "@/features/web/ai-custom/helpers/count-words";
```

Insert bean logic after `const { content, formatType } = parsed.data;`, before the `try` block that calls DeepSeek:

```typescript
const wordCount = countWords(content);
const consumeSlug = formatType === "sentence"
  ? BEAN_SLUGS.AI_FORMAT_SENTENCE_CONSUME
  : BEAN_SLUGS.AI_FORMAT_VOCAB_CONSUME;
const consumeReason = formatType === "sentence"
  ? BEAN_REASONS.AI_FORMAT_SENTENCE_CONSUME
  : BEAN_REASONS.AI_FORMAT_VOCAB_CONSUME;
const refundSlug = formatType === "sentence"
  ? BEAN_SLUGS.AI_FORMAT_SENTENCE_REFUND
  : BEAN_SLUGS.AI_FORMAT_VOCAB_REFUND;
const refundReason = formatType === "sentence"
  ? BEAN_REASONS.AI_FORMAT_SENTENCE_REFUND
  : BEAN_REASONS.AI_FORMAT_VOCAB_REFUND;
const beanData = { operation: `ai-format-${formatType}`, wordCount };

try {
  await consumeBeans(session.user.id, wordCount, consumeSlug, consumeReason, beanData);
} catch (err) {
  if (err instanceof InsufficientBeansError) {
    return NextResponse.json(
      { error: "能量豆不足", code: "INSUFFICIENT_BEANS", required: err.required, available: err.available },
      { status: 402 }
    );
  }
  throw err;
}
```

In the existing `catch` block for DeepSeek errors, add refund before returning:

```typescript
catch (err) {
  await refundBeans(session.user.id, wordCount, refundSlug, refundReason, beanData);

  if (err instanceof DeepSeekError) {
    const mapped = mapDeepSeekError(err, "格式化服务");
    return NextResponse.json({ error: mapped.error }, { status: mapped.status });
  }
  return NextResponse.json({ error: "格式化服务暂时不可用" }, { status: 502 });
}
```

**Note:** WARNING responses (content moderation, type mismatch) are NOT failures — beans consumed, not refunded.

**Step 2: Verify build**

Run: `npm run build`

**Step 3: Commit**

```
feat: add bean consumption to format metadata service
```

---

### Task 8: Add bean consumption to break-metadata service (op 4)

**Files:**
- Modify: `src/features/web/ai-custom/services/break-metadata.service.ts`

**Step 1: Add imports**

```typescript
import { consumeBeans, refundBeans, InsufficientBeansError } from "@/features/web/ai-custom/helpers/bean-consume";
import { BEAN_SLUGS } from "@/consts/bean-slug";
import { BEAN_REASONS } from "@/consts/bean-reason";
import { countWords } from "@/features/web/ai-custom/helpers/count-words";
```

**Step 2: Add bean deduction after metas are fetched**

After `const metas = await getContentMetasForBreak(gameLevelId, session.user.id);` and after the `metas.length === 0` early return, add:

```typescript
const metaWordCounts = metas.map((m) => ({ id: m.id, words: countWords(m.sourceData) }));
const totalCost = metaWordCounts.reduce((sum, m) => sum + m.words, 0);

try {
  await consumeBeans(
    session.user.id,
    totalCost,
    BEAN_SLUGS.AI_BREAK_CONSUME,
    BEAN_REASONS.AI_BREAK_CONSUME,
    { operation: "ai-break", gameLevelId, metaCount: metas.length, totalWords: totalCost }
  );
} catch (err) {
  if (err instanceof InsufficientBeansError) {
    return NextResponse.json(
      { error: "能量豆不足", code: "INSUFFICIENT_BEANS", required: err.required, available: err.available },
      { status: 402 }
    );
  }
  throw err;
}
```

**Step 3: Add refund logic after batch completes**

The current code runs `runWithConcurrency` and in the `.then()` writes the final SSE event and closes. Modify the `.then()` to refund failed items:

```typescript
runWithConcurrency(tasks, CONCURRENCY_LIMIT, (done, total, success) => {
  writer.write({ done, total, status: success ? "ok" : "failed" });
  if (!success) {
    // Track failed meta index for refund calculation
    failedIndices.push(done - 1);
  }
}).then(async ({ processed, failed }) => {
  // Refund beans for failed metas
  if (failed > 0) {
    const failedWords = failedIndices.reduce(
      (sum, idx) => sum + (metaWordCounts[idx]?.words ?? 0),
      0
    );
    if (failedWords > 0) {
      await refundBeans(
        session.user.id,
        failedWords,
        BEAN_SLUGS.AI_BREAK_REFUND,
        BEAN_REASONS.AI_BREAK_REFUND,
        { operation: "ai-break", gameLevelId, failedCount: failed, refundedWords: failedWords }
      );
    }
  }

  writer.write({
    done: metas.length,
    total: metas.length,
    processed,
    failed,
    complete: true,
  });
  writer.close();
});
```

Add `const failedIndices: number[] = [];` before the `runWithConcurrency` call.

**Important:** The `runWithConcurrency` callback receives `(done, total, success)` — `done` is 1-indexed, so the meta index is `done - 1`. Verify this matches the callback signature in `src/features/web/ai-custom/helpers/concurrency-runner.ts`.

**Step 4: Verify build**

Run: `npm run build`

**Step 5: Commit**

```
feat: add bean consumption to break metadata service
```

---

### Task 9: Add bean consumption to generate-content-items service (op 5)

**Files:**
- Modify: `src/features/web/ai-custom/services/generate-content-items.service.ts`
- Modify: `src/models/content-item/content-item.query.ts` (add batch query)

**Step 1: Add a batch query to fetch all content items for multiple metas**

In `src/models/content-item/content-item.query.ts`, add:

```typescript
/** Fetch all content items pending generation for a list of meta IDs */
export async function getContentItemsForGenerateBatch(metaIds: string[]) {
  return db.contentItem.findMany({
    where: {
      contentMetaId: { in: metaIds },
      items: { equals: Prisma.DbNull },
    },
    select: {
      id: true,
      content: true,
      contentType: true,
      contentMetaId: true,
    },
    orderBy: { order: "asc" },
  });
}
```

**Step 2: Add imports to generate-content-items service**

```typescript
import { consumeBeans, refundBeans, InsufficientBeansError } from "@/features/web/ai-custom/helpers/bean-consume";
import { BEAN_SLUGS } from "@/consts/bean-slug";
import { BEAN_REASONS } from "@/consts/bean-reason";
import { countWords } from "@/features/web/ai-custom/helpers/count-words";
import { getContentItemsForGenerateBatch } from "@/models/content-item/content-item.query";
```

**Step 3: Add bean deduction after metas are fetched**

After `const metas = await getContentMetasForGenerate(gameLevelId, session.user.id);` and after the `metas.length === 0` early return, add:

```typescript
const allPendingItems = await getContentItemsForGenerateBatch(metas.map((m) => m.id));
const metaItemWordCounts = new Map<string, number>();
for (const item of allPendingItems) {
  const prev = metaItemWordCounts.get(item.contentMetaId) ?? 0;
  metaItemWordCounts.set(item.contentMetaId, prev + countWords(item.content));
}
const totalCost = Array.from(metaItemWordCounts.values()).reduce((sum, w) => sum + w, 0);

try {
  await consumeBeans(
    session.user.id,
    totalCost,
    BEAN_SLUGS.AI_GEN_ITEMS_CONSUME,
    BEAN_REASONS.AI_GEN_ITEMS_CONSUME,
    { operation: "ai-gen-items", gameLevelId, metaCount: metas.length, totalWords: totalCost }
  );
} catch (err) {
  if (err instanceof InsufficientBeansError) {
    return NextResponse.json(
      { error: "能量豆不足", code: "INSUFFICIENT_BEANS", required: err.required, available: err.available },
      { status: 402 }
    );
  }
  throw err;
}
```

**Step 4: Add refund logic after batch completes**

Same pattern as Task 8. Add `const failedMetaIds: string[] = [];` before `runWithConcurrency`. Track failed meta IDs in the progress callback. After batch completes:

```typescript
if (failed > 0 && failedMetaIds.length > 0) {
  const failedWords = failedMetaIds.reduce(
    (sum, id) => sum + (metaItemWordCounts.get(id) ?? 0),
    0
  );
  if (failedWords > 0) {
    await refundBeans(
      session.user.id,
      failedWords,
      BEAN_SLUGS.AI_GEN_ITEMS_REFUND,
      BEAN_REASONS.AI_GEN_ITEMS_REFUND,
      { operation: "ai-gen-items", gameLevelId, failedCount: failed, refundedWords: failedWords }
    );
  }
}
```

**Note:** Need to adjust the `tasks` array and progress callback to track which meta failed. The simplest approach: index the tasks array to match `metas` order, so `failedIndices[i]` maps to `metas[i].id`.

**Step 5: Verify build**

Run: `npm run build`

**Step 6: Commit**

```
feat: add bean consumption to generate content items service
```

---

### Task 10: Extend client helpers with INSUFFICIENT_BEANS error handling

**Files:**
- Modify: `src/features/web/ai-custom/helpers/format-api.ts`
- Modify: `src/features/web/ai-custom/helpers/generate-api.ts`
- Modify: `src/features/web/ai-custom/helpers/stream-progress.ts`

**Step 1: Update `format-api.ts`**

Change the `FormatResult` error branch:
```typescript
type FormatResult =
  | { ok: true; formatted: string; sourceTypes: SourceType[] }
  | { ok: false; message: string; code?: string; required?: number; available?: number };
```

In the `!res.ok` branch, extract extra fields:
```typescript
if (!res.ok) {
  return {
    ok: false,
    message: data.error ?? "格式化失败",
    code: data.code,
    required: data.required,
    available: data.available,
  };
}
```

**Step 2: Update `generate-api.ts`**

Same pattern — change `GenerateResult` error branch and extract fields:
```typescript
type GenerateResult =
  | { ok: true; generated: string; sourceType: SourceType }
  | { ok: false; message: string; code?: string; required?: number; available?: number };
```

```typescript
if (!res.ok) {
  return {
    ok: false,
    message: data.error ?? "生成失败",
    code: data.code,
    required: data.required,
    available: data.available,
  };
}
```

**Step 3: Update `stream-progress.ts`**

Change `StreamResult` error branch:
```typescript
type StreamResult =
  | { ok: true; processed: number; failed: number }
  | { ok: false; message: string; code?: string; required?: number; available?: number };
```

In the `!res.ok` branch:
```typescript
if (!res.ok) {
  const data = await res.json();
  return {
    ok: false,
    message: data.error ?? "请求失败",
    code: data.code,
    required: data.required,
    available: data.available,
  };
}
```

**Step 4: Verify build**

Run: `npm run build`

**Step 5: Commit**

```
feat: extend client helpers with INSUFFICIENT_BEANS error fields
```

---

### Task 11: Add insufficient beans handling to add-metadata-dialog

**Files:**
- Modify: `src/features/web/ai-custom/components/add-metadata-dialog.tsx`

**Step 1: Add dialog import and state**

Add import:
```typescript
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";
```

Add state inside `AddMetadataDialog`:
```typescript
const [beanDialogOpen, setBeanDialogOpen] = useState(false);
const [beanRequired, setBeanRequired] = useState(0);
const [beanAvailable, setBeanAvailable] = useState(0);
```

**Step 2: Add helper to check and show dialog**

```typescript
function handleBeanError(result: { code?: string; required?: number; available?: number }) {
  if (result.code === "INSUFFICIENT_BEANS") {
    setBeanRequired(result.required ?? 0);
    setBeanAvailable(result.available ?? 0);
    setBeanDialogOpen(true);
    return true;
  }
  return false;
}
```

**Step 3: Integrate into `handleFormat`**

In `handleFormat`, after `const result = await formatMetadata(text, formatType);` and `setFormattingType(null);`:

Replace the existing `if (!result.ok)` block:
```typescript
if (!result.ok) {
  if (handleBeanError(result)) return;
  setErrorMessage(result.message);
  toast.warning(result.message);
  return;
}
```

**Step 4: Integrate into `handleGenerate`**

In `handleGenerate`, replace the existing `if (!result.ok)` block:
```typescript
if (!result.ok) {
  if (handleBeanError(result)) return;
  setAiErrorMessage(result.message);
  toast.warning(result.message);
  return;
}
```

**Step 5: Render the dialog**

Add before the closing `</Dialog>`:
```tsx
<InsufficientBeansDialog
  open={beanDialogOpen}
  onOpenChange={setBeanDialogOpen}
  required={beanRequired}
  available={beanAvailable}
/>
```

**Step 6: Reset dialog state on close**

In `handleOpenChange`, add:
```typescript
setBeanDialogOpen(false);
setBeanRequired(0);
setBeanAvailable(0);
```

**Step 7: Verify build**

Run: `npm run build`

**Step 8: Commit**

```
feat: add insufficient beans dialog to add-metadata-dialog
```

---

### Task 12: Add insufficient beans handling to level-units-panel

**Files:**
- Modify: `src/features/web/ai-custom/components/level-units-panel.tsx`

**Step 1: Add dialog import and state**

Add import:
```typescript
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";
```

Add state inside `LevelUnitsPanel`:
```typescript
const [beanDialogOpen, setBeanDialogOpen] = useState(false);
const [beanRequired, setBeanRequired] = useState(0);
const [beanAvailable, setBeanAvailable] = useState(0);
```

**Step 2: Integrate into `handleBreak`**

After `const result = await breakMetadata(...)`, in the `if (!result.ok)` block:
```typescript
if (!result.ok) {
  if (result.message !== "已取消") {
    if (result.code === "INSUFFICIENT_BEANS") {
      setBeanRequired(result.required ?? 0);
      setBeanAvailable(result.available ?? 0);
      setBeanDialogOpen(true);
    } else {
      toast.error(result.message);
    }
  }
  return;
}
```

**Step 3: Integrate into `handleGenerate`**

Same pattern as `handleBreak`:
```typescript
if (!result.ok) {
  if (result.message !== "已取消") {
    if (result.code === "INSUFFICIENT_BEANS") {
      setBeanRequired(result.required ?? 0);
      setBeanAvailable(result.available ?? 0);
      setBeanDialogOpen(true);
    } else {
      toast.error(result.message);
    }
  }
  return;
}
```

**Step 4: Render the dialog**

Add after the last `</AlertDialog>`, before `</>`:
```tsx
<InsufficientBeansDialog
  open={beanDialogOpen}
  onOpenChange={setBeanDialogOpen}
  required={beanRequired}
  available={beanAvailable}
/>
```

**Step 5: Verify build**

Run: `npm run build`

**Step 6: Commit**

```
feat: add insufficient beans dialog to level-units-panel
```

---

### Task 13: Update EnergyBeanRule.md

**Files:**
- Modify: `rules/EnergyBeanRule.md`

**Step 1: Add AI Consumption section**

Add after the "Monthly Reset" section, before end of file:

```markdown
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
```

**Step 2: Update the Slugs table in the existing "Ledger Model" section**

Add the 10 new slugs to the existing table.

**Step 3: Commit**

```
docs: update EnergyBeanRule with AI consumption rules
```

---

### Task 14: Final build verification

**Step 1: Full build**

Run: `npm run build`
Expected: Clean build, no errors

**Step 2: Lint**

Run: `npm run lint`
Expected: No new lint warnings

**Step 3: Manual smoke test**

1. Visit `/recharge` while logged out → should redirect to `/auth/signin`
2. Visit `/recharge` while logged in → should show placeholder page
3. Trigger any AI operation with 0 beans → should see insufficient beans dialog with correct numbers
4. Trigger any AI operation with sufficient beans → should work as before, bean balance decremented
5. Check `user_beans` table for new ledger entries with correct slugs/reasons
