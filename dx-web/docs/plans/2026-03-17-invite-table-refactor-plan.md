# Invite Table Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor the 邀请记录 table to use server-side pagination (15 items/page) with digital pagination controls at bottom right.

**Architecture:** Server action with client-side page state. Initial SSR loads page 1 + stats. Page changes call a server action that returns the requested page. A reusable `DataTablePagination` component handles the pagination UI.

**Tech Stack:** Next.js 16 server actions, Prisma offset/limit, shadcn/ui Pagination primitives, React state

---

### Task 1: Add Pagination to Model Query

**Files:**
- Modify: `src/models/user-referral/user-referral.query.ts`

**Step 1: Add count query and pagination params**

Add a constant and a count function, then modify the existing query:

```typescript
/** Default number of referrals per page */
export const REFERRALS_PAGE_SIZE = 15;

/** Count total referrals made by a specific user */
export async function countReferralsByReferrerId(referrerId: string) {
  return db.userReferral.count({
    where: { referrerId },
  });
}
```

Modify `findReferralsByReferrerId` to accept optional pagination:

```typescript
/** Find referrals made by a specific user, with optional pagination */
export async function findReferralsByReferrerId(
  referrerId: string,
  page = 1,
  pageSize = REFERRALS_PAGE_SIZE,
) {
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
    skip: (page - 1) * pageSize,
    take: pageSize,
  });
}
```

**Step 2: Verify**

Run: `npm run build` — should compile with no errors.

**Step 3: Commit**

```
feat: add pagination params to referral query
```

---

### Task 2: Update Invite Service

**Files:**
- Modify: `src/features/web/invite/services/invite.service.ts`

**Step 1: Return totalPages and pre-computed stats**

The service needs to:
1. Fetch all referrals (without pagination) just for computing stats
2. Fetch page 1 of referrals for display
3. Count total referrals for totalPages

```typescript
import "server-only";

import { headers } from "next/headers";
import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import {
  countReferralsByReferrerId,
  findReferralsByReferrerId,
  REFERRALS_PAGE_SIZE,
} from "@/models/user-referral/user-referral.query";
import { computeInviteStats } from "@/features/web/invite/helpers/invite-stats.helper";

/** Build the base URL from request headers */
async function getBaseUrl() {
  const h = await headers();
  const proto = h.get("x-forwarded-proto") ?? "http";
  const host = h.get("host") ?? "localhost:3000";
  return `${proto}://${host}`;
}

/** Fetch the current user's invite URL, first page of referrals, total pages, and stats */
export async function fetchInviteData() {
  const profile = await fetchUserProfile();

  if (!profile) {
    return {
      inviteUrl: "",
      referrals: [],
      totalPages: 0,
      stats: computeInviteStats([]),
    };
  }

  const baseUrl = await getBaseUrl();

  const [referrals, totalCount, allReferrals] = await Promise.all([
    findReferralsByReferrerId(profile.id, 1),
    countReferralsByReferrerId(profile.id),
    // Fetch all referrals (unpaginated) for stats computation
    findReferralsByReferrerId(profile.id, 1, Number.MAX_SAFE_INTEGER),
  ]);

  const totalPages = Math.ceil(totalCount / REFERRALS_PAGE_SIZE);

  return {
    inviteUrl: `${baseUrl}/invite/${profile.inviteCode}`,
    referrals,
    totalPages,
    stats: computeInviteStats(allReferrals),
  };
}
```

> **Note:** Using `Number.MAX_SAFE_INTEGER` as pageSize to fetch all referrals for stats is pragmatic. If referral counts grow large in the future, replace with a dedicated SQL aggregation query.

**Step 2: Verify**

Run: `npm run build` — may show type errors in downstream components (expected, fixed in later tasks).

**Step 3: Commit**

```
feat: return totalPages and stats from invite service
```

---

### Task 3: Create Server Action

**Files:**
- Create: `src/features/web/invite/actions/invite.action.ts`

**Step 1: Write the server action**

```typescript
"use server";

