# Redeem Page Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a full redeem code system where admin generates codes and users redeem them to upgrade membership grade.

**Architecture:** Server actions → services → models pattern. Transaction for redeem operation (mark code + update user atomically). Admin-only features gated by `username === "rainson"` on both server and client side.

**Tech Stack:** Next.js 16 App Router, Prisma v7, Zod v4, shadcn/ui, TailwindCSS v4, ulid, crypto.getRandomValues

---

### Task 1: Prisma Schema — Rename RedeemCode to UserRedeem & Rename payDueAt to vipDueAt

**Files:**
- Modify: `prisma/schema/redeem-code.prisma` (rename to `prisma/schema/user-redeem.prisma`)
- Modify: `prisma/schema/user.prisma:16,43`

**Step 1: Rename the schema file and rewrite the model**

Rename `prisma/schema/redeem-code.prisma` → `prisma/schema/user-redeem.prisma`.

Replace the entire contents with:

```prisma
model UserRedeem {
  id         String    @id @db.Char(26)
  code       String    @unique @db.VarChar(19)
  grade      String    @db.VarChar(20)
  userId     String?   @map("user_id") @db.Char(26)
  redeemedAt DateTime? @map("redeemed_at") @db.Timestamptz

  createdAt DateTime @default(now()) @map("created_at") @db.Timestamptz
  updatedAt DateTime @updatedAt @map("updated_at") @db.Timestamptz

  user User? @relation(fields: [userId], references: [id], onDelete: Restrict, onUpdate: Restrict)

  @@index([userId])
  @@index([createdAt])
  @@map("user_redeems")
}
```

**Step 2: Update User model**

In `prisma/schema/user.prisma`:

- Line 16: Change `payDueAt DateTime? @map("pay_due_at")` → `vipDueAt DateTime? @map("vip_due_at")`
- Line 43: Change `redeemedCodes RedeemCode[]` → `redeems UserRedeem[]`

**Step 3: Create and run migration**

```bash
npx prisma migrate dev --name rename-redeem-code-to-user-redeem
```

This will drop the old `redeem_codes` table and create `user_redeems` (no existing data to preserve). Also renames `pay_due_at` → `vip_due_at` in the users table.

If Prisma warns about data loss on the `redeem_codes` table, confirm it — the table is empty.

**Step 4: Regenerate Prisma client**

```bash
npm run prisma:generate
```

**Step 5: Verify build**

```bash
npx tsc --noEmit
```

Expected: should pass (no code references RedeemCode or payDueAt yet beyond the schema).

**Step 6: Commit**

```
feat: rename RedeemCode to UserRedeem and payDueAt to vipDueAt
```

---

### Task 2: User Grade Duration Constants

**Files:**
- Modify: `src/consts/user-grade.ts`

**Step 1: Add grade duration map**

Add at the bottom of `src/consts/user-grade.ts`:

```typescript
/** Number of months each grade adds when redeemed */
export const USER_GRADE_MONTHS: Record<UserGrade, number | null> = {
  free: null,
  month: 1,
  season: 3,
  year: 12,
  lifetime: null,
};
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add USER_GRADE_MONTHS duration constant
```

---

### Task 3: Redeem Code Generation Helper

**Files:**
- Create: `src/features/web/redeem/helpers/redeem-code.helper.ts`

**Step 1: Create the helper**

```typescript
const CHARSET = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
const GROUP_SIZE = 4;
const GROUP_COUNT = 4;

/** Generate a cryptographically random redeem code in XXXX-XXXX-XXXX-XXXX format */
export function generateRedeemCode(): string {
  const totalChars = GROUP_SIZE * GROUP_COUNT;
  const bytes = new Uint8Array(totalChars);
  crypto.getRandomValues(bytes);

  const chars = Array.from(bytes, (b) => CHARSET[b % CHARSET.length]);
  const groups: string[] = [];

  for (let i = 0; i < GROUP_COUNT; i++) {
    groups.push(chars.slice(i * GROUP_SIZE, (i + 1) * GROUP_SIZE).join(""));
  }

  return groups.join("-");
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add redeem code generation helper
```

---

### Task 4: VIP Due At Calculation Helper

**Files:**
- Create: `src/features/web/redeem/helpers/vip-due-at.helper.ts`

**Step 1: Create the helper**

