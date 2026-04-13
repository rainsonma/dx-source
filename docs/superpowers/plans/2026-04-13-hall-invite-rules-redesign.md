# Hall Invite Rules Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the hardcoded `活动规则` section on `/hall/invite` with a richer rules block — a 5-item rules list plus a two-tier commission comparison plus an invitee discount list — with the commission and discount sub-sections gated behind paid membership (free users see a single locked placeholder with an upgrade CTA to `/purchase/membership`).

**Architecture:** Extract rules/commission/discount data into a typed constants module (`helpers/invite-rules.ts`), render it through a dedicated client component (`components/invite-rules-section.tsx`) that handles grade-based gating with all sub-components colocated in one file, and wire the profile fetch through SWR in the existing page component. No backend changes.

**Tech Stack:** Next.js 16 (app router), React 19, TypeScript 5, TailwindCSS v4 (oklch theme, teal-600 accent), lucide-react icons, SWR v2 for client-side data fetching, shadcn/ui component library (read-only).

**Spec:** `docs/superpowers/specs/2026-04-13-hall-invite-rules-redesign-design.md`

---

## Environment Setup

**Working directory:** `/Users/rainsen/Programs/Projects/douxue/dx-source`

**Files already explored (context):**
- `dx-web/src/features/web/invite/components/invite-content.tsx` — current target, client component with inline `rules` array and amber info alert
- `dx-web/src/app/(web)/hall/(main)/invite/page.tsx` — page component using `useEffect` + `useState` for invite data
- `dx-web/src/consts/user-grade.ts` — defines `UserGrade = "free" | "month" | "season" | "year" | "lifetime"`
- `dx-web/src/features/web/me/types/me.types.ts` — exports `ApiProfileData` with `grade: string` field
- `dx-web/src/lib/swr.tsx` — global SWR provider with `swrFetcher` configured; `useSWR("/api/user/profile")` works without local fetcher
- `dx-web/src/app/(web)/purchase/membership/page.tsx` — destination for the "升级会员" CTA

**Verification commands (no test framework exists):**
- `cd dx-web && npx tsc --noEmit` — fast TypeScript type check (no emit)
- `cd dx-web && npm run lint` — ESLint check
- `cd dx-web && npm run build` — full Next.js build (includes type check)
- `cd dx-web && npm run dev` — dev server on `http://localhost:3000`

**Project conventions to follow:**
- Absolute imports via `@/*` alias
- No `any`, no `!` non-null assertions, no `@ts-ignore`
- No `console.log` in production code
- Many small files over few large files (per CLAUDE.md)
- Never edit files in `dx-web/src/components/ui/` (shadcn-managed)
- Commit style: `feat(web):`, `fix(web):`, etc.
- Always ask before `git commit` — the executing agent should confirm with the user before running the commit command in each task's final step

---

## Task 1: Create the invite rules data module

**Files:**
- Create: `dx-web/src/features/web/invite/helpers/invite-rules.ts`

**Purpose:** Pure TypeScript module exposing typed constants for rules, commission tiers, invitee discounts, plus a single `formatRewardValue` formatter. No React or Next.js imports. This module is consumed by the section component in Task 2.

- [ ] **Step 1: Write the data module**

Create `dx-web/src/features/web/invite/helpers/invite-rules.ts` with the following content:

```typescript
export const inviteRules: string[] = [
  "邀请好友通过您的专属链接、邀请码或二维码注册斗学账号",
  "好友成功注册并完成首次购买会员即算邀请成功",
  "佣金数额实时反馈，邀请所得收入清晰体现，一目了然",
  "邀请人数不设上限，邀请越多佣金越多",
  "佣金收入按时结算，可随时申请提现",
];

export type RewardValue =
  | { kind: "fixed"; amount: number }
  | { kind: "percent"; value: number };

export const commissionRewardKeys = [
  "lifetime",
  "year",
  "season",
  "month",
  "renewal",
] as const;

export type CommissionRewardKey = (typeof commissionRewardKeys)[number];

export const rewardRowLabels: Record<CommissionRewardKey, string> = {
  lifetime: "邀请永久会员",
  year: "邀请年度会员",
  season: "邀请季度会员",
  month: "邀请月度会员",
  renewal: "持续续费返佣",
};

export type InviterTierId = "standard" | "lifetime";

export type CommissionTier = {
  id: InviterTierId;
  label: string;
  sublabel: string;
  rewards: Record<CommissionRewardKey, RewardValue>;
};

export const commissionTiers: CommissionTier[] = [
  {
    id: "standard",
    label: "普通付费会员",
    sublabel: "月度 / 季度 / 年度",
    rewards: {
      lifetime: { kind: "fixed", amount: 500 },
      year: { kind: "percent", value: 30 },
      season: { kind: "percent", value: 30 },
      month: { kind: "percent", value: 30 },
      renewal: { kind: "percent", value: 10 },
    },
  },
  {
    id: "lifetime",
    label: "终身会员",
    sublabel: "永久会员",
    rewards: {
      lifetime: { kind: "fixed", amount: 600 },
      year: { kind: "percent", value: 50 },
      season: { kind: "percent", value: 50 },
      month: { kind: "percent", value: 50 },
      renewal: { kind: "percent", value: 20 },
    },
  },
];

export type InviteeDiscountGrade = "lifetime" | "year" | "season" | "month";

export type InviteeDiscount = {
  grade: InviteeDiscountGrade;
  label: string;
  value: RewardValue;
};

export const inviteeDiscounts: InviteeDiscount[] = [
  { grade: "lifetime", label: "购买永久会员", value: { kind: "fixed", amount: 99 } },
  { grade: "year", label: "购买年度会员", value: { kind: "percent", value: 10 } },
  { grade: "season", label: "购买季度会员", value: { kind: "percent", value: 10 } },
  { grade: "month", label: "购买月度会员", value: { kind: "percent", value: 10 } },
];

export function formatRewardValue(value: RewardValue): string {
  switch (value.kind) {
    case "fixed":
      return `¥${value.amount}`;
    case "percent":
      return `${value.value}%`;
  }
}
```

- [ ] **Step 2: Run TypeScript type check**

```bash
cd dx-web && npx tsc --noEmit
```

Expected: exits with code 0 and no output. If errors appear, they will point at the new file — fix syntax or types until the check passes.

- [ ] **Step 3: Run lint**

```bash
cd dx-web && npm run lint
```

Expected: exits with code 0. No errors or warnings in the new file.

- [ ] **Step 4: Ask user, then commit**

Ask the user: "Task 1 complete — data module written and checks pass. Commit now?"

On approval, run:

```bash
git add dx-web/src/features/web/invite/helpers/invite-rules.ts
git commit -m "feat(web): add invite rules data module with commission tiers and discounts"
```

---

## Task 2: Create the invite rules section component

**Files:**
- Create: `dx-web/src/features/web/invite/components/invite-rules-section.tsx`

**Purpose:** Client component that renders the full rules section with grade-based gating. All sub-components (`SectionHeader`, `SubHeader`, `RulesList`, `CommissionTierCard`, `CommissionTiersBlock`, `InviteeDiscountsBlock`, `LockedHint`, `LoadingPlaceholder`) are colocated in one file because they are tightly coupled to this layout and have zero reuse potential. The main exported component is `InviteRulesSection`; everything else is a file-local helper.

- [ ] **Step 1: Write the complete section component**

Create `dx-web/src/features/web/invite/components/invite-rules-section.tsx` with the following content:

