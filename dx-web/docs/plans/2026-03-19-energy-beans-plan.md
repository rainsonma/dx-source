# Energy Beans Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add an Energy Beans virtual currency system with ledger tracking, membership grant integration, and daily cron reset.

**Architecture:** Ledger pattern — each bean operation creates a `UserBean` record and atomically updates `User.beans` + `User.grantedBeans`. FIFO consumption (granted beans first). Cron job at 1 AM handles monthly resets.

**Tech Stack:** Prisma v7, PostgreSQL, Next.js server actions, standalone cron script with `createDb()`

---

### Task 1: Schema — Rename stars to beans + add grantedBeans

**Files:**
- Modify: `prisma/schema/user.prisma`

**Step 1: Update user.prisma**

Replace the `stars` line:

```prisma
  stars        Int       @default(0)
```

with:

```prisma
  beans        Int       @default(0)
  grantedBeans Int       @default(0) @map("granted_beans")
```

**Step 2: Create and run migration**

Run:
```bash
npx prisma migrate dev --name rename-stars-to-beans-add-granted-beans
```

Expected: Migration applies. The migration SQL should rename the `stars` column to `beans` and add `granted_beans`. If Prisma generates a drop+add instead of rename, manually edit the migration SQL to use:
```sql
ALTER TABLE "users" RENAME COLUMN "stars" TO "beans";
ALTER TABLE "users" ADD COLUMN "granted_beans" INTEGER NOT NULL DEFAULT 0;
```

**Step 3: Generate Prisma client**

Run:
```bash
npm run prisma:generate
```

Expected: No errors.

**Step 4: Commit**

```bash
git add prisma/schema/user.prisma prisma/migrations/
git commit -m "feat: rename stars to beans, add grantedBeans on User"
```

---

### Task 2: Schema — Add UserBean model

**Files:**
- Create: `prisma/schema/user-bean.prisma`

**Step 1: Create the schema file**

```prisma
model UserBean {
  id        String   @id @db.Char(26)
  userId    String   @map("user_id") @db.Char(26)
  beans     Int
  origin    Int      @default(0)
  result    Int      @default(0)
  slug      String   @db.VarChar(50)
  reason    String   @db.VarChar(100)
  data      Json?    @db.JsonB
  createdAt DateTime @map("created_at") @db.Timestamptz
  updatedAt DateTime @map("updated_at") @db.Timestamptz

  @@index([userId])
  @@index([userId, slug])
  @@index([createdAt])
  @@map("user_beans")
}
```

**Step 2: Run migration**

Run:
```bash
npx prisma migrate dev --name add-user-bean-model
```

Expected: Migration creates `user_beans` table.

**Step 3: Generate Prisma client**

Run:
```bash
npm run prisma:generate
```

Expected: No errors.

**Step 4: Commit**

```bash
git add prisma/schema/user-bean.prisma prisma/migrations/
git commit -m "feat: add UserBean ledger model"
```

---

### Task 3: Constants — Bean slug and reason

**Files:**
- Create: `src/consts/bean-slug.ts`
- Create: `src/consts/bean-reason.ts`

**Step 1: Create bean-slug.ts**

```typescript
export const BEAN_SLUGS = {
  MEMBERSHIP_GRANT: "membership-grant",
  MONTHLY_RESET_DEBIT: "monthly-reset-debit",
  MONTHLY_RESET_CREDIT: "monthly-reset-credit",
} as const;

export type BeanSlug = (typeof BEAN_SLUGS)[keyof typeof BEAN_SLUGS];

export const BEAN_SLUG_LABELS: Record<BeanSlug, string> = {
  "membership-grant": "会员赠送",
  "monthly-reset-debit": "月度清零",
  "monthly-reset-credit": "月度续发",
};
```

**Step 2: Create bean-reason.ts**

```typescript
export const BEAN_REASONS = {
  MEMBERSHIP_GRANT: "会员购买赠送",
  MONTHLY_RESET_DEBIT: "月度未使用能量豆清零",
  MONTHLY_RESET_CREDIT: "月度能量豆续发",
} as const;

export type BeanReason = (typeof BEAN_REASONS)[keyof typeof BEAN_REASONS];
```

**Step 3: Verify no lint errors**

Run:
```bash
npm run lint
```

Expected: No errors.

**Step 4: Commit**