```typescript
import type { UserGrade } from "@/consts/user-grade";
import { USER_GRADE_MONTHS } from "@/consts/user-grade";

/**
 * Add N calendar months to a date, then subtract 1 day.
 * If the target month has fewer days, clamp to the last day.
 * Example: March 17 + 1 month = April 16
 * Example: Jan 31 + 1 month = Feb 28 (or 29)
 */
function addMonthsMinusOneDay(base: Date, months: number): Date {
  const year = base.getFullYear();
  const month = base.getMonth() + months;
  const day = base.getDate();

  // Create a date in the target month with the same day
  const target = new Date(year, month, day);

  // If the day overflowed (e.g., Jan 31 → Mar 3), clamp to last day of target month
  if (target.getMonth() !== ((base.getMonth() + months) % 12 + 12) % 12) {
    // Set to last day of the intended target month
    return new Date(year, month, 0);
  }

  // Subtract 1 day
  target.setDate(target.getDate() - 1);
  return target;
}

/**
 * Calculate the new vipDueAt date after redeeming a code.
 *
 * Rules:
 * - lifetime → return null (never expires)
 * - free/expired → base = today, add grade months
 * - not expired → base = current vipDueAt, add grade months
 */
export function calcVipDueAt(
  grade: UserGrade,
  currentVipDueAt: Date | null,
): Date | null {
  const months = USER_GRADE_MONTHS[grade];

  // Lifetime never expires
  if (months === null) return null;

  const now = new Date();
  const isExpired = !currentVipDueAt || currentVipDueAt < now;
  const base = isExpired ? now : currentVipDueAt;

  return addMonthsMinusOneDay(base, months);
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add VIP due-at date calculation helper
```

---

### Task 5: Zod Validation Schemas

**Files:**
- Create: `src/features/web/redeem/schemas/redeem.schema.ts`

**Step 1: Create the schema file**

```typescript
import { z } from "zod/v4";
import { USER_GRADES } from "@/consts/user-grade";

/** Schema for redeeming a code */
export const redeemCodeSchema = z.object({
  code: z
    .string()
    .trim()
    .toUpperCase()
    .length(19, "兑换码格式不正确")
    .regex(/^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/, "兑换码格式不正确"),
});

export type RedeemCodeInput = z.infer<typeof redeemCodeSchema>;

const redeemableGrades = [
  USER_GRADES.MONTH,
  USER_GRADES.SEASON,
  USER_GRADES.YEAR,
  USER_GRADES.LIFETIME,
] as const;

/** Schema for generating codes (admin only) */
export const generateCodesSchema = z.object({
  grade: z.enum(redeemableGrades, {
    error: "请选择生成类型",
  }),
  quantity: z.enum(["10", "50", "100", "500"], {
    error: "请选择生成数量",
  }).transform(Number),
});

export type GenerateCodesInput = z.infer<typeof generateCodesSchema>;
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add redeem Zod validation schemas
```

---

### Task 6: Model Layer — user-redeem queries and mutations

**Files:**
- Create: `src/models/user-redeem/user-redeem.query.ts`
- Create: `src/models/user-redeem/user-redeem.mutation.ts`
- Modify: `src/models/user/user.mutation.ts`

**Step 1: Create query file**

Create `src/models/user-redeem/user-redeem.query.ts`:

```typescript
import "server-only";

import { db } from "@/lib/db";

/** Default number of redeem records per page */
export const REDEEMS_PAGE_SIZE = 15;

/** Find a redeem record by its code */
export async function findRedeemByCode(code: string) {
  return db.userRedeem.findUnique({
    where: { code },
    select: {
      id: true,
      code: true,
      grade: true,
      userId: true,
      redeemedAt: true,
    },
  });
}

/** Count redeems by a specific user */
export async function countRedeemsByUserId(userId: string) {
  return db.userRedeem.count({
    where: { userId },
  });
}

/** Find redeems by a specific user with pagination */
export async function findRedeemsByUserId(
  userId: string,
  page = 1,
  pageSize = REDEEMS_PAGE_SIZE,
) {
  return db.userRedeem.findMany({
    where: { userId },
    select: {
      id: true,
      code: true,
      grade: true,
      redeemedAt: true,
    },
    orderBy: { redeemedAt: "desc" },
    skip: (page - 1) * pageSize,
    take: pageSize,
  });
}

/** Count all redeem records (admin) */
export async function countAllRedeems() {
  return db.userRedeem.count();
}

/** Find all redeem records with pagination (admin) */
export async function findAllRedeems(
  page = 1,
  pageSize = REDEEMS_PAGE_SIZE,
) {
  return db.userRedeem.findMany({
    select: {
      id: true,
      code: true,
      grade: true,
      userId: true,
      redeemedAt: true,
      createdAt: true,
      user: {
        select: {
          username: true,
          nickname: true,
        },
      },
    },
    orderBy: { createdAt: "desc" },
    skip: (page - 1) * pageSize,
    take: pageSize,
  });
}
```

