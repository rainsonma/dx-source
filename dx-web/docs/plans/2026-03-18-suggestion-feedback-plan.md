# Suggestion Feedback (建议反馈) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a "建议反馈" modal that lets users submit feedback by type, with duplicate detection that increments a counter.

**Architecture:** Server Action + shadcn Dialog. Form submits via server action → service → model. Duplicate type+description increments count instead of creating new record.

**Tech Stack:** Next.js 16, Prisma v7, Zod, shadcn Dialog, sonner toast, ulid

---

### Task 1: Constants

**Files:**
- Create: `src/consts/feedback-type.ts`

**Step 1: Create constants file**

Follow the pattern in `src/consts/user-grade.ts`.

```typescript
// src/consts/feedback-type.ts
export const FEEDBACK_TYPES = {
  FEATURE: "feature",
  CONTENT: "content",
  UX: "ux",
  BUG: "bug",
  OTHER: "other",
} as const;

export type FeedbackType = (typeof FEEDBACK_TYPES)[keyof typeof FEEDBACK_TYPES];

export const FEEDBACK_TYPE_LABELS: Record<FeedbackType, string> = {
  feature: "功能建议",
  content: "内容纠错",
  ux: "界面体验",
  bug: "Bug 报告",
  other: "其它",
};
```

**Step 2: Commit**

```
feat: add feedback type constants
```

---

### Task 2: Prisma Schema

**Files:**
- Create: `prisma/schema/feedback.prisma`

**Step 1: Create the schema file**

```prisma
model Feedback {
  id          String @id @db.Char(26)
  userId      String @map("user_id") @db.Char(26)
  type        String @db.VarChar(20)
  description String @db.VarChar(600)
  count       Int    @default(1)

  createdAt DateTime @default(now()) @map("created_at") @db.Timestamptz
  updatedAt DateTime @updatedAt @map("updated_at") @db.Timestamptz

  @@unique([type, description])
  @@index([userId])
  @@index([createdAt])
  @@map("feedbacks")
}
```

No FK constraint — project uses code-level checks only.

**Step 2: Run migration**

Run: `npx prisma migrate dev --name add-feedback`

**Step 3: Generate client**

Run: `npm run prisma:generate`

**Step 4: Commit**

```
feat: add Feedback prisma schema
```

---

### Task 3: Model Layer

**Files:**
- Create: `src/models/feedback/feedback.query.ts`
- Create: `src/models/feedback/feedback.mutation.ts`

**Step 1: Create query file**

```typescript
// src/models/feedback/feedback.query.ts
import "server-only";

import { db } from "@/lib/db";

/** Find a feedback record by type and description */
export async function findFeedbackByTypeAndDescription(
  type: string,
  description: string,
) {
  return db.feedback.findUnique({
    where: { type_description: { type, description } },
    select: { id: true, count: true },
  });
}
```

**Step 2: Create mutation file**

```typescript
// src/models/feedback/feedback.mutation.ts
import "server-only";

import { ulid } from "ulid";
import { db } from "@/lib/db";

type CreateFeedbackData = {
  userId: string;
  type: string;
  description: string;
};

/** Create a new feedback record */
export async function createFeedback(data: CreateFeedbackData) {
  return db.feedback.create({
    data: {
      id: ulid(),
      userId: data.userId,
      type: data.type,
      description: data.description,
    },
    select: { id: true },
  });
}

/** Increment the count of an existing feedback record */
export async function incrementFeedbackCount(id: string) {
  return db.feedback.update({
    where: { id },
    data: { count: { increment: 1 } },
    select: { id: true, count: true },
  });
}
```

**Step 3: Commit**

```
feat: add feedback model layer (query + mutation)
```

---

### Task 4: Zod Schema

**Files:**
- Create: `src/features/web/hall/schemas/feedback.schema.ts`

**Step 1: Create validation schema**