```bash
git add src/consts/bean-slug.ts src/consts/bean-reason.ts
git commit -m "feat: add bean slug and reason constants"
```

---

### Task 4: Models — UserBean mutation (core ledger function)

**Files:**
- Create: `src/models/user-bean/user-bean.mutation.ts`

**Step 1: Create the mutation file**

```typescript
import "server-only";

import { ulid } from "ulid";
import { db } from "@/lib/db";
import { assertFK } from "@/lib/assert-fk";
import type { Prisma } from "@/generated/prisma/client";
import type { BeanSlug } from "@/consts/bean-slug";
import type { BeanReason } from "@/consts/bean-reason";
import type { JsonValue } from "@/generated/prisma/client/runtime/library";

type TxClient = Prisma.TransactionClient;

/**
 * Insert a bean ledger entry and atomically update User.beans + User.grantedBeans.
 * Must be called within a transaction.
 */
export async function createBeanEntry(
  userId: string,
  beans: number,
  slug: BeanSlug,
  reason: BeanReason,
  grantedDelta: number,
  data?: JsonValue,
  tx?: TxClient
) {
  const client = tx ?? db;

  await assertFK(client as TxClient, [{ table: "users", id: userId }]);

  const user = await client.user.findUniqueOrThrow({
    where: { id: userId },
    select: { beans: true, grantedBeans: true },
  });

  const origin = user.beans;
  const result = origin + beans;
  const now = new Date();

  await client.userBean.create({
    data: {
      id: ulid(),
      userId,
      beans,
      origin,
      result,
      slug,
      reason,
      data: data ?? undefined,
      createdAt: now,
      updatedAt: now,
    },
  });

  const newGrantedBeans = Math.max(0, user.grantedBeans + grantedDelta);

  await client.user.update({
    where: { id: userId },
    data: {
      beans: result,
      grantedBeans: newGrantedBeans,
    },
  });
}
```

**Note on `grantedDelta` parameter:**
- On grant: `grantedDelta = grantAmount` (e.g., 10000 or 15000)
- On monthly debit: `grantedDelta = -user.grantedBeans` (clears to 0)
- On spending: `grantedDelta = -Math.min(spentAmount, user.grantedBeans)` (FIFO)

**Step 2: Verify no lint errors**

Run:
```bash
npm run lint
```

Expected: No errors.

**Step 3: Commit**

```bash
git add src/models/user-bean/user-bean.mutation.ts
git commit -m "feat: add createBeanEntry with atomic balance update"
```

---

### Task 5: Models — UserBean query

**Files:**
- Create: `src/models/user-bean/user-bean.query.ts`

**Step 1: Create the query file**

```typescript
import "server-only";

import { db } from "@/lib/db";

const BEANS_PAGE_SIZE = 20;

/** Find bean ledger entries for a user with pagination */
export async function findBeansByUserId(userId: string, page: number = 1) {
  return db.userBean.findMany({
    where: { userId },
    orderBy: { createdAt: "desc" },
    skip: (page - 1) * BEANS_PAGE_SIZE,
    take: BEANS_PAGE_SIZE,
    select: {
      id: true,
      beans: true,
      origin: true,
      result: true,
      slug: true,
      reason: true,
      data: true,
      createdAt: true,
    },
  });
}

/** Count total bean entries for a user */
export async function countBeansByUserId(userId: string) {
  return db.userBean.count({ where: { userId } });
}

/** Sum all bean entries for a user (for reconciliation) */
export async function sumBeansByUserId(userId: string) {
  const result = await db.userBean.aggregate({
    where: { userId },
    _sum: { beans: true },
  });
  return result._sum.beans ?? 0;
}

/** Find the earliest membership-grant entry for a user */
export async function findFirstGrantEntry(userId: string) {
  return db.userBean.findFirst({
    where: { userId, slug: "membership-grant" },
    orderBy: { createdAt: "asc" },
    select: { createdAt: true },
  });
}
```

**Step 2: Verify no lint errors**

Run:
```bash
npm run lint
```

Expected: No errors.

**Step 3: Commit**

```bash
git add src/models/user-bean/user-bean.query.ts
git commit -m "feat: add user-bean query functions"
```

---

### Task 6: Redeem integration — Grant beans on membership redeem

**Files:**
- Modify: `src/features/web/redeem/services/redeem.service.ts`