**Step 2: Create mutation file**

Create `src/models/user-redeem/user-redeem.mutation.ts`:

```typescript
import "server-only";

import { ulid } from "ulid";
import { db, Prisma } from "@/lib/db";

type TxClient = Prisma.TransactionClient;

type CreateRedeemData = {
  code: string;
  grade: string;
};

/** Bulk-create redeem codes */
export async function createRedeemCodes(items: CreateRedeemData[]) {
  return db.userRedeem.createMany({
    data: items.map((item) => ({
      id: ulid(),
      code: item.code,
      grade: item.grade,
    })),
  });
}

/** Mark a redeem code as used by a user (within transaction) */
export async function markRedeemUsed(
  redeemId: string,
  userId: string,
  tx: TxClient,
) {
  return tx.userRedeem.update({
    where: { id: redeemId },
    data: {
      userId,
      redeemedAt: new Date(),
    },
    select: { id: true },
  });
}
```

**Step 3: Add updateUserVip to user mutations**

Add to the bottom of `src/models/user/user.mutation.ts`:

```typescript
/** Update a user's grade and VIP expiration (within transaction) */
export async function updateUserVip(
  userId: string,
  grade: string,
  vipDueAt: Date | null,
  tx: TxClient,
) {
  return tx.user.update({
    where: { id: userId },
    data: { grade, vipDueAt },
    select: { id: true, grade: true, vipDueAt: true },
  });
}
```

Also add `TxClient` type import at the top if not already present (it is already defined on line 6).

**Step 4: Verify build**

```bash
npx tsc --noEmit
```

**Step 5: Commit**

```
feat: add user-redeem model layer and updateUserVip mutation
```

---

### Task 7: Redeem Service — Business Logic

**Files:**
- Create: `src/features/web/redeem/services/redeem.service.ts`

**Step 1: Create the service**

```typescript
import "server-only";

import { db } from "@/lib/db";
import type { UserGrade } from "@/consts/user-grade";
import { findRedeemByCode } from "@/models/user-redeem/user-redeem.query";
import { createRedeemCodes, markRedeemUsed } from "@/models/user-redeem/user-redeem.mutation";
import { updateUserVip } from "@/models/user/user.mutation";
import { generateRedeemCode } from "@/features/web/redeem/helpers/redeem-code.helper";
import { calcVipDueAt } from "@/features/web/redeem/helpers/vip-due-at.helper";

/** Generate a batch of redeem codes for a given grade */
export async function generateCodes(grade: string, quantity: number) {
  const codes: { code: string; grade: string }[] = [];
  const usedCodes = new Set<string>();

  for (let i = 0; i < quantity; i++) {
    let code: string;
    do {
      code = generateRedeemCode();
    } while (usedCodes.has(code));
    usedCodes.add(code);
    codes.push({ code, grade });
  }

  await createRedeemCodes(codes);
  return { count: codes.length };
}

/** Redeem a code for a user — updates grade and vipDueAt atomically */
export async function redeemCode(
  userId: string,
  code: string,
  currentGrade: string,
  currentVipDueAt: Date | null,
) {
  const record = await findRedeemByCode(code);

  if (!record) {
    return { error: "兑换码不存在" };
  }

  if (record.userId) {
    return { error: "该兑换码已被使用" };
  }

  const redeemGrade = record.grade as UserGrade;
  const newVipDueAt = calcVipDueAt(redeemGrade, currentVipDueAt);

  await db.$transaction(async (tx) => {
    await markRedeemUsed(record.id, userId, tx);
    await updateUserVip(userId, redeemGrade, newVipDueAt, tx);
  });

  return { success: true, grade: redeemGrade };
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add redeem service with generate and redeem logic
```

---

### Task 8: Server Actions

**Files:**
- Create: `src/features/web/redeem/actions/redeem.action.ts`

**Step 1: Create the actions file**

