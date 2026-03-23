# Invite Code System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add permanent invite codes to users, track referral relationships, and render QR codes on the invite page using easyqrcodejs.

**Architecture:** `User.inviteCode` stores a unique 8-char alphanumeric code generated at sign-up. Visiting `/invite/{code}` sets a `ref` cookie and redirects to `/auth/signup`. Sign-up services read the cookie, create a `UserReferral` record, and clear it. The invite page renders a real QR code client-side with easyqrcodejs.

**Tech Stack:** Next.js 16 App Router, Prisma 7, easyqrcodejs, cookies from `next/headers`

---

### Task 1: Install easyqrcodejs

**Files:**
- Modify: `package.json`

**Step 1: Install the package**

Run: `npm install easyqrcodejs`

**Step 2: Verify installation**

Run: `grep easyqrcodejs package.json`
Expected: `"easyqrcodejs": "^x.x.x"` in dependencies

**Step 3: Commit**

```bash
git add package.json package-lock.json
git commit -m "chore: install easyqrcodejs for QR code rendering"
```

---

### Task 2: Update Prisma schemas

**Files:**
- Modify: `prisma/schema/user.prisma`
- Modify: `prisma/schema/user-referral.prisma`

**Step 1: Add `inviteCode` to User model**

In `prisma/schema/user.prisma`, add after the `exp` field (line 14):

```prisma
  inviteCode   String    @unique @map("invite_code") @db.VarChar(8)
```

**Step 2: Update UserReferral model**

Replace `prisma/schema/user-referral.prisma` with:

```prisma
model UserReferral {
  id           String    @id @db.Char(26)
  referrerId   String    @map("referrer_id") @db.Char(26)
  inviteeId    String?   @map("invitee_id") @db.Char(26)
  status       String    @default("pending") @db.VarChar(20) // pending, paid, rewarded
  rewardAmount Decimal   @default(0) @map("reward_amount") @db.Decimal(10, 2)
  rewardedAt   DateTime? @map("rewarded_at") @db.Timestamptz

  createdAt DateTime @default(now()) @map("created_at") @db.Timestamptz
  updatedAt DateTime @updatedAt @map("updated_at") @db.Timestamptz

  referrer User  @relation("Referrals", fields: [referrerId], references: [id], onDelete: Restrict, onUpdate: Restrict)
  invitee  User? @relation("ReferredBy", fields: [inviteeId], references: [id], onDelete: Restrict, onUpdate: Restrict)

  @@index([referrerId])
  @@index([inviteeId])
  @@index([status])
  @@index([createdAt])
  @@map("user_referrals")
}
```

**Step 3: Generate migration**

Run: `npx prisma migrate dev --schema prisma/schema --name add-invite-code`

**Step 4: Generate Prisma client**

Run: `npm run prisma:generate`

**Step 5: Commit**

```bash
git add prisma/ src/generated/
git commit -m "feat: add User.inviteCode field, update UserReferral schema"
```

---

### Task 3: Add referral status constants

**Files:**
- Create: `src/consts/referral-status.ts`

**Step 1: Create the constants file**

Follow the pattern from `src/consts/game-status.ts`:

```typescript
export const REFERRAL_STATUSES = {
  PENDING: "pending",
  PAID: "paid",
  REWARDED: "rewarded",
} as const;

export type ReferralStatus =
  (typeof REFERRAL_STATUSES)[keyof typeof REFERRAL_STATUSES];

export const REFERRAL_STATUS_LABELS: Record<ReferralStatus, string> = {
  pending: "待验证",
  paid: "已付费",
  rewarded: "已发放",
};
```

**Step 2: Commit**

```bash
git add src/consts/referral-status.ts
git commit -m "feat: add referral status constants"
```

---

### Task 4: Add invite code generator helper

**Files:**
- Create: `src/features/web/invite/helpers/invite-code.helper.ts`

**Step 1: Create the helper**

```typescript
const CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
const CODE_LENGTH = 8;

/** Generate an 8-character alphanumeric invite code */
export function generateInviteCode(): string {
  const bytes = new Uint8Array(CODE_LENGTH);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => CHARSET[b % CHARSET.length]).join("");
}
```

**Step 2: Commit**

```bash
git add src/features/web/invite/helpers/invite-code.helper.ts
git commit -m "feat: add invite code generator helper"
```

---

### Task 5: Add user-referral model operations

**Files:**
- Create: `src/models/user-referral/user-referral.mutation.ts`
- Create: `src/models/user-referral/user-referral.query.ts`