import { fetchUserProfile } from "@/features/web/auth/services/user.service";
import {
  countReferralsByReferrerId,
  findReferralsByReferrerId,
  REFERRALS_PAGE_SIZE,
} from "@/models/user-referral/user-referral.query";

/** Fetch a specific page of referral records for the current user */
export async function fetchReferralPage(page: number) {
  try {
    const profile = await fetchUserProfile();
    if (!profile) {
      return { error: "未登录" };
    }

    const safePage = Math.max(1, Math.floor(page));

    const [referrals, totalCount] = await Promise.all([
      findReferralsByReferrerId(profile.id, safePage),
      countReferralsByReferrerId(profile.id),
    ]);

    const totalPages = Math.ceil(totalCount / REFERRALS_PAGE_SIZE);

    return { data: { referrals, totalPages } };
  } catch {
    return { error: "获取邀请记录失败" };
  }
}
```

**Step 2: Verify**

Run: `npm run build` — action file should compile.

**Step 3: Commit**

```
feat: add fetchReferralPage server action
```

---

### Task 4: Create Reusable Pagination Component

**Files:**
- Create: `src/components/in/data-table-pagination.tsx`

**Step 1: Write the pagination component**

Uses shadcn/ui Pagination primitives. Layout: `Page X of Y  << < 1 2 3 ... N > >>`

```typescript
"use client";

import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationEllipsis,
} from "@/components/ui/pagination";
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from "lucide-react";

type DataTablePaginationProps = {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
};

/** Generate the visible page numbers with ellipsis gaps */
function getPageNumbers(current: number, total: number): (number | "ellipsis")[] {
  if (total <= 5) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }

  const pages: (number | "ellipsis")[] = [1];

  if (current > 3) {
    pages.push("ellipsis");
  }

  const start = Math.max(2, current - 1);
  const end = Math.min(total - 1, current + 1);

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  if (current < total - 2) {
    pages.push("ellipsis");
  }

  pages.push(total);

  return pages;
}