```typescript
"use server";

import { auth } from "@/lib/auth";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { generateCodesSchema, redeemCodeSchema } from "@/features/web/redeem/schemas/redeem.schema";
import { generateCodes, redeemCode } from "@/features/web/redeem/services/redeem.service";
import {
  countRedeemsByUserId,
  findRedeemsByUserId,
  countAllRedeems,
  findAllRedeems,
  REDEEMS_PAGE_SIZE,
} from "@/models/user-redeem/user-redeem.query";

const ADMIN_USERNAME = "rainson";

/** Generate redeem codes (admin only) */
export async function generateCodesAction(input: { grade: string; quantity: string }) {
  try {
    const session = await auth();
    if (!session?.user?.name || session.user.name !== ADMIN_USERNAME) {
      return { error: "无权限" };
    }

    const parsed = generateCodesSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "参数错误" };
    }

    const result = await generateCodes(parsed.data.grade, parsed.data.quantity);
    return { data: result };
  } catch {
    return { error: "生成兑换码失败" };
  }
}

/** Redeem a code for the current user */
export async function redeemCodeAction(input: { code: string }) {
  try {
    const profile = await fetchUserProfile();
    if (!profile) {
      return { error: "未登录" };
    }

    const parsed = redeemCodeSchema.safeParse(input);
    if (!parsed.success) {
      return { error: parsed.error.issues[0]?.message ?? "兑换码格式不正确" };
    }

    // Fetch current vipDueAt for the user
    const { auth: authFn } = await import("@/lib/auth");
    const session = await authFn();
    if (!session?.user?.id) return { error: "未登录" };

    // We need the user's current vipDueAt — fetch from db directly
    const { db } = await import("@/lib/db");
    const user = await db.user.findUnique({
      where: { id: profile.id },
      select: { grade: true, vipDueAt: true },
    });

    if (!user) return { error: "用户不存在" };

    const result = await redeemCode(
      profile.id,
      parsed.data.code,
      user.grade,
      user.vipDueAt,
    );

    return result;
  } catch {
    return { error: "兑换失败" };
  }
}

/** Fetch a page of the current user's redeem records */
export async function fetchUserRedeemsAction(page: number) {
  try {
    const profile = await fetchUserProfile();
    if (!profile) {
      return { error: "未登录" };
    }

    const safePage = Math.max(1, Math.floor(page));

    const [redeems, totalCount] = await Promise.all([
      findRedeemsByUserId(profile.id, safePage),
      countRedeemsByUserId(profile.id),
    ]);

    const totalPages = Math.ceil(totalCount / REDEEMS_PAGE_SIZE);
    return { data: { redeems, totalPages } };
  } catch {
    return { error: "获取兑换记录失败" };
  }
}

/** Fetch a page of all redeem records (admin only) */
export async function fetchAllRedeemsAction(page: number) {
  try {
    const session = await auth();
    if (!session?.user?.name || session.user.name !== ADMIN_USERNAME) {
      return { error: "无权限" };
    }

    const safePage = Math.max(1, Math.floor(page));

    const [redeems, totalCount] = await Promise.all([
      findAllRedeems(safePage),
      countAllRedeems(),
    ]);

    const totalPages = Math.ceil(totalCount / REDEEMS_PAGE_SIZE);
    return { data: { redeems, totalPages } };
  } catch {
    return { error: "获取兑换码记录失败" };
  }
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add redeem server actions
```

---

### Task 9: Redeem Page — Server Component Setup

**Files:**
- Modify: `src/app/(web)/hall/(main)/redeem/page.tsx`

**Step 1: Update the page to fetch initial data and pass username**

```typescript
import { auth } from "@/lib/auth";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { RedeemContent } from "@/features/web/redeem/components/redeem-content";
import {
  countRedeemsByUserId,
  findRedeemsByUserId,
  countAllRedeems,
  findAllRedeems,
  REDEEMS_PAGE_SIZE,
} from "@/models/user-redeem/user-redeem.query";

export default async function RedeemPage() {
  const profile = await fetchUserProfile();
  const session = await auth();
  const username = session?.user?.name ?? null;
  const isAdmin = username === "rainson";

  // Fetch user's redeem history (page 1)
  const [userRedeems, userTotalCount] = profile
    ? await Promise.all([
        findRedeemsByUserId(profile.id, 1),
        countRedeemsByUserId(profile.id),
      ])
    : [[], 0];

  const userTotalPages = Math.ceil(userTotalCount / REDEEMS_PAGE_SIZE);

  // Fetch all redeems for admin (page 1)
  const [allRedeems, allTotalCount] = isAdmin
    ? await Promise.all([findAllRedeems(1), countAllRedeems()])
    : [[], 0];

  const allTotalPages = Math.ceil(allTotalCount / REDEEMS_PAGE_SIZE);

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="兑换码"
        subtitle="输入兑换码激活会员权益"
      />
      <RedeemContent
        username={username}
        initialUserRedeems={userRedeems}
        initialUserTotalPages={userTotalPages}
        initialAllRedeems={allRedeems}
        initialAllTotalPages={allTotalPages}
      />
    </div>
  );
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

Expected: will fail because RedeemContent doesn't accept these props yet. That's OK — we fix it in the next task.

**Step 3: Commit**

```
feat: update redeem page with server-side data fetching
```

---

### Task 10: Redeem Input Card Component

**Files:**
- Create: `src/features/web/redeem/components/redeem-input-card.tsx`

**Step 1: Create the component**

```typescript
"use client";