**Step 1: Create the mutation file**

```typescript
import "server-only";

import { ulid } from "ulid";
import { db } from "@/lib/db";

type CreateReferralData = {
  referrerId: string;
  inviteeId: string;
};

/** Create a referral record linking referrer to invitee */
export async function createUserReferral(data: CreateReferralData) {
  return db.userReferral.create({
    data: {
      id: ulid(),
      referrerId: data.referrerId,
      inviteeId: data.inviteeId,
      status: "pending",
    },
    select: {
      id: true,
      referrerId: true,
      inviteeId: true,
      status: true,
    },
  });
}
```

**Step 2: Create the query file**

```typescript
import "server-only";

import { db } from "@/lib/db";

/** Find all referrals made by a specific user */
export async function findReferralsByReferrerId(referrerId: string) {
  return db.userReferral.findMany({
    where: { referrerId },
    select: {
      id: true,
      status: true,
      rewardAmount: true,
      rewardedAt: true,
      createdAt: true,
      invitee: {
        select: {
          id: true,
          username: true,
          nickname: true,
          email: true,
          grade: true,
        },
      },
    },
    orderBy: { createdAt: "desc" },
  });
}
```

**Step 3: Commit**

```bash
git add src/models/user-referral/
git commit -m "feat: add user-referral model query and mutation"
```

---

### Task 6: Update user model operations

**Files:**
- Modify: `src/models/user/user.query.ts` (add `findUserByInviteCode`, update `getUserProfile`)
- Modify: `src/models/user/user.mutation.ts` (add `inviteCode` to `createUser`)

**Step 1: Add `findUserByInviteCode` to `src/models/user/user.query.ts`**

Add after the `findUserByAccount` function:

```typescript
/** Find a user by their invite code */
export async function findUserByInviteCode(inviteCode: string) {
  return db.user.findUnique({
    where: { inviteCode },
    select: { id: true, username: true },
  });
}
```

**Step 2: Add `inviteCode` to `getUserProfile` select**

In `getUserProfile`, add `inviteCode: true` to the select object (after `exp: true`).

Also add `inviteCode` to the return object:

```typescript
return {
  id: user.id,
  username: user.username,
  nickname: user.nickname,
  email: user.email,
  grade: user.grade as UserGrade,
  exp: user.exp,
  inviteCode: user.inviteCode,
  avatarUrl: user.avatar?.url ?? null,
};
```

**Step 3: Update `createUser` in `src/models/user/user.mutation.ts`**

Add `inviteCode` to the `CreateUserData` type:

```typescript
type CreateUserData = {
  email: string;
  username: string;
  password: string;
  inviteCode: string;
};
```

Add `inviteCode: data.inviteCode` to the `data` object inside `db.user.create`.

Add `inviteCode: true` to the `select` object.

**Step 4: Commit**

```bash
git add src/models/user/
git commit -m "feat: add inviteCode to user model operations"
```

---

### Task 7: Update sign-up service with referral logic

**Files:**
- Modify: `src/features/web/auth/services/signup.service.ts`

**Step 1: Add imports**

Add at the top:

```typescript
import { cookies } from "next/headers";
import { generateInviteCode } from "@/features/web/invite/helpers/invite-code.helper";
import { findUserByInviteCode } from "@/models/user/user.query";
import { createUserReferral } from "@/models/user-referral/user-referral.mutation";
```

**Step 2: Update the `signUp` function**

After the user is created (line 87 `const user = await createUser(...)`) and before `redis.del`, add referral handling:

```typescript
  const inviteCode = generateInviteCode();

  const user = await createUser({
    email: data.email,
    username,
    password: hashedPassword,
    inviteCode,
  });

  // Handle referral tracking
  const cookieStore = await cookies();
  const refCode = cookieStore.get("ref")?.value;
  if (refCode) {
    const referrer = await findUserByInviteCode(refCode);
    if (referrer) {
      await createUserReferral({
        referrerId: referrer.id,
        inviteeId: user.id,
      });
    }
    cookieStore.delete("ref");
  }

  await redis.del(codeKey(data.email));
```

**Step 3: Commit**

```bash
git add src/features/web/auth/services/signup.service.ts
git commit -m "feat: generate invite code and track referral on sign-up"
```

---

### Task 8: Update sign-in service with referral logic (auto-register path)

**Files:**
- Modify: `src/features/web/auth/services/signin.service.ts`

**Step 1: Add imports**

Add at the top (same as signup):