/** Reusable pagination controls for data tables */
export function DataTablePagination({
  currentPage,
  totalPages,
  onPageChange,
}: DataTablePaginationProps) {
  if (totalPages <= 1) return null;

  const isFirst = currentPage === 1;
  const isLast = currentPage === totalPages;
  const pages = getPageNumbers(currentPage, totalPages);

  return (
    <div className="flex items-center justify-end gap-4 px-4 py-3">
      <span className="text-sm text-slate-500">
        Page {currentPage} of {totalPages}
      </span>
      <Pagination className="mx-0 w-auto">
        <PaginationContent>
          {/* First page */}
          <PaginationItem>
            <button
              type="button"
              disabled={isFirst}
              onClick={() => onPageChange(1)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-100 disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronsLeft className="h-4 w-4" />
            </button>
          </PaginationItem>

          {/* Previous page */}
          <PaginationItem>
            <button
              type="button"
              disabled={isFirst}
              onClick={() => onPageChange(currentPage - 1)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-100 disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
          </PaginationItem>

          {/* Page numbers */}
          {pages.map((page, i) =>
            page === "ellipsis" ? (
              <PaginationItem key={`ellipsis-${i}`}>
                <PaginationEllipsis />
              </PaginationItem>
            ) : (
              <PaginationItem key={page}>
                <PaginationLink
                  href="#"
                  isActive={page === currentPage}
                  onClick={(e) => {
                    e.preventDefault();
                    onPageChange(page);
                  }}
                  className="h-8 w-8 text-sm"
                >
                  {page}
                </PaginationLink>
              </PaginationItem>
            ),
          )}

          {/* Next page */}
          <PaginationItem>
            <button
              type="button"
              disabled={isLast}
              onClick={() => onPageChange(currentPage + 1)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-100 disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </PaginationItem>

          {/* Last page */}
          <PaginationItem>
            <button
              type="button"
              disabled={isLast}
              onClick={() => onPageChange(totalPages)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-100 disabled:pointer-events-none disabled:opacity-50"
            >
              <ChevronsRight className="h-4 w-4" />
            </button>
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  );
}
```

**Step 2: Verify**

Run: `npm run build` — component should compile.

**Step 3: Commit**

```
feat: add reusable DataTablePagination component
```

---

### Task 5: Extract Table Helpers

**Files:**
- Create: `src/features/web/invite/helpers/referral-table.helper.ts`

**Step 1: Move helper functions from invite-content.tsx**

```typescript
import { REFERRAL_STATUSES } from "@/consts/referral-status";

/** Color palette for avatar initials, cycled by index */
export const AVATAR_COLORS = [
  { bg: "bg-blue-100", text: "text-blue-600" },
  { bg: "bg-purple-100", text: "text-purple-600" },
  { bg: "bg-teal-100", text: "text-teal-600" },
  { bg: "bg-amber-100", text: "text-amber-600" },
  { bg: "bg-red-100", text: "text-red-600" },
];

/** Get the display name for a referral's invitee */
export function getDisplayName(invitee: { nickname: string | null; username: string } | null): string {
  if (!invitee) return "-";
  return invitee.nickname || invitee.username;
}

/** Mask an email address for privacy display */
export function maskEmail(email: string | null): string {
  if (!email) return "-";
  const [local, domain] = email.split("@");
  if (!domain) return email;
  const visible = local.slice(0, 3);
  return `${visible}***@${domain}`;
}

/** Format a date to YYYY-MM-DD string */
export function formatDate(date: Date): string {
  return new Date(date).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}

/** Format reward amount for display */
export function formatReward(amount: unknown, status: string): string {
  if (status === REFERRAL_STATUSES.PENDING) return "-";
  const num = Number(amount);
  if (!num) return "-";
  return `¥ ${num.toFixed(2)}`;
}

/** Get CSS classes for referral status badge */
export function getStatusClasses(status: string): string {
  if (status === REFERRAL_STATUSES.PENDING) {
    return "bg-amber-100 text-amber-700";
  }
  return "bg-teal-600/10 text-teal-600";
}
```

**Step 2: Commit**

```
refactor: extract referral table helpers
```

---

### Task 6: Create InviteReferralTable Component

**Files:**
- Create: `src/features/web/invite/components/invite-referral-table.tsx`

**Step 1: Write the table component**

This component holds page state, calls the server action on page change, and renders the table with pagination.

```typescript
"use client";

import { useState, useTransition } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  REFERRAL_STATUS_LABELS,
} from "@/consts/referral-status";
import type { ReferralStatus } from "@/consts/referral-status";
import { DataTablePagination } from "@/components/in/data-table-pagination";
import { fetchReferralPage } from "@/features/web/invite/actions/invite.action";
import {
  AVATAR_COLORS,
  getDisplayName,
  maskEmail,
  formatDate,
  formatReward,
  getStatusClasses,
} from "@/features/web/invite/helpers/referral-table.helper";

import type { findReferralsByReferrerId } from "@/models/user-referral/user-referral.query";

type Referral = Awaited<ReturnType<typeof findReferralsByReferrerId>>[number];

type InviteReferralTableProps = {
  initialReferrals: Referral[];
  initialTotalPages: number;
};

/** Table displaying paginated invite referral records */
export function InviteReferralTable({
  initialReferrals,
  initialTotalPages,
}: InviteReferralTableProps) {
  const [referrals, setReferrals] = useState(initialReferrals);
  const [totalPages, setTotalPages] = useState(initialTotalPages);
  const [currentPage, setCurrentPage] = useState(1);
  const [isPending, startTransition] = useTransition();

  /** Load a specific page of referrals */
  const handlePageChange = (page: number) => {
    startTransition(async () => {
      const result = await fetchReferralPage(page);
      if ("data" in result) {
        setReferrals(result.data.referrals);
        setTotalPages(result.data.totalPages);
        setCurrentPage(page);
      }
    });
  };

  return (
    <div className={isPending ? "opacity-60 transition-opacity" : ""}>
      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-slate-50 hover:bg-slate-50">
              <TableHead className="pl-6 text-xs font-semibold text-slate-500">
                好友信息
              </TableHead>
              <TableHead className="w-[120px] text-xs font-semibold text-slate-500">
                注册日期
              </TableHead>
              <TableHead className="w-[100px] text-xs font-semibold text-slate-500">
                状态
              </TableHead>
              <TableHead className="w-[120px] text-xs font-semibold text-slate-500">
                佣金
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {referrals.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="py-10 text-center text-sm text-slate-400">
                  暂无邀请记录
                </TableCell>
              </TableRow>
            )}
            {referrals.map((referral, index) => {
              const colors = AVATAR_COLORS[index % AVATAR_COLORS.length];
              const displayName = getDisplayName(referral.invitee);
              const initial = displayName.charAt(0);

              return (
                <TableRow key={referral.id}>
                  <TableCell className="pl-6">
                    <div className="flex items-center gap-2.5">
                      <div
                        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full ${colors.bg}`}
                      >
                        <span className={`text-[13px] font-semibold ${colors.text}`}>
                          {initial}
                        </span>
                      </div>
                      <div className="flex flex-col gap-0.5">
                        <span className="text-[13px] font-semibold text-slate-900">
                          {displayName}
                        </span>
                        <span className="text-[11px] text-slate-400">
                          {maskEmail(referral.invitee?.email ?? null)}
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="text-[13px] text-slate-500">
                    {formatDate(referral.createdAt)}
                  </TableCell>
                  <TableCell>
                    <span
                      className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${getStatusClasses(referral.status)}`}
                    >
                      {REFERRAL_STATUS_LABELS[referral.status as ReferralStatus] ?? referral.status}
                    </span>
                  </TableCell>
                  <TableCell className="text-[13px] font-semibold text-teal-600">
                    {formatReward(referral.rewardAmount, referral.status)}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>

      <DataTablePagination
        currentPage={currentPage}
        totalPages={totalPages}
        onPageChange={handlePageChange}
      />
    </div>
  );
}
```

**Step 2: Verify**

Run: `npm run build` — may still fail until invite-content.tsx is updated (next task).

**Step 3: Commit**

```
feat: add InviteReferralTable with server-side pagination
```

---

### Task 7: Refactor InviteContent and Page

**Files:**
- Modify: `src/features/web/invite/components/invite-content.tsx`
- Modify: `src/app/(web)/hall/(main)/invite/page.tsx`

**Step 1: Update page.tsx to pass new props**

```typescript
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { InviteContent } from "@/features/web/invite/components/invite-content";
import { fetchInviteData } from "@/features/web/invite/services/invite.service";

export default async function InvitePage() {
  const { inviteUrl, referrals, totalPages, stats } = await fetchInviteData();

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="邀请推广"
        subtitle="邀请好友加入斗学，成功即可获得佣金奖励"
      />
      <InviteContent
        inviteUrl={inviteUrl}
        referrals={referrals}
        totalPages={totalPages}
        stats={stats}
      />
    </div>
  );
}
```

**Step 2: Rewrite invite-content.tsx**

Remove: `AVATAR_COLORS`, `getDisplayName`, `maskEmail`, `formatDate`, `formatReward`, `getStatusClasses`, the entire table rendering block (lines 180-261), the `Referral` type.

Replace with `InviteReferralTable`. Update filter labels. Receive `stats` as prop instead of computing client-side.

The component should:
- Accept `{ inviteUrl, referrals, totalPages, stats }` props
- Remove `computeInviteStats` call and import
- Remove `Referral` type alias (no longer needed here)
- Remove all 6 helper functions and `AVATAR_COLORS`
- Use `stats` prop directly for the stats row
- Replace the table `div` block with `<InviteReferralTable>`
- Update filter button labels to 全部/待激活/已激活

**Key changes to imports — remove:**
- `findReferralsByReferrerId` type import
- `REFERRAL_STATUS_LABELS`, `REFERRAL_STATUSES`, `ReferralStatus`
- `computeInviteStats`

**Key changes to imports — add:**
- `InviteReferralTable` from `./invite-referral-table`
- `InviteStats` type from `@/features/web/invite/helpers/invite-stats.helper`

**Updated props type:**
```typescript
type InviteContentProps = {
  inviteUrl: string;
  referrals: Awaited<ReturnType<typeof findReferralsByReferrerId>>;
  totalPages: number;
  stats: InviteStats;
};
```

**Replace table block (lines 179-261) with:**
```tsx
<div className="overflow-hidden rounded-[14px] border border-slate-200 bg-white">
  <div className="flex items-center justify-between px-4 py-4 md:px-6">
    <div className="flex items-center gap-2">
      <Users className="h-[18px] w-[18px] text-teal-600" />
      <span className="text-base font-semibold text-slate-900">
        邀请记录
      </span>
    </div>
    <div className="flex items-center gap-2">
      <button type="button" className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-500">
        全部
      </button>
      <button type="button" className="rounded-full px-3 py-1 text-xs font-medium text-slate-400">
        待激活
      </button>
      <button type="button" className="rounded-full px-3 py-1 text-xs font-medium text-slate-400">
        已激活
      </button>
    </div>
  </div>
  <InviteReferralTable
    initialReferrals={referrals}
    initialTotalPages={totalPages}
  />
</div>
```

**Update stats row** to use `stats` prop directly instead of `inviteStats`:
```tsx
const statCards = [
  { icon: DollarSign, iconColor: "text-teal-600", value: stats.totalReward, label: "累计获得推广佣金" },
  { icon: Users, iconColor: "text-blue-500", value: String(stats.totalFriends), label: `本月新增 ${stats.newThisMonth} 位好友` },
  { icon: UserCheck, iconColor: "text-amber-500", value: String(stats.pendingCount), label: "好友已注册待验证" },
  { icon: TrendingUp, iconColor: "text-purple-600", value: stats.conversionRate, label: "邀请成功转化比例" },
];
```

**Step 3: Verify**

Run: `npm run build` — should compile with no errors.
Run: `npm run dev` — navigate to `/hall/invite`, verify:
- Table shows max 15 rows
- Pagination appears at bottom right when >15 referrals
- Page navigation works
- Stats reflect all-time totals
- Filter labels read 全部/待激活/已激活

**Step 4: Commit**

```
refactor: wire up paginated invite table and update filter labels
```

---

### Task 8: Export InviteStats Type

**Files:**
- Modify: `src/features/web/invite/helpers/invite-stats.helper.ts`

**Step 1: Ensure the `InviteStats` type is exported**

Check if `InviteStats` is already exported. If not, add `export` to its type declaration so it can be imported by `invite-content.tsx`.

**Step 2: Verify**

Run: `npm run build` — should compile cleanly.

**Step 3: Commit (if changes needed)**

```
refactor: export InviteStats type
```

---

### Task 9: Final Verification and Lint

**Step 1: Run linter**

Run: `npm run lint`

Fix any issues.

**Step 2: Run build**

Run: `npm run build`

Confirm clean build.

**Step 3: Manual test**

- Navigate to `/hall/invite`
- Verify table renders with max 15 rows
- Click pagination buttons — pages load correctly
- First/last page buttons work
- Ellipsis appears when many pages
- "Page X of Y" label updates
- Loading state (opacity) shows during page transitions
- Stats row unchanged (all-time totals)
- Filter buttons show: 全部 / 待激活 / 已激活
- Empty state still works ("暂无邀请记录")

**Step 4: Commit any fixes**

```
fix: address lint and build issues
```
