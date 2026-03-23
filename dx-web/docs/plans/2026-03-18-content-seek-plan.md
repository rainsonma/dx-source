# Content Seek (求课程) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a "求课程" modal that lets users request courses, with duplicate detection that increments a counter.

**Architecture:** Server Action + shadcn Dialog. Form submits via server action → service → model. Duplicate courseName increments count instead of creating new record.

**Tech Stack:** Next.js 16, Prisma v7, Zod, shadcn Dialog, sonner toast, ulid

---

### Task 1: Prisma Schema

**Files:**
- Create: `prisma/schema/content-seek.prisma`

**Step 1: Create the schema file**

```prisma
model ContentSeek {
  id         String  @id @db.Char(26)
  userId     String  @map("user_id") @db.Char(26)
  courseName String  @unique @map("course_name") @db.VarChar(90)
  description String? @db.VarChar(90)
  diskUrl    String? @map("disk_url") @db.VarChar(90)
  count      Int     @default(1)

  createdAt DateTime @default(now()) @map("created_at") @db.Timestamptz
  updatedAt DateTime @updatedAt @map("updated_at") @db.Timestamptz

  @@index([userId])
  @@index([createdAt])
  @@map("content_seeks")
}
```

No FK constraint — project uses code-level checks only.

**Step 2: Run migration**

Run: `npx prisma migrate dev --name add-content-seek`

**Step 3: Generate client**

Run: `npm run prisma:generate`

**Step 4: Commit**

```
feat: add ContentSeek prisma schema
```

---

### Task 2: Model Layer

**Files:**
- Create: `src/models/content-seek/content-seek.query.ts`
- Create: `src/models/content-seek/content-seek.mutation.ts`

**Step 1: Create query file**

```typescript
// src/models/content-seek/content-seek.query.ts
import "server-only";

import { db } from "@/lib/db";

/** Find a content seek record by course name */
export async function findContentSeekByCourseName(courseName: string) {
  return db.contentSeek.findUnique({
    where: { courseName },
    select: { id: true, count: true },
  });
}
```

**Step 2: Create mutation file**

```typescript
// src/models/content-seek/content-seek.mutation.ts
import "server-only";

import { ulid } from "ulid";
import { db } from "@/lib/db";

type CreateContentSeekData = {
  userId: string;
  courseName: string;
  description?: string;
  diskUrl?: string;
};

/** Create a new content seek record */
export async function createContentSeek(data: CreateContentSeekData) {
  return db.contentSeek.create({
    data: {
      id: ulid(),
      userId: data.userId,
      courseName: data.courseName,
      description: data.description ?? null,
      diskUrl: data.diskUrl ?? null,
    },
    select: { id: true },
  });
}

/** Increment the count of an existing content seek record */
export async function incrementContentSeekCount(id: string) {
  return db.contentSeek.update({
    where: { id },
    data: { count: { increment: 1 } },
    select: { id: true, count: true },
  });
}
```

**Step 3: Commit**

```
feat: add content-seek model layer (query + mutation)
```

---

### Task 3: Zod Schema

**Files:**
- Create: `src/features/web/hall/schemas/content-seek.schema.ts`

**Step 1: Create validation schema**

```typescript
// src/features/web/hall/schemas/content-seek.schema.ts
import { z } from "zod";

/** Schema for submitting a course request */
export const contentSeekSchema = z.object({
  courseName: z
    .string()
    .trim()
    .min(1, "请输入课程名称")
    .max(30, "最多30个字符"),
  description: z
    .string()
    .trim()
    .max(30, "最多30个字符")
    .optional()
    .transform((v) => v || undefined),
  diskUrl: z
    .string()
    .trim()
    .max(30, "最多30个字符")
    .optional()
    .transform((v) => v || undefined),
});

export type ContentSeekInput = z.infer<typeof contentSeekSchema>;
```

**Step 2: Commit**

```
feat: add content-seek zod validation schema
```

---

### Task 4: Service

**Files:**
- Create: `src/features/web/hall/services/content-seek.service.ts`

**Step 1: Create service**

```typescript
// src/features/web/hall/services/content-seek.service.ts
import "server-only";

import { findContentSeekByCourseName } from "@/models/content-seek/content-seek.query";
import { createContentSeek, incrementContentSeekCount } from "@/models/content-seek/content-seek.mutation";

type SubmitContentSeekData = {
  userId: string;
  courseName: string;
  description?: string;
  diskUrl?: string;
};

/** Submit a course request — creates new or increments count on duplicate */
export async function submitContentSeek(data: SubmitContentSeekData) {
  const existing = await findContentSeekByCourseName(data.courseName);

  if (existing) {
    await incrementContentSeekCount(existing.id);
    return { duplicate: true };
  }

  await createContentSeek(data);
  return { success: true };
}
```

**Step 2: Commit**

```
feat: add content-seek service with duplicate detection
```

---

### Task 5: Server Action

**Files:**
- Create: `src/features/web/hall/actions/content-seek.action.ts`

**Step 1: Create action**

Reference pattern: `src/features/web/redeem/actions/redeem.action.ts`