```typescript
import { cookies } from "next/headers";
import { generateInviteCode } from "@/features/web/invite/helpers/invite-code.helper";
import { findUserByInviteCode } from "@/models/user/user.query";
import { createUserReferral } from "@/models/user-referral/user-referral.mutation";
```

**Step 2: Update the `emailSignIn` auto-register block**

In `emailSignIn`, the auto-register section (lines 67-83) creates a new user. Update it:

```typescript
  // Auto-register: derive username from email prefix
  let username = email.split("@")[0];
  const taken = await findUserByUsername(username);
  if (taken) {
    username = `${username}_${Date.now().toString(36)}`;
  }

  const rawPassword = crypto.randomUUID().slice(0, 16);
  const hashedPassword = await bcrypt.hash(rawPassword, 12);
  const inviteCode = generateInviteCode();

  const user = await createUser({
    email,
    username,
    password: hashedPassword,
    inviteCode,
  });

  // Handle referral tracking
  const cookieStore = await cookies();
  const refCode = cookieStore.get("ref")?.value;
  if (refCode) {
    const referrer = await findUserByInviteCode(refCode);
    if (referrer) {
      await createUserReferral({
        referrerId: referrer.id,
        inviteeId: user.id,
      });
    }
    cookieStore.delete("ref");
  }

  return { success: true, userId: user.id, username: user.username };
```

**Step 3: Commit**

```bash
git add src/features/web/auth/services/signin.service.ts
git commit -m "feat: generate invite code and track referral on email sign-in auto-register"
```

---

### Task 9: Create the invite redirect page

**Files:**
- Create: `src/app/(web)/invite/[code]/page.tsx`

**Step 1: Create the page**

```typescript
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { findUserByInviteCode } from "@/models/user/user.query";

type Props = {
  params: Promise<{ code: string }>;
};

/** Validate invite code, set ref cookie, redirect to sign-up */
export default async function InviteRedirectPage({ params }: Props) {
  const { code } = await params;

  const referrer = await findUserByInviteCode(code);
  if (referrer) {
    const cookieStore = await cookies();
    cookieStore.set("ref", code, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      maxAge: 60 * 60 * 24 * 7, // 7 days
      path: "/",
    });
  }

  redirect("/auth/signup");
}
```

**Step 2: Commit**

```bash
git add src/app/\(web\)/invite/
git commit -m "feat: add invite redirect page with ref cookie"
```

---

### Task 10: Create QR code hook

**Files:**
- Create: `src/features/web/invite/hooks/use-invite-qrcode.ts`

**Step 1: Create the hook**

```typescript
"use client";

import { useEffect, useRef } from "react";

/** Render a QR code into a container element using easyqrcodejs */
export function useInviteQrcode(url: string) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current || !url) return;

    let qrcode: unknown = null;

    const renderQrcode = async () => {
      const QRCode = (await import("easyqrcodejs")).default;

      // Clear any previous QR code
      if (containerRef.current) {
        containerRef.current.innerHTML = "";
      }

      qrcode = new QRCode(containerRef.current, {
        text: url,
        width: 100,
        height: 100,
        colorDark: "#0d9488",
        colorLight: "#ffffff",
        correctLevel: QRCode.CorrectLevel.M,
        quietZone: 4,
        quietZoneColor: "#ffffff",
      });
    };

    renderQrcode();

    return () => {
      if (containerRef.current) {
        containerRef.current.innerHTML = "";
      }
    };
  }, [url]);

  return containerRef;
}
```

Note: `easyqrcodejs` is a browser-only lib, so we dynamically import it inside `useEffect`. The hook returns a ref to attach to a container div.

**Step 2: Commit**

```bash
git add src/features/web/invite/hooks/use-invite-qrcode.ts
git commit -m "feat: add useInviteQrcode hook with easyqrcodejs"
```

---

### Task 11: Wire real data to the invite page

**Files:**
- Create: `src/features/web/invite/services/invite.service.ts`
- Create: `src/features/web/invite/hooks/use-invite-data.ts`
- Create: `src/features/web/invite/components/invite-qr-card.tsx`
- Modify: `src/features/web/invite/components/invite-content.tsx`
- Modify: `src/app/(web)/hall/(main)/invite/page.tsx`

**Step 1: Create the invite service**

`src/features/web/invite/services/invite.service.ts`:

```typescript
import "server-only";

import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import { findReferralsByReferrerId } from "@/models/user-referral/user-referral.query";

/** Fetch invite data for the current user */
export async function fetchInviteData() {
  const profile = await fetchUserProfile();
  if (!profile) return null;

  const referrals = await findReferralsByReferrerId(profile.id);

  return {
    inviteCode: profile.inviteCode,
    referrals,
  };
}
```