```typescript
"use client";

import Link from "next/link";
import { ScrollText, Lock, Crown } from "lucide-react";
import type { UserGrade } from "@/consts/user-grade";
import {
  inviteRules,
  commissionTiers,
  commissionRewardKeys,
  rewardRowLabels,
  inviteeDiscounts,
  formatRewardValue,
  type CommissionTier,
} from "@/features/web/invite/helpers/invite-rules";

type Props = {
  userGrade: UserGrade | null;
};

export function InviteRulesSection({ userGrade }: Props) {
  const isLoading = userGrade === null;
  const isFreeUser = userGrade === "free";

  return (
    <div className="flex flex-col gap-5 rounded-[14px] border border-border bg-card p-4 lg:gap-6 lg:p-6">
      <SectionHeader />
      <RulesList />
      {isLoading && <LoadingPlaceholder />}
      {!isLoading && isFreeUser && <LockedHint />}
      {!isLoading && !isFreeUser && (
        <>
          <CommissionTiersBlock />
          <InviteeDiscountsBlock />
        </>
      )}
    </div>
  );
}

function SectionHeader() {
  return (
    <div className="flex items-center gap-2">
      <ScrollText className="h-[18px] w-[18px] text-teal-600" />
      <span className="text-base font-semibold text-foreground">活动规则</span>
    </div>
  );
}

function SubHeader({ title }: { title: string }) {
  return (
    <div className="flex items-center gap-2">
      <span className="h-3 w-1 rounded-full bg-teal-600" />
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
    </div>
  );
}

function RulesList() {
  return (
    <div className="flex flex-col gap-3">
      {inviteRules.map((rule, i) => (
        <div key={i} className="flex gap-2.5">
          <span className="text-sm font-semibold text-teal-600">{i + 1}.</span>
          <span className="text-sm text-muted-foreground">{rule}</span>
        </div>
      ))}
    </div>
  );
}

function CommissionTiersBlock() {
  return (
    <div className="flex flex-col gap-3">
      <SubHeader title="佣金体系" />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {commissionTiers.map((tier) => (
          <CommissionTierCard key={tier.id} tier={tier} />
        ))}
      </div>
    </div>
  );
}

function CommissionTierCard({ tier }: { tier: CommissionTier }) {
  const isLifetime = tier.id === "lifetime";
  const cardClass = isLifetime
    ? "rounded-[10px] border border-teal-500/40 bg-gradient-to-b from-teal-50/40 to-transparent p-4"
    : "rounded-[10px] border border-border bg-card p-4";

  return (
    <div className={`flex flex-col ${cardClass}`}>
      <div className="flex items-center gap-1.5">
        <span className="text-sm font-semibold text-foreground">{tier.label}</span>
        {isLifetime && <Crown className="h-3.5 w-3.5 text-teal-600" />}
      </div>
      <span className="pb-2 pt-0.5 text-xs text-muted-foreground">{tier.sublabel}</span>
      <div className="flex flex-col">
        {commissionRewardKeys.map((key) => (
          <div
            key={key}
            className="flex items-center justify-between border-b border-border/40 py-2 last:border-0"
          >
            <span className="text-sm text-muted-foreground">{rewardRowLabels[key]}</span>
            <span className="text-sm font-semibold text-foreground">
              {formatRewardValue(tier.rewards[key])}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function InviteeDiscountsBlock() {
  return (
    <div className="flex flex-col gap-3">
      <SubHeader title="被邀请者专属折扣" />
      <div className="flex flex-col">
        {inviteeDiscounts.map((discount) => (
          <div
            key={discount.grade}
            className="flex items-center justify-between border-b border-border/40 py-2 last:border-0"
          >
            <span className="text-sm text-muted-foreground">{discount.label}</span>
            <span className="text-sm font-semibold text-foreground">
              {formatRewardValue(discount.value)}
            </span>
          </div>
        ))}
      </div>
      <p className="pt-1 text-xs text-muted-foreground">
        * 未获邀请直接购买者无折扣优惠
      </p>
    </div>
  );
}

function LockedHint() {
  return (
    <div className="flex min-h-[280px] flex-col items-center justify-center gap-3 rounded-[10px] border border-dashed border-border/60 bg-muted/30 p-6">
      <Lock className="h-6 w-6 text-muted-foreground" />
      <div className="flex flex-col items-center gap-1 text-center">
        <span className="text-sm font-medium text-foreground">
          佣金体系与会员折扣仅会员可见
        </span>
        <span className="text-xs text-muted-foreground">
          升级会员解锁完整邀请奖励
        </span>
      </div>
      <Link
        href="/purchase/membership"
        className="inline-flex items-center gap-1.5 rounded-[10px] bg-teal-600 px-4 py-2 text-xs font-semibold text-white transition-colors hover:bg-teal-700"
      >
        升级会员
      </Link>
    </div>
  );
}

function LoadingPlaceholder() {
  return (
    <div className="min-h-[280px] animate-pulse rounded-[10px] border border-border/40 bg-muted/10" />
  );
}
```

- [ ] **Step 2: Run TypeScript type check**

```bash
cd dx-web && npx tsc --noEmit
```