```typescript
// src/features/web/hall/actions/content-seek.action.ts
"use server";

import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { contentSeekSchema } from "@/features/web/hall/schemas/content-seek.schema";
import { submitContentSeek } from "@/features/web/hall/services/content-seek.service";

/** Submit a course request for the current user */
export async function submitContentSeekAction(input: {
  courseName: string;
  description?: string;
  diskUrl?: string;
}) {
  try {
    const profile = await fetchUserProfile();
    if (!profile) {
      return { error: "未登录" };
    }

    const parsed = contentSeekSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "参数错误" };
    }

    const result = await submitContentSeek({
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
feat: add content-seek server action
```

---

### Task 6: Modal Component

**Files:**
- Create: `src/features/web/hall/components/content-seek-modal.tsx`

**Step 1: Create modal**

Reference pattern: `src/features/web/redeem/components/generate-codes-modal.tsx`

Design reference: `design/dx-web.pen` node `vg607` (求课程弹窗) — but simplified to 3 fields: courseName, description, diskUrl.

```tsx
// src/features/web/hall/components/content-seek-modal.tsx
"use client";

import { useState, useTransition } from "react";
import { Lightbulb, Send, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog, DialogContent, DialogTitle, DialogDescription,
} from "@/components/ui/dialog";
import { submitContentSeekAction } from "@/features/web/hall/actions/content-seek.action";

type ContentSeekModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** Modal form for submitting a course request */
export function ContentSeekModal({ open, onOpenChange }: ContentSeekModalProps) {
  const [courseName, setCourseName] = useState("");
  const [description, setDescription] = useState("");
  const [diskUrl, setDiskUrl] = useState("");
  const [isPending, startTransition] = useTransition();

  /** Reset form fields */
  function resetForm() {
    setCourseName("");
    setDescription("");
    setDiskUrl("");
  }

  /** Handle form submission */
  function handleSubmit() {
    if (!courseName.trim()) return;

    startTransition(async () => {
      const result = await submitContentSeekAction({
        courseName: courseName.trim(),
        description: description.trim() || undefined,
        diskUrl: diskUrl.trim() || undefined,
      });

      if ("error" in result) {
        toast.error(result.error);
        return;
      }

      if ("duplicate" in result) {
        toast.info("已有相同申请，正在处理中...");
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
            <Lightbulb className="h-[18px] w-[18px] text-teal-600" />
            求课程
          </DialogTitle>
          <DialogDescription className="sr-only">
            提交您想学习的课程请求
          </DialogDescription>

          <div className="h-px bg-slate-100" />

          <p className="text-sm leading-[1.5] text-slate-500">
            告诉我们您想学什么课程，我们会尽快安排上线！
          </p>

          {/* Course name */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-slate-900">
              课程名称
            </label>
            <input
              type="text"
              placeholder="例如：新概念英语第一册"
              value={courseName}
              onChange={(e) => setCourseName(e.target.value)}
              disabled={isPending}
              maxLength={30}
              className="h-11 rounded-[10px] border border-slate-200 bg-slate-50 px-4 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
            />
          </div>

          {/* Description */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-slate-900">
              补充说明（选填）
            </label>
            <input
              type="text"
              placeholder="例如：希望增加同步练习"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={isPending}
              maxLength={30}
              className="h-11 rounded-[10px] border border-slate-200 bg-slate-50 px-4 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
            />
          </div>

          {/* Disk URL */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-slate-900">
              网盘链接（选填）
            </label>
            <input
              type="text"
              placeholder="例如：百度网盘/阿里云盘链接"
              value={diskUrl}
              onChange={(e) => setDiskUrl(e.target.value)}
              disabled={isPending}
              maxLength={30}
              className="h-11 rounded-[10px] border border-slate-200 bg-slate-50 px-4 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
            />
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
              disabled={isPending || !courseName.trim()}
              className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl bg-teal-600 text-[15px] font-semibold text-white transition-colors hover:bg-teal-700 disabled:opacity-50"
            >
              {isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
              提交请求
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
feat: add content-seek modal component
```

---

### Task 7: Trigger Button + TopActions Integration

**Files:**
- Create: `src/features/web/hall/components/content-seek-button.tsx`
- Modify: `src/features/web/hall/components/top-actions.tsx`

**Step 1: Create the client wrapper button**

```tsx
// src/features/web/hall/components/content-seek-button.tsx
"use client";

import { useState } from "react";
import { Lightbulb } from "lucide-react";
import { ContentSeekModal } from "@/features/web/hall/components/content-seek-modal";

/** Client button that opens the content seek modal */
export function ContentSeekButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="flex h-10 items-center gap-1.5 rounded-[10px] border border-slate-200 bg-white px-3.5 text-slate-500 hover:bg-slate-50"
      >
        <Lightbulb className="h-3.5 w-3.5" />
        <span className="text-[13px] font-semibold">求课程</span>
      </button>
      <ContentSeekModal open={open} onOpenChange={setOpen} />
    </>
  );
}
```

**Step 2: Modify TopActions**

In `src/features/web/hall/components/top-actions.tsx`:

- Add import: `import { ContentSeekButton } from "@/features/web/hall/components/content-seek-button";`
- Replace the static "求课程" button (lines 26-29) with: `<ContentSeekButton />`
- Remove the `Lightbulb` import if no longer used elsewhere

**Step 3: Verify**

Run: `npm run build`
Expected: Build succeeds with no errors.

Run: `npm run dev` and click the "求课程" button in the hall — modal should open.

**Step 4: Commit**

```
feat: integrate content-seek button and modal into top actions
```