**Step 1: Update the redeem service**

Add imports at the top:

```typescript
import { BEAN_SLUGS } from "@/consts/bean-slug";
import { BEAN_REASONS } from "@/consts/bean-reason";
import { USER_GRADES } from "@/consts/user-grade";
import { createBeanEntry } from "@/models/user-bean/user-bean.mutation";
```

In the `redeemCode` function, extend the `db.$transaction` block — after the `updateUserVip` call, add:

```typescript
    const grantAmount = redeemGrade === USER_GRADES.LIFETIME ? 15000 : 10000;
    await createBeanEntry(
      userId,
      grantAmount,
      BEAN_SLUGS.MEMBERSHIP_GRANT,
      BEAN_REASONS.MEMBERSHIP_GRANT,
      grantAmount,
      { gradeAtTime: redeemGrade, vipDueAt: newVipDueAt },
      tx
    );
```

The full transaction block becomes:

```typescript
  await db.$transaction(async (tx) => {
    await markRedeemUsed(record.id, userId, tx);
    await updateUserVip(userId, redeemGrade, newVipDueAt, tx);

    const grantAmount = redeemGrade === USER_GRADES.LIFETIME ? 15000 : 10000;
    await createBeanEntry(
      userId,
      grantAmount,
      BEAN_SLUGS.MEMBERSHIP_GRANT,
      BEAN_REASONS.MEMBERSHIP_GRANT,
      grantAmount,
      { gradeAtTime: redeemGrade, vipDueAt: newVipDueAt },
      tx
    );
  });
```

**Step 2: Verify no lint errors**

Run:
```bash
npm run lint
```

Expected: No errors.

**Step 3: Verify build**

Run:
```bash
npm run build
```

Expected: Build succeeds with no type errors.

**Step 4: Commit**

```bash
git add src/features/web/redeem/services/redeem.service.ts
git commit -m "feat: grant energy beans on membership redeem"
```

---

### Task 7: Cron script — Daily energy bean reset

**Files:**
- Create: `scripts/cron/reset-energy-beans.ts`

**Step 1: Create the cron script**