```typescript
// src/features/web/hall/schemas/feedback.schema.ts
import { z } from "zod";
import { FEEDBACK_TYPES } from "@/consts/feedback-type";

const feedbackTypeValues = Object.values(FEEDBACK_TYPES) as [string, ...string[]];

/** Schema for submitting feedback */
export const feedbackSchema = z.object({
  type: z.enum(feedbackTypeValues, {
    error: "请选择建议类型",
  }),
  description: z
    .string()
    .trim()
    .min(1, "请输入详细描述")
    .max(200, "最多200个字符"),
});

export type FeedbackInput = z.infer<typeof feedbackSchema>;
```

**Step 2: Commit**

```
feat: add feedback zod validation schema
```

---

### Task 5: Service

**Files:**
- Create: `src/features/web/hall/services/feedback.service.ts`

**Step 1: Create service**

```typescript
// src/features/web/hall/services/feedback.service.ts
import "server-only";

import { findFeedbackByTypeAndDescription } from "@/models/feedback/feedback.query";
import { createFeedback, incrementFeedbackCount } from "@/models/feedback/feedback.mutation";

type SubmitFeedbackData = {
  userId: string;
  type: string;
  description: string;
};

/** Submit feedback — creates new or increments count on duplicate type+description */
export async function submitFeedback(data: SubmitFeedbackData) {
  const existing = await findFeedbackByTypeAndDescription(
    data.type,
    data.description,
  );

  if (existing) {
    await incrementFeedbackCount(existing.id);
    return { duplicate: true };
  }

  await createFeedback(data);
  return { success: true };
}
```

**Step 2: Commit**

```
feat: add feedback service with duplicate detection
```

---

### Task 6: Server Action

**Files:**
- Create: `src/features/web/hall/actions/feedback.action.ts`

**Step 1: Create action**

Reference pattern: `src/features/web/hall/actions/content-seek.action.ts`

```typescript
// src/features/web/hall/actions/feedback.action.ts
"use server";

import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { feedbackSchema } from "@/features/web/hall/schemas/feedback.schema";
import { submitFeedback } from "@/features/web/hall/services/feedback.service";

/** Submit feedback for the current user */
export async function submitFeedbackAction(input: {
  type: string;
  description: string;
}) {
  try {
    const profile = await fetchUserProfile();
    if (!profile) {
      return { error: "未登录" };
    }

    const parsed = feedbackSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "参数错误" };
    }

    const result = await submitFeedback({
      userId: profile.id,
      ...parsed.data,
    });

    return result;
  } catch {
    return { error: "提交失败" };
  }
}
```

**Step 2: Commit**

```
feat: add feedback server action
```

---

### Task 7: Modal Component

**Files:**
- Create: `src/features/web/hall/components/feedback-modal.tsx`

**Step 1: Create modal**

Reference pattern: `src/features/web/hall/components/content-seek-modal.tsx`

Design reference: `design/dx-web.pen` node `IGkUF` (建议弹窗) — type tag selector + textarea.

