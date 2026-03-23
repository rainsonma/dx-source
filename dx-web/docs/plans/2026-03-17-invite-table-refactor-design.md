# Invite Table Refactor Design

## Goal

Refactor the 邀请记录 data table in the invite page to use server-side pagination with a fixed 15 items per page and digital pagination controls at the bottom right.

## Requirements

- 15 items per page maximum
- Server-side pagination (offset/limit)
- Pagination controls: `Page X of Y  << < 1 2 3 ... N > >>`
- Filter button labels updated to: 全部 / 待激活 / 已激活 (non-functional)
- Page state is local (resets to page 1 on refresh)
- Stats row always reflects all-time totals, not current page

## Approach

Server action with client-side page state. Initial data fetched via SSR (page 1), subsequent pages fetched via server action.

## Data Layer

### Model (`src/models/user-referral/user-referral.query.ts`)

- Add `countReferralsByReferrerId(referrerId)` — returns total referral count
- Add pagination params to `findReferralsByReferrerId(referrerId, page, pageSize)` — `skip`/`take` based on page number, default `pageSize = 15`

### Service (`src/features/web/invite/services/invite.service.ts`)

- `fetchInviteData()` returns `{ inviteUrl, referrals, totalPages, stats }` where stats are pre-computed server-side using `computeInviteStats()` on all referrals

### Server Action (`src/features/web/invite/actions/invite.action.ts`)

- `fetchReferralPage(page: number)` — validates page, calls model with offset/limit, returns `{ referrals, totalPages }`

## Pagination Component

### Shared component (`src/components/in/data-table-pagination.tsx`)

Reusable pagination built on shadcn/ui Pagination primitives.

Props:
- `currentPage: number`
- `totalPages: number`
- `onPageChange: (page: number) => void`

Layout: `Page X of Y  << < 1 2 3 ... N > >>`

Behavior:
- First/last (`<<`/`>>`) and prev/next (`<`/`>`) buttons, disabled at boundaries
- Up to ~5 visible page numbers with ellipsis for gaps
- Active page highlighted
- Hidden when `totalPages <= 1`

## Table Refactor

### New component (`src/features/web/invite/components/invite-referral-table.tsx`)

- Receives initial `{ referrals, totalPages }`
- Holds `page` and `isLoading` state
- On page change: calls `fetchReferralPage()` server action
- Uses semantic `<Table>` from shadcn/ui
- Renders `DataTablePagination` at bottom right

### Extracted helpers (`src/features/web/invite/helpers/referral-table.helper.ts`)

Moved from `invite-content.tsx`:
- `getDisplayName()`, `maskEmail()`, `formatDate()`, `formatReward()`, `getStatusClasses()`, `AVATAR_COLORS`

### Updated `invite-content.tsx`

- Remove all table rendering code and helper functions
- Replace with `<InviteReferralTable />`
- Update filter labels to 全部/待激活/已激活
- Receive and pass `totalPages` and `stats` from service

### Updated `page.tsx`

- Pass `totalPages` and `stats` through to `InviteContent`

## File Changes

| File | Action |
|------|--------|
| `src/models/user-referral/user-referral.query.ts` | Add count query, add pagination params |
| `src/features/web/invite/services/invite.service.ts` | Return `totalPages` and `stats` |
| `src/features/web/invite/actions/invite.action.ts` | New — server action |
| `src/features/web/invite/helpers/referral-table.helper.ts` | New — extracted helpers |
| `src/components/in/data-table-pagination.tsx` | New — reusable pagination |
| `src/features/web/invite/components/invite-referral-table.tsx` | New — table with pagination |
| `src/features/web/invite/components/invite-content.tsx` | Slim down, update filter labels |
| `src/app/(web)/hall/(main)/invite/page.tsx` | Pass through new props |