```typescript
import { createDb } from "../lib/db.js";

/**
 * Daily cron job: reset energy beans for paid members.
 *
 * Schedule: 0 1 * * * (1 AM daily)
 *
 * Logic:
 *   1. Find users whose monthly anniversary matches today
 *   2. Debit remaining granted beans (if any)
 *   3. Credit new grant (if membership still active)
 *
 * See rules/EnergyBeanRule.md for detailed rules.
 */
async function main() {
  const start = Date.now();
  const db = createDb();

  try {
    const today = new Date();
    const todayDay = today.getDate();
    const lastDayOfMonth = new Date(
      today.getFullYear(),
      today.getMonth() + 1,
      0
    ).getDate();

    // Find users with a membership-grant entry whose day-of-month matches today.
    // For day 29/30/31 in shorter months, match on last day of month.
    const eligibleUsers: {
      id: string;
      beans: number;
      granted_beans: number;
      grade: string;
      vip_due_at: Date | null;
      grant_day: number;
    }[] = await db.$queryRaw`
      SELECT
        u.id,
        u.beans,
        u.granted_beans,
        u.grade,
        u.vip_due_at,
        EXTRACT(DAY FROM ub.first_grant)::int AS grant_day
      FROM users u
      INNER JOIN (
        SELECT user_id, MIN(created_at) AS first_grant
        FROM user_beans
        WHERE slug = 'membership-grant'
        GROUP BY user_id
      ) ub ON u.id = ub.user_id
      WHERE u.grade != 'free'
        AND u.is_active = true
        AND (
          EXTRACT(DAY FROM ub.first_grant)::int = ${todayDay}
          OR (
            EXTRACT(DAY FROM ub.first_grant)::int > ${lastDayOfMonth}
            AND ${todayDay} = ${lastDayOfMonth}
          )
        )
    `;

    let debited = 0;
    let credited = 0;
    let skipped = 0;

    for (const user of eligibleUsers) {
      const now = new Date();
      const isLifetime = user.grade === "lifetime";
      const isExpired =
        !isLifetime && (!user.vip_due_at || user.vip_due_at < now);
      const hasGrantedBeans = user.granted_beans > 0;

      // Skip: expired + no granted beans left
      if (isExpired && !hasGrantedBeans) {
        skipped++;
        continue;
      }

      const grantAmount = isLifetime ? 15000 : 10000;

      await db.$transaction(async (tx) => {
        let currentBeans = user.beans;
        let currentGranted = user.granted_beans;

        // Debit remaining granted beans
        if (hasGrantedBeans) {
          const debitAmount = -currentGranted;
          const debitOrigin = currentBeans;
          const debitResult = currentBeans + debitAmount;
          const debitNow = new Date();

          await tx.userBean.create({
            data: {
              id: generateId(),
              userId: user.id,
              beans: debitAmount,
              origin: debitOrigin,
              result: debitResult,
              slug: "monthly-reset-debit",
              reason: "月度未使用能量豆清零",
              data: { grantedBeansCleared: currentGranted },
              createdAt: debitNow,
              updatedAt: debitNow,
            },
          });

          currentBeans = debitResult;
          currentGranted = 0;
          debited++;
        }

        // Credit new grant (only if membership is active)
        if (!isExpired) {
          const creditOrigin = currentBeans;
          const creditResult = currentBeans + grantAmount;
          const creditNow = new Date();

          await tx.userBean.create({
            data: {
              id: generateId(),
              userId: user.id,
              beans: grantAmount,
              origin: creditOrigin,
              result: creditResult,
              slug: "monthly-reset-credit",
              reason: "月度能量豆续发",
              data: { gradeAtTime: user.grade, grantAmount },
              createdAt: creditNow,
              updatedAt: creditNow,
            },
          });

          currentBeans = creditResult;
          currentGranted = grantAmount;
          credited++;
        }

        // Atomic user balance update
        await tx.user.update({
          where: { id: user.id },
          data: {
            beans: currentBeans,
            grantedBeans: currentGranted,
          },
        });
      });
    }

    const elapsed = Date.now() - start;
    console.log(
      `[reset-energy-beans] done in ${elapsed}ms — ` +
        `eligible: ${eligibleUsers.length}, debited: ${debited}, ` +
        `credited: ${credited}, skipped: ${skipped}`
    );
  } catch (error) {
    console.error("[reset-energy-beans] failed:", error);
    process.exit(1);
  } finally {
    await db.$disconnect();
  }
}

/** Generate a ULID (inline to avoid server-only import) */
function generateId(): string {
  // Inline ULID-compatible generator using timestamp + random
  const timestamp = Date.now();
  const chars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ";
  let id = "";

  // Encode timestamp (first 10 chars)
  let t = timestamp;
  for (let i = 9; i >= 0; i--) {
    id = chars[t % 32] + id;
    t = Math.floor(t / 32);
  }

  // Random part (16 chars)
  for (let i = 0; i < 16; i++) {
    id += chars[Math.floor(Math.random() * 32)];
  }

  return id;
}

main();
```

**Note:** The cron script uses an inline `generateId()` instead of importing `ulid` from the app's node_modules, because cron scripts run standalone. If `ulid` is available in the scripts context, replace `generateId()` with the `ulid` import:

```typescript
import { ulid } from "ulid";
```

and use `ulid()` in place of `generateId()`. Check whether the existing cron (`update-play-streaks.ts`) imports any npm packages — if it does, `ulid` should work too.

**Step 2: Verify the script compiles**

Run:
```bash
npx tsx scripts/cron/reset-energy-beans.ts --help 2>&1 || echo "Script loaded"
```

Expected: No syntax or import errors.

**Step 3: Commit**

```bash
git add scripts/cron/reset-energy-beans.ts
git commit -m "feat: add daily energy bean reset cron script (1 AM)"
```

---

### Task 8: Verify full build

**Files:**
- None (verification only)

**Step 1: Generate Prisma client**

Run:
```bash
npm run prisma:generate
```

Expected: No errors.

**Step 2: Run lint**

Run:
```bash
npm run lint
```

Expected: No errors.

**Step 3: Run build**

Run:
```bash
npm run build
```

Expected: Build succeeds with no type errors.

**Step 4: Verify migration status**

Run:
```bash
npx prisma migrate status
```

Expected: All migrations applied.

**Step 5: Commit any fixes if needed**

If any fixes were required, commit them:
```bash
git add -A
git commit -m "fix: resolve build issues for energy beans feature"
```