```tsx
// src/features/web/hall/components/feedback-modal.tsx
"use client";

import { useState, useTransition } from "react";
import { Flag, Send, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog, DialogContent, DialogTitle, DialogDescription,
} from "@/components/ui/dialog";
import { FEEDBACK_TYPES, FEEDBACK_TYPE_LABELS } from "@/consts/feedback-type";
import type { FeedbackType } from "@/consts/feedback-type";
import { submitFeedbackAction } from "@/features/web/hall/actions/feedback.action";

const MAX_LEN = 200;

const TYPE_OPTIONS = Object.values(FEEDBACK_TYPES);

type FeedbackModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** Modal form for submitting feedback */
export function FeedbackModal({ open, onOpenChange }: FeedbackModalProps) {
  const [type, setType] = useState<FeedbackType>(FEEDBACK_TYPES.FEATURE);
  const [description, setDescription] = useState("");
  const [isPending, startTransition] = useTransition();

  /** Reset form fields */
  function resetForm() {
    setType(FEEDBACK_TYPES.FEATURE);
    setDescription("");
  }

  /** Handle form submission */
  function handleSubmit() {
    if (!description.trim()) return;

    startTransition(async () => {
      const result = await submitFeedbackAction({
        type,
        description: description.trim(),
      });

      if ("error" in result) {
        toast.error(result.error);
        return;
      }

      if ("duplicate" in result) {
        toast.info("已有相同建议，正在处理中...");
        return;
      }

      toast.success("提交成功");
      resetForm();
      onOpenChange(false);
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton
        className="max-w-[460px] gap-0 rounded-[20px] border-none p-0"
      >
        <div className="flex flex-col gap-5 p-7">
          {/* Header */}
          <DialogTitle className="flex items-center gap-2.5 text-xl font-bold text-slate-900">
            <Flag className="h-[18px] w-[18px] text-teal-600" />
            建议反馈
          </DialogTitle>
          <DialogDescription className="sr-only">
            提交您的建议或反馈
          </DialogDescription>

          <div className="h-px bg-slate-100" />

          <p className="text-sm leading-[1.5] text-slate-500">
            您的每一条建议都能帮助我们做得更好！
          </p>

          {/* Type tags */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-slate-900">
              建议类型
            </label>
            <div className="flex flex-wrap gap-1.5">
              {TYPE_OPTIONS.map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setType(t)}
                  disabled={isPending}
                  className={
                    t === type
                      ? "rounded-lg bg-teal-600/10 px-3 py-1.5 text-[13px] font-semibold text-teal-600"
                      : "rounded-lg border border-slate-200 bg-slate-50 px-3 py-1.5 text-[13px] font-medium text-slate-500 hover:bg-slate-100"
                  }
                >
                  {FEEDBACK_TYPE_LABELS[t]}
                </button>
              ))}
            </div>
          </div>

          {/* Description */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-slate-900">
              详细描述
            </label>
            <div className="relative">
              <textarea
                placeholder="请详细描述你的建议或遇到的问题，我们会认真阅读每一条反馈..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isPending}
                maxLength={MAX_LEN}
                rows={4}
                className="w-full resize-none rounded-[10px] border border-slate-200 bg-slate-50 px-4 py-3 text-sm leading-[1.5] text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
              {description.length > 0 && (
                <span className="pointer-events-none absolute right-3 bottom-2 text-xs text-slate-400">
                  {description.length}/{MAX_LEN}
                </span>
              )}
            </div>
          </div>

          {/* Buttons */}
          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
              className="flex h-12 flex-1 items-center justify-center rounded-xl border border-slate-200 bg-slate-50 text-[15px] font-medium text-slate-500 transition-colors hover:bg-slate-100 disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={isPending || !description.trim()}
              className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl bg-teal-600 text-[15px] font-semibold text-white transition-colors hover:bg-teal-700 disabled:opacity-50"
            >
              {isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
              提交建议
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Commit**

```
feat: add feedback modal component
```

---

### Task 8: Trigger Button + TopActions Integration

**Files:**
- Create: `src/features/web/hall/components/feedback-button.tsx`
- Modify: `src/features/web/hall/components/top-actions.tsx:30-32`

**Step 1: Create the client wrapper button**

```tsx
// src/features/web/hall/components/feedback-button.tsx
"use client";

import { useState } from "react";
import { Flag } from "lucide-react";
import { FeedbackModal } from "@/features/web/hall/components/feedback-modal";

/** Client button that opens the feedback modal */
export function FeedbackButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="flex h-10 w-10 items-center justify-center rounded-[10px] border border-slate-200 bg-white text-slate-500 hover:bg-slate-50"
      >
        <Flag className="h-[18px] w-[18px]" />
      </button>
      <FeedbackModal open={open} onOpenChange={setOpen} />
    </>
  );
}
```

**Step 2: Modify TopActions**

In `src/features/web/hall/components/top-actions.tsx`:

- Add import: `import { FeedbackButton } from "@/features/web/hall/components/feedback-button";`
- Replace the static Flag button (lines 30-32) with: `<FeedbackButton />`
- Remove the `Flag` import from lucide-react since it's no longer used in this file

Current top-actions.tsx lines 1-2:
```tsx
import Link from "next/link";
import { Search, Flag, Bell, Sun } from "lucide-react";
```

After change:
```tsx
import Link from "next/link";
import { Search, Bell, Sun } from "lucide-react";
```

**Step 3: Verify**

Run: `npm run build`
Expected: Build succeeds with no errors.

**Step 4: Commit**

```
feat: integrate feedback button and modal into top actions
```