import { useState, useTransition } from "react";
import { Ticket, ShoppingCart, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { redeemCodeAction } from "@/features/web/redeem/actions/redeem.action";

type RedeemInputCardProps = {
  onRedeemed: () => void;
};

/** Card with code input field and redeem button */
export function RedeemInputCard({ onRedeemed }: RedeemInputCardProps) {
  const [code, setCode] = useState("");
  const [isPending, startTransition] = useTransition();

  /** Handle redeem submission */
  const handleRedeem = () => {
    if (!code.trim()) return;

    startTransition(async () => {
      const result = await redeemCodeAction({ code: code.trim() });

      if ("error" in result) {
        toast.error(result.error);
        return;
      }

      toast.success("兑换成功");
      setCode("");
      onRedeemed();
    });
  };

  return (
    <div className="flex flex-col gap-5 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm md:p-7">
      <div className="flex flex-col gap-1.5">
        <span className="text-base font-semibold text-slate-900">输入兑换码</span>
        <span className="text-[13px] text-slate-400">
          兑换码区分大小写，请按格式正确输入（如：ABCD-1234-EFGH-5678）
        </span>
      </div>
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="flex h-12 flex-1 items-center rounded-[10px] border border-slate-200 bg-slate-50 px-4">
          <Ticket className="mr-2.5 h-[18px] w-[18px] text-slate-400" />
          <input
            type="text"
            placeholder="请输入兑换码"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            disabled={isPending}
            maxLength={19}
            className="flex-1 bg-transparent text-[13px] text-slate-900 outline-none placeholder:text-slate-400 disabled:opacity-50"
          />
        </div>
        <button
          type="button"
          onClick={handleRedeem}
          disabled={isPending || !code.trim()}
          className="flex h-12 items-center justify-center gap-1.5 rounded-[10px] bg-teal-600 px-6 text-sm font-semibold text-white hover:bg-teal-700 disabled:opacity-50"
        >
          {isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          立即兑换
        </button>
      </div>
      <div className="flex items-center justify-center gap-2">
        <span className="text-[13px] text-slate-400">没有兑换码？</span>
        <button
          type="button"
          className="flex items-center gap-1.5 rounded-lg bg-teal-50 px-4 py-2 text-[13px] font-medium text-teal-600"
        >
          <ShoppingCart className="h-3.5 w-3.5" />
          购买会员
        </button>
      </div>
    </div>
  );
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add redeem input card component
```

---

### Task 11: Redeem History Table Component (User's Own Records)

**Files:**
- Create: `src/features/web/redeem/components/redeem-history-table.tsx`

**Step 1: Create the component**

```typescript
"use client";

import { useState, useTransition } from "react";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { USER_GRADE_LABELS } from "@/consts/user-grade";
import type { UserGrade } from "@/consts/user-grade";
import { fetchUserRedeemsAction } from "@/features/web/redeem/actions/redeem.action";
import type { findRedeemsByUserId } from "@/models/user-redeem/user-redeem.query";

type Redeem = Awaited<ReturnType<typeof findRedeemsByUserId>>[number];

type RedeemHistoryTableProps = {
  initialRedeems: Redeem[];
  initialTotalPages: number;
  refreshKey: number;
};

/** Table displaying the current user's redeem records with pagination */
export function RedeemHistoryTable({
  initialRedeems,
  initialTotalPages,
  refreshKey,
}: RedeemHistoryTableProps) {
  const [redeems, setRedeems] = useState(initialRedeems);
  const [totalPages, setTotalPages] = useState(initialTotalPages);
  const [currentPage, setCurrentPage] = useState(1);
  const [isPending, startTransition] = useTransition();

  /** Reload when refreshKey changes (after a redeem) */
  const [prevKey, setPrevKey] = useState(refreshKey);
  if (refreshKey !== prevKey) {
    setPrevKey(refreshKey);
    startTransition(async () => {
      const result = await fetchUserRedeemsAction(1);
      if ("data" in result && result.data) {
        setRedeems(result.data.redeems);
        setTotalPages(result.data.totalPages);
        setCurrentPage(1);
      }
    });
  }

  /** Load a specific page */
  const handlePageChange = (page: number) => {
    startTransition(async () => {
      const result = await fetchUserRedeemsAction(page);
      if ("data" in result && result.data) {
        setRedeems(result.data.redeems);
        setTotalPages(result.data.totalPages);
        setCurrentPage(page);
      }
    });
  };

  /** Format date for display */
  const formatDate = (date: Date | null) => {
    if (!date) return "-";
    return new Date(date).toLocaleDateString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  };

  return (
    <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div className="flex items-center justify-between px-4 py-5 md:px-6">
        <span className="text-base font-semibold text-slate-900">兑换记录</span>
      </div>
      <div className="h-px w-full bg-slate-100" />

      <div className={isPending ? "opacity-60 transition-opacity" : ""}>
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow className="bg-slate-50 hover:bg-slate-50">
                <TableHead className="pl-6 text-xs font-semibold text-slate-500">兑换码</TableHead>
                <TableHead className="w-[140px] text-xs font-semibold text-slate-500">兑换等级</TableHead>
                <TableHead className="w-[140px] text-xs font-semibold text-slate-500">兑换时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {redeems.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="py-10 text-center text-sm text-slate-400">
                    暂无兑换记录
                  </TableCell>
                </TableRow>
              )}
              {redeems.map((redeem) => (
                <TableRow key={redeem.id}>
                  <TableCell className="pl-6 font-mono text-[13px] font-medium text-slate-700">
                    {redeem.code}
                  </TableCell>
                  <TableCell className="text-[13px] text-slate-600">
                    {USER_GRADE_LABELS[redeem.grade as UserGrade] ?? redeem.grade}
                  </TableCell>
                  <TableCell className="text-[13px] text-slate-400">
                    {formatDate(redeem.redeemedAt)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>

        <DataTablePagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={handlePageChange}
        />
      </div>
    </div>
  );
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add redeem history table component
```

---

### Task 12: Generate Codes Modal Component

**Files:**
- Create: `src/features/web/redeem/components/generate-codes-modal.tsx`

**Step 1: Create the component**

```typescript
"use client";

import { useState } from "react";
import { Loader2, TicketPlus } from "lucide-react";
import {
  Dialog, DialogContent, DialogTitle, DialogDescription,
} from "@/components/ui/dialog";
import { USER_GRADE_LABELS } from "@/consts/user-grade";

const GRADE_OPTIONS = [
  { value: "month", label: USER_GRADE_LABELS.month },
  { value: "season", label: USER_GRADE_LABELS.season },
  { value: "year", label: USER_GRADE_LABELS.year },
  { value: "lifetime", label: USER_GRADE_LABELS.lifetime },
];

const QUANTITY_OPTIONS = [
  { value: "10", label: "10" },
  { value: "50", label: "50" },
  { value: "100", label: "100" },
  { value: "500", label: "500" },
];

type GenerateCodesModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onGenerate: (input: { grade: string; quantity: string }) => Promise<boolean>;
};

/** Modal form for generating redeem codes (admin only) */
export function GenerateCodesModal({ open, onOpenChange, onGenerate }: GenerateCodesModalProps) {
  const [grade, setGrade] = useState(GRADE_OPTIONS[0].value);
  const [quantity, setQuantity] = useState(QUANTITY_OPTIONS[0].value);
  const [submitting, setSubmitting] = useState(false);

  /** Handle generation submit */
  async function handleSubmit() {
    setSubmitting(true);
    const ok = await onGenerate({ grade, quantity });
    setSubmitting(false);
    if (ok) {
      onOpenChange(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton
        className="max-w-[520px] gap-0 rounded-[20px] border-none p-0"
      >
        <div className="flex flex-col gap-5 px-7 pt-7 pb-6">
          <DialogTitle className="flex items-center gap-2.5 text-xl font-bold text-slate-900">
            <TicketPlus className="h-[18px] w-[18px] text-teal-600" />
            生成兑换码
          </DialogTitle>
          <DialogDescription className="sr-only">
            选择兑换码类型和数量
          </DialogDescription>

          <div className="h-px bg-slate-100" />

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <label htmlFor="generate-grade" className="text-[13px] font-medium text-slate-700">
                生成类型 *
              </label>
              <select
                id="generate-grade"
                value={grade}
                onChange={(e) => setGrade(e.target.value)}
                disabled={submitting}
                className="h-10 rounded-lg border border-slate-200 bg-white px-3.5 text-[13px] text-slate-900 outline-none transition-colors focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              >
                {GRADE_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-2">
              <label htmlFor="generate-quantity" className="text-[13px] font-medium text-slate-700">
                生成数量 *
              </label>
              <select
                id="generate-quantity"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                disabled={submitting}
                className="h-10 rounded-lg border border-slate-200 bg-white px-3.5 text-[13px] text-slate-900 outline-none transition-colors focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              >
                {QUANTITY_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex justify-end gap-2.5">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
              className="rounded-lg border border-slate-200 px-4 py-2 text-[13px] font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={submitting}
              className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-4 py-2 text-[13px] font-medium text-white transition-colors hover:bg-teal-700 disabled:opacity-50"
            >
              {submitting && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              确认生成
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add generate codes modal component
```

---

### Task 13: Admin Section Component (Generate Button + All Codes Table)

**Files:**
- Create: `src/features/web/redeem/components/redeem-admin-section.tsx`

**Step 1: Create the component**

```typescript
"use client";

import { useState, useTransition } from "react";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { TicketPlus } from "lucide-react";
import { toast } from "sonner";
import { USER_GRADE_LABELS } from "@/consts/user-grade";
import type { UserGrade } from "@/consts/user-grade";
import { generateCodesAction, fetchAllRedeemsAction } from "@/features/web/redeem/actions/redeem.action";
import { GenerateCodesModal } from "@/features/web/redeem/components/generate-codes-modal";
import type { findAllRedeems } from "@/models/user-redeem/user-redeem.query";

type AdminRedeem = Awaited<ReturnType<typeof findAllRedeems>>[number];

type RedeemAdminSectionProps = {
  initialRedeems: AdminRedeem[];
  initialTotalPages: number;
};

/** Admin-only section: generate button + all codes data table */
export function RedeemAdminSection({
  initialRedeems,
  initialTotalPages,
}: RedeemAdminSectionProps) {
  const [redeems, setRedeems] = useState(initialRedeems);
  const [totalPages, setTotalPages] = useState(initialTotalPages);
  const [currentPage, setCurrentPage] = useState(1);
  const [isPending, startTransition] = useTransition();
  const [modalOpen, setModalOpen] = useState(false);

  /** Generate codes and refresh the table */
  const handleGenerate = async (input: { grade: string; quantity: string }): Promise<boolean> => {
    const result = await generateCodesAction(input);

    if ("error" in result) {
      toast.error(result.error);
      return false;
    }

    toast.success(`成功生成 ${input.quantity} 个兑换码`);

    // Refresh table to page 1
    startTransition(async () => {
      const res = await fetchAllRedeemsAction(1);
      if ("data" in res && res.data) {
        setRedeems(res.data.redeems);
        setTotalPages(res.data.totalPages);
        setCurrentPage(1);
      }
    });

    return true;
  };

  /** Load a specific page */
  const handlePageChange = (page: number) => {
    startTransition(async () => {
      const result = await fetchAllRedeemsAction(page);
      if ("data" in result && result.data) {
        setRedeems(result.data.redeems);
        setTotalPages(result.data.totalPages);
        setCurrentPage(page);
      }
    });
  };

  /** Format date for display */
  const formatDate = (date: Date | null) => {
    if (!date) return "-";
    return new Date(date).toLocaleDateString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  };

  /** Get display name for a redeemed user */
  const getUserDisplay = (redeem: AdminRedeem) => {
    if (!redeem.user) return "-";
    return redeem.user.nickname ?? redeem.user.username;
  };

  return (
    <>
      <button
        type="button"
        onClick={() => setModalOpen(true)}
        className="flex items-center gap-1.5 self-start rounded-lg bg-teal-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-teal-700"
      >
        <TicketPlus className="h-4 w-4" />
        生成兑换码
      </button>

      <GenerateCodesModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        onGenerate={handleGenerate}
      />

      <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
        <div className="flex items-center justify-between px-4 py-5 md:px-6">
          <span className="text-base font-semibold text-slate-900">兑换码管理</span>
        </div>
        <div className="h-px w-full bg-slate-100" />

        <div className={isPending ? "opacity-60 transition-opacity" : ""}>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-slate-50 hover:bg-slate-50">
                  <TableHead className="pl-6 text-xs font-semibold text-slate-500">兑换码</TableHead>
                  <TableHead className="w-[100px] text-xs font-semibold text-slate-500">等级</TableHead>
                  <TableHead className="w-[80px] text-xs font-semibold text-slate-500">状态</TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-slate-500">兑换用户</TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-slate-500">兑换时间</TableHead>
                  <TableHead className="w-[120px] text-xs font-semibold text-slate-500">创建时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {redeems.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="py-10 text-center text-sm text-slate-400">
                      暂无兑换码
                    </TableCell>
                  </TableRow>
                )}
                {redeems.map((redeem) => (
                  <TableRow key={redeem.id}>
                    <TableCell className="pl-6 font-mono text-[13px] font-medium text-slate-700">
                      {redeem.code}
                    </TableCell>
                    <TableCell className="text-[13px] text-slate-600">
                      {USER_GRADE_LABELS[redeem.grade as UserGrade] ?? redeem.grade}
                    </TableCell>
                    <TableCell>
                      <span
                        className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${
                          redeem.userId
                            ? "bg-teal-600/10 text-teal-600"
                            : "bg-amber-500/10 text-amber-600"
                        }`}
                      >
                        {redeem.userId ? "已兑换" : "未使用"}
                      </span>
                    </TableCell>
                    <TableCell className="text-[13px] text-slate-600">
                      {getUserDisplay(redeem)}
                    </TableCell>
                    <TableCell className="text-[13px] text-slate-400">
                      {formatDate(redeem.redeemedAt)}
                    </TableCell>
                    <TableCell className="text-[13px] text-slate-400">
                      {formatDate(redeem.createdAt)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          <DataTablePagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </div>
      </div>
    </>
  );
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

**Step 3: Commit**

```
feat: add redeem admin section with generate button and codes table
```

---

### Task 14: Redeem Content — Main Orchestrating Component

**Files:**
- Modify: `src/features/web/redeem/components/redeem-content.tsx` (full rewrite)

**Step 1: Rewrite the component**

```typescript
"use client";

import { useState } from "react";
import { RedeemInputCard } from "@/features/web/redeem/components/redeem-input-card";
import { RedeemHistoryTable } from "@/features/web/redeem/components/redeem-history-table";
import { RedeemAdminSection } from "@/features/web/redeem/components/redeem-admin-section";
import type { findRedeemsByUserId, findAllRedeems } from "@/models/user-redeem/user-redeem.query";

type UserRedeem = Awaited<ReturnType<typeof findRedeemsByUserId>>[number];
type AdminRedeem = Awaited<ReturnType<typeof findAllRedeems>>[number];

type RedeemContentProps = {
  username: string | null;
  initialUserRedeems: UserRedeem[];
  initialUserTotalPages: number;
  initialAllRedeems: AdminRedeem[];
  initialAllTotalPages: number;
};

/** Main redeem page content orchestrating all sections */
export function RedeemContent({
  username,
  initialUserRedeems,
  initialUserTotalPages,
  initialAllRedeems,
  initialAllTotalPages,
}: RedeemContentProps) {
  const isAdmin = username === "rainson";
  const [refreshKey, setRefreshKey] = useState(0);

  /** Called after a successful redeem to refresh history tables */
  const handleRedeemed = () => {
    setRefreshKey((k) => k + 1);
  };

  return (
    <>
      {isAdmin && (
        <RedeemAdminSection
          initialRedeems={initialAllRedeems}
          initialTotalPages={initialAllTotalPages}
        />
      )}

      <RedeemInputCard onRedeemed={handleRedeemed} />

      <RedeemHistoryTable
        initialRedeems={initialUserRedeems}
        initialTotalPages={initialUserTotalPages}
        refreshKey={refreshKey}
      />
    </>
  );
}
```

**Step 2: Verify build**

```bash
npx tsc --noEmit
```

Expected: PASS — all components and types should align.

**Step 3: Run dev server and verify visually**

```bash
npm run dev
```

Navigate to `http://localhost:3000/hall/redeem` and verify:
- Input card renders with working input field
- 兑换记录 table renders (empty if no records)
- Admin section visible only when logged in as rainson
- Generate modal opens and submits
- Redeem flow works end-to-end

**Step 4: Commit**

```
feat: rewrite redeem content with full redeem flow
```

---

### Task 15: Final Verification & Cleanup

**Step 1: Run lint**

```bash
npm run lint
```

Fix any lint issues.

**Step 2: Run build**

```bash
npm run build
```

Expected: PASS — production build should succeed.

**Step 3: Manual testing checklist**

1. As normal user: enter valid code → grade updates, vipDueAt set, record shows in history
2. As normal user: enter invalid code → error toast
3. As normal user: enter already-used code → error toast "该兑换码已被使用"
4. As rainson: generate 10 month codes → success toast, table shows new codes
5. As rainson: all codes table pagination works
6. As non-rainson user: generate button and admin table are not visible
7. Lifetime redeem: vipDueAt set to null, grade set to lifetime

**Step 4: Commit**

```
chore: final cleanup for redeem page implementation
```