**Step 2: Create the InviteQrCard client component**

`src/features/web/invite/components/invite-qr-card.tsx`:

```typescript
"use client";

import { useInviteQrcode } from "@/features/web/invite/hooks/use-invite-qrcode";

type Props = {
  url: string;
  title: string;
  subtitle: string;
};

/** QR code card that renders a real QR code for the invite URL */
export function InviteQrCard({ url, title, subtitle }: Props) {
  const qrRef = useInviteQrcode(url);

  return (
    <div className="flex w-full flex-col gap-3.5 rounded-[14px] border border-slate-200 bg-white p-5 sm:w-[260px]">
      <div className="flex items-center gap-4">
        <div
          ref={qrRef}
          className="flex h-[100px] w-[100px] items-center justify-center rounded-[10px] border border-slate-200 bg-slate-50 p-2"
        />
        <div className="flex flex-col gap-2">
          <span className="text-sm font-semibold text-slate-900">{title}</span>
          <span className="text-xs text-slate-400">{subtitle}</span>
          <button
            type="button"
            className="rounded-lg bg-teal-600 px-3 py-1.5 text-xs font-medium text-white"
          >
            保存图片
          </button>
        </div>
      </div>
    </div>
  );
}
```

**Step 3: Update the invite page to pass data**

Modify `src/app/(web)/hall/(main)/invite/page.tsx` to fetch real data:

```typescript
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { InviteContent } from "@/features/web/invite/components/invite-content";
import { fetchInviteData } from "@/features/web/invite/services/invite.service";

export default async function InvitePage() {
  const inviteData = await fetchInviteData();

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="邀请推广"
        subtitle="邀请好友加入斗学，成功即可获得佣金奖励"
      />
      <InviteContent
        inviteCode={inviteData?.inviteCode ?? ""}
        referrals={inviteData?.referrals ?? []}
      />
    </div>
  );
}
```

**Step 4: Update `InviteContent` to accept props and use real QR code**

Convert `InviteContent` to a client component that:
- Accepts `inviteCode` and `referrals` as props
- Builds the invite URL from `inviteCode`
- Replaces the `QrCard` placeholder with `InviteQrCard`
- Replaces hardcoded invite link with the real URL
- Adds a working copy-to-clipboard button
- Replaces mock `invitedFriends` data with real `referrals` (keeping the same table layout)
- Keeps the stats and rules sections as-is (they'll be wired to real aggregation later)

Full details of the component updates:
- Import `InviteQrCard` from `./invite-qr-card`
- Remove `QrCard` local component and `QrCode` icon import
- Add props type with `inviteCode: string` and `referrals` array
- Build `inviteUrl` as `` `${window.location.origin}/invite/${inviteCode}` `` (or use an env var)
- Replace hardcoded `https://douxue.com/invite/abc123` with `inviteUrl`
- Add `navigator.clipboard.writeText(inviteUrl)` to the copy button
- Replace `<QrCard>` usages with `<InviteQrCard url={inviteUrl} ... />`
- Map `referrals` to the table rows (map `status` through `REFERRAL_STATUS_LABELS`)

**Step 5: Commit**

```bash
git add src/features/web/invite/ src/app/\(web\)/hall/\(main\)/invite/
git commit -m "feat: wire real invite data and QR code to invite page"
```

---

### Task 12: Verify the full flow

**Step 1: Generate Prisma client and run dev server**

Run: `npm run prisma:generate && npm run dev`

**Step 2: Verify invite code generation**

- Sign up a new user via `/auth/signup`
- Check the database: the new user should have an `inviteCode` value

**Step 3: Verify invite redirect**

- Visit `/invite/{inviteCode}` using the code from step 2
- Should redirect to `/auth/signup`
- Check browser cookies: should see `ref` cookie with the invite code

**Step 4: Verify referral tracking**

- Sign up another user while the `ref` cookie is set
- Check `user_referrals` table: should have a record with `referrer_id` = first user, `invitee_id` = second user, `status` = "pending"

**Step 5: Verify QR code on invite page**

- Log in and visit `/hall/invite`
- Should see a real QR code rendered (not a placeholder icon)
- Scan the QR code: should contain the invite URL

**Step 6: Run lint**

Run: `npm run lint`
Fix any issues.

**Step 7: Commit any fixes**

```bash
git add -A
git commit -m "fix: address lint issues"
```