Expected: exits with code 0 and no output. Common issues to watch for:
- `Cannot find module '@/features/web/invite/helpers/invite-rules'` → Task 1 wasn't completed or the path is wrong.
- `Cannot find name 'X'` for a sub-component → a sub-component definition is missing from the file.
- `'X' is declared but its value is never read` → an unused import; remove it.

- [ ] **Step 3: Run lint**

```bash
cd dx-web && npm run lint
```

Expected: exits with code 0. If ESLint complains about `import/order`, reorder imports to match the rule; if about `key` in `.map`, verify every mapped element has a `key` prop.

- [ ] **Step 4: Ask user, then commit**

Ask the user: "Task 2 complete — section component written and checks pass. Commit now?"

On approval, run:

```bash
git add dx-web/src/features/web/invite/components/invite-rules-section.tsx
git commit -m "feat(web): add invite rules section component with grade-based gating"
```

---

## Task 3: Wire up InviteContent and the InvitePage

**Files:**
- Modify: `dx-web/src/features/web/invite/components/invite-content.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/invite/page.tsx`

**Purpose:** Replace the old inline rules card in `InviteContent` with `<InviteRulesSection userGrade={userGrade} />`, add `userGrade` as a required prop, remove unused imports and the old `rules` array. In `InvitePage`, add a parallel SWR fetch for `/api/user/profile`, compute `userGrade` with error-fallback-to-`"free"`, and pass it to `InviteContent`. Both files change together because `userGrade` is a required prop.

- [ ] **Step 1: Modify invite-content.tsx — update imports**

In `dx-web/src/features/web/invite/components/invite-content.tsx`, replace the lucide-react import block (currently lines 4-14) from:

```typescript
import {
  Copy,
  Check,
  Users,
  DollarSign,
  UserCheck,
  TrendingUp,
  ScrollText,
  Info,
  Share2,
} from "lucide-react";
```

to:

```typescript
import {
  Copy,
  Check,
  Users,
  DollarSign,
  UserCheck,
  TrendingUp,
  Share2,
} from "lucide-react";
```

(Removed: `ScrollText`, `Info`.)

Then add these two imports below the existing local imports (below the `ShareSnippetsModal` import line):

```typescript
import { InviteRulesSection } from "@/features/web/invite/components/invite-rules-section";
import type { UserGrade } from "@/consts/user-grade";
```

- [ ] **Step 2: Modify invite-content.tsx — remove old rules array**

Delete the entire `rules` const block (currently around lines 29-35):

```typescript
const rules = [
  "邀请好友通过您的专属链接注册斗学平台",
  "好友成功注册并完成首次购买会员即算邀请成功",
  "月度会员佣金 ¥9.90，季度会员 ¥29.70，年度会员 ¥89.70",
  "佣金每月 15 日统一结算，可提现至绑定的银行账户",
  "邀请人数不设上限，邀请越多佣金越多",
];
```

Also remove any blank line left behind to keep the file clean.

- [ ] **Step 3: Modify invite-content.tsx — update props type and function signature**

Change the `InviteContentProps` type from:

```typescript
type InviteContentProps = {
  inviteUrl: string;
  referrals: ReferralItem[];
  totalPages: number;
  stats: InviteStats;
};
```

to:

```typescript
type InviteContentProps = {
  inviteUrl: string;
  referrals: ReferralItem[];
  totalPages: number;
  stats: InviteStats;
  userGrade: UserGrade | null;
};
```

Then change the function signature from:

```typescript
export function InviteContent({ inviteUrl, referrals, totalPages, stats }: InviteContentProps) {
```

to:

```typescript
export function InviteContent({
  inviteUrl,
  referrals,
  totalPages,
  stats,
  userGrade,
}: InviteContentProps) {
```

- [ ] **Step 4: Modify invite-content.tsx — replace the old rules card JSX**

Find the rules card block (currently around lines 160-180) which looks like this:

```tsx
{/* Rules card */}
<div className="flex flex-col gap-4 rounded-[14px] border border-border bg-card p-4 lg:p-6">
  <div className="flex items-center gap-2">
    <ScrollText className="h-[18px] w-[18px] text-teal-600" />
    <span className="text-base font-semibold text-foreground">活动规则</span>
  </div>
  <div className="flex flex-col gap-3">
    {rules.map((rule, i) => (
      <div key={i} className="flex gap-2.5">
        <span className="text-sm font-semibold text-teal-600">{i + 1}.</span>
        <span className="text-sm text-muted-foreground">{rule}</span>
      </div>
    ))}
  </div>
  <div className="flex gap-2 rounded-[10px] border border-amber-500/20 bg-amber-50/10 p-3">
    <Info className="h-3.5 w-3.5 shrink-0 text-amber-500" />
    <span className="text-xs leading-relaxed text-amber-800">
      注意：佣金奖励仅限被邀请好友首次购买会员时发放，续费订单不参与佣金计算。
    </span>
  </div>
</div>
```

Replace the entire block (including the `{/* Rules card */}` comment) with:

```tsx
{/* Rules card */}
<InviteRulesSection userGrade={userGrade} />
```

At this point, `invite-content.tsx` is complete. A partial type check would now fail on `page.tsx` because it doesn't yet pass `userGrade` — that's expected, and we fix it next.

- [ ] **Step 5: Modify page.tsx — update imports**

In `dx-web/src/app/(web)/hall/(main)/invite/page.tsx`, replace the entire import block at the top:

Current:
```typescript
"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { InviteContent } from "@/features/web/invite/components/invite-content";
import type { InviteStats } from "@/features/web/invite/helpers/invite-stats.helper";
import type { ReferralItem } from "@/features/web/invite/actions/invite.action";
```

New:
```typescript
"use client";

import { useEffect, useState } from "react";
import useSWR from "swr";
import { apiClient } from "@/lib/api-client";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { InviteContent } from "@/features/web/invite/components/invite-content";
import type { InviteStats } from "@/features/web/invite/helpers/invite-stats.helper";
import type { ReferralItem } from "@/features/web/invite/actions/invite.action";
import type { ApiProfileData } from "@/features/web/me/types/me.types";
import type { UserGrade } from "@/consts/user-grade";
```

(Added: `useSWR` from `swr`, `ApiProfileData`, `UserGrade` types.)

- [ ] **Step 6: Modify page.tsx — add SWR fetch and compute userGrade**

Inside the `InvitePage` function body, directly after the four existing `useState` declarations and before the existing `useEffect`, add:

```typescript
  const { data: profileData, error: profileError } = useSWR<ApiProfileData>(
    "/api/user/profile"
  );

  const userGrade: UserGrade | null = profileError
    ? "free"
    : profileData
      ? (profileData.grade as UserGrade)
      : null;
```

The top of the function should now read:

```typescript
export default function InvitePage() {
  const [inviteUrl, setInviteUrl] = useState("");
  const [referrals, setReferrals] = useState<ReferralItem[]>([]);
  const [totalPages, setTotalPages] = useState(0);
  const [stats, setStats] = useState<InviteStats>(emptyStats);

  const { data: profileData, error: profileError } = useSWR<ApiProfileData>(
    "/api/user/profile"
  );

  const userGrade: UserGrade | null = profileError
    ? "free"
    : profileData
      ? (profileData.grade as UserGrade)
      : null;

  useEffect(() => {
    // ... existing body unchanged
```

- [ ] **Step 7: Modify page.tsx — pass userGrade to InviteContent**

Find the `<InviteContent>` JSX call in the return statement:

Current:
```tsx
<InviteContent
  inviteUrl={inviteUrl}
  referrals={referrals}
  totalPages={totalPages}
  stats={stats}
/>
```

New:
```tsx
<InviteContent
  inviteUrl={inviteUrl}
  referrals={referrals}
  totalPages={totalPages}
  stats={stats}
  userGrade={userGrade}
/>
```

- [ ] **Step 8: Run TypeScript type check**

```bash
cd dx-web && npx tsc --noEmit
```

Expected: exits with code 0 and no output. If errors persist:
- `Property 'userGrade' is missing` → Step 7 was skipped or the JSX wasn't updated.
- `Cannot find module '@/features/web/me/types/me.types'` → verify the me types file exists; if not, the spec's assumption is wrong and you need to re-read profile types.
- `'ScrollText' is declared but its value is never read` in `invite-content.tsx` → Step 1 import cleanup was incomplete.

- [ ] **Step 9: Run lint**

```bash
cd dx-web && npm run lint
```

Expected: exits with code 0. No errors or warnings.

- [ ] **Step 10: Ask user, then commit**

Ask the user: "Task 3 complete — InviteContent and InvitePage wired up, type check and lint pass. Commit now?"

On approval, run:

```bash
git add "dx-web/src/features/web/invite/components/invite-content.tsx" "dx-web/src/app/(web)/hall/(main)/invite/page.tsx"
git commit -m "feat(web): wire hall invite page to new rules section with grade gating"
```

(The path with route-group parentheses must be quoted in bash.)

---

## Task 4: Final verification

**Purpose:** Run the complete build, start the dev server, and verify the change works end-to-end in a browser. Check for regressions in the rest of the invite page. No code changes — this is pure verification.

- [ ] **Step 1: Run full Next.js build**

```bash
cd dx-web && npm run build
```

Expected: build completes with "Compiled successfully" and no TypeScript errors. This is more thorough than `tsc --noEmit` because it also checks Next.js-specific things (route types, server/client boundaries, etc.).

If the build fails, read the error carefully — it usually points at a specific file and line. Fix and re-run.

- [ ] **Step 2: Start the dev server in the background**

```bash
cd dx-web && npm run dev
```

Wait for the line `✓ Ready in Xs` indicating the server is running at `http://localhost:3000`.

If using the Bash tool with `run_in_background: true`, capture the PID and monitor stdout until the ready line appears.

- [ ] **Step 3: Navigate to the invite page**

Open `http://localhost:3000/hall/invite` in a browser.

If not logged in, the page will redirect to sign-in. Log in with whatever test account is available, then navigate back to `/hall/invite`.

- [ ] **Step 4: Check the browser console and Next.js terminal for errors**

- Open browser DevTools → Console tab
- Verify: no JavaScript errors, no React hydration errors, no warnings about missing `key` props
- Check the Next.js terminal output: no red error messages, no unhandled promise rejections

- [ ] **Step 5: Verify the new rules section appears above the share modal**

Scroll to the bottom of the page (below the 邀请记录 table). Confirm:

- [ ] Section container has the teal-accented `活动规则` header with a `ScrollText` icon (upper-left)
- [ ] 5 numbered rules render (1 through 5), text matches the spec exactly
- [ ] The old amber "佣金奖励仅限首次购买" info alert is GONE

- [ ] **Step 6: Verify grade-based rendering (test at least one state)**

Depending on the logged-in user's grade, verify one of these:

**If the logged-in user is FREE:**
- [ ] A single dashed-border card with a lock icon appears below the rules list
- [ ] Text: "佣金体系与会员折扣仅会员可见"
- [ ] Subtitle: "升级会员解锁完整邀请奖励"
- [ ] Button: "升级会员"
- [ ] Clicking "升级会员" navigates to `/purchase/membership`
- [ ] No commission tier cards visible
- [ ] No discount list visible

**If the logged-in user is PAID (month/season/year/lifetime):**
- [ ] `佣金体系` sub-header with teal accent bar appears below rules
- [ ] Two commission tier cards side-by-side (on desktop)
- [ ] Left card: "普通付费会员" with sublabel "月度 / 季度 / 年度"; 5 reward rows: ¥500, 30%, 30%, 30%, 10%
- [ ] Right card: "终身会员" with Crown icon next to the label, sublabel "永久会员"; 5 reward rows: ¥600, 50%, 50%, 50%, 20%
- [ ] Right card has visibly different border (teal-500/40) and a subtle gradient background (teal-50/40 at top fading down)
- [ ] Below commission tiers: `被邀请者专属折扣` sub-header
- [ ] 4 discount rows: 购买永久会员 ¥99, 购买年度会员 10%, 购买季度会员 10%, 购买月度会员 10%
- [ ] Footnote at the bottom: `* 未获邀请直接购买者无折扣优惠`

Document in the task completion comment which grade was actually tested and what you observed.

- [ ] **Step 7: Verify responsive layout**

Open browser DevTools → Toggle device toolbar (or resize the window) to a width below 1024px (e.g., iPhone 12 at 390px).

Verify:
- [ ] Commission tier cards stack vertically (if visible for the current user's grade)
- [ ] No horizontal scrolling on any part of the page
- [ ] Rules text wraps cleanly
- [ ] Button and CTA remain tappable (not clipped)

- [ ] **Step 8: Regression check — other sections untouched**

On the same page (back at desktop width), verify every other element still renders and functions:

- [ ] Top banner with teal gradient and "快速分享词" button (upper-right of banner)
- [ ] Invite URL input field showing the full URL
- [ ] "复制链接" button — click it and verify it shows a checkmark for ~2 seconds (clipboard copy feedback)
- [ ] Two QR code cards to the right of the banner on desktop
- [ ] Stats row: 4 cards in a row (累计获得推广佣金, 好友数量, 待验证, 转化率)
- [ ] Invite records table with header "邀请记录" and filter pills (全部 / 待激活 / 已激活)
- [ ] If there are enough referral records, pagination controls work
- [ ] Click "快速分享词" → modal opens with share snippets
- [ ] Close the modal with the X or by clicking outside

If any of these regressed, STOP and investigate. The most likely cause is an accidental edit to `invite-content.tsx` outside the rules card area.

- [ ] **Step 9: Stop the dev server**

Press Ctrl+C in the terminal running `npm run dev`, or if using the Bash tool with background mode, kill the background process.

- [ ] **Step 10: Confirm working tree is clean**

```bash
git status
```

Expected: `nothing to commit, working tree clean`. If anything is dirty, either commit remaining intentional changes (with user approval) or stash/discard accidental changes.

- [ ] **Step 11: Report completion to user**

Summarize for the user:
- What was tested (which grade, which viewport sizes, which regressions checked)
- What wasn't tested (other grades not available in your dev session)
- Any observations worth flagging (minor visual tweaks that could be improved, unexpected behavior, etc.)
- Confirm all automated checks passed: `tsc --noEmit`, `npm run lint`, `npm run build`

---

## Self-Review Checklist (executed by the plan author — already completed)

Before handing this plan off, the author verified:

**Spec coverage:**
- [x] Rules list (5 items, renumbered) — Task 1 (data), Task 2 (`RulesList`)
- [x] Commission system (two tiers, 5 reward rows each) — Task 1 (data), Task 2 (`CommissionTiersBlock`, `CommissionTierCard`)
- [x] Invitee discount list (4 rows + footnote) — Task 1 (data), Task 2 (`InviteeDiscountsBlock`)
- [x] Free-user gating with single locked hint — Task 2 (`LockedHint`)
- [x] Loading state — Task 2 (`LoadingPlaceholder`)
- [x] Error handling (SWR error → treat as free) — Task 3 Step 6
- [x] `userGrade` prop threading — Task 3 Steps 3, 7
- [x] Remove old `rules` array — Task 3 Step 2
- [x] Remove amber info alert — Task 3 Step 4
- [x] Remove unused imports — Task 3 Step 1
- [x] Upgrade CTA to `/purchase/membership` — Task 2 (`LockedHint`)
- [x] Lifetime tier elevated styling — Task 2 (`CommissionTierCard` conditional)
- [x] `min-h-[280px]` on both placeholder and hint — Task 2 (`LockedHint`, `LoadingPlaceholder`)
- [x] Lint + type check + build gates — Tasks 1, 2, 3 verification steps + Task 4

**Placeholder scan:**
- [x] No "TBD", "TODO", "implement later" in plan body
- [x] No "add error handling" or "handle edge cases" without specifics
- [x] Every code step shows complete code
- [x] No "similar to Task N" — code is repeated where needed

**Type consistency:**
- [x] `UserGrade` imported from `@/consts/user-grade` consistently
- [x] `ApiProfileData` imported from `@/features/web/me/types/me.types` consistently
- [x] `CommissionTier` type used as `CommissionTierCard` prop type
- [x] `commissionRewardKeys` array matches `rewardRowLabels` keys
- [x] `formatRewardValue` handles all `RewardValue.kind` variants
- [x] `userGrade` prop type (`UserGrade | null`) is consistent across `InviteContent`, `InviteRulesSection`, and the page component

**Scope:**
- [x] Single-subsystem plan (frontend-only, one page, one feature)
- [x] No unrelated refactors
- [x] No backend changes
- [x] No tests beyond what the project provides (no test framework exists)
