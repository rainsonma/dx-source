# Hall Invite Rules Redesign — Design Spec

**Date:** 2026-04-13
**Scope:** `dx-web` frontend only (no backend changes)
**Target page:** `/hall/invite`
**Target file (primary):** `dx-web/src/features/web/invite/components/invite-content.tsx`

## Purpose

Replace the existing hardcoded `活动规则` (Activity Rules) section on the `/hall/invite` page with a richer rules block that includes:

1. A renumbered 5-item rules list reflecting the current invite program policy
2. A `佣金体系` (Commission System) block comparing rewards for two inviter tiers
3. A `被邀请者专属折扣` (Invitee Exclusive Discount) block listing discounts for invited purchases
4. Grade-based gating: commission and discount blocks are hidden for free users and replaced with a single locked-state placeholder that nudges upgrade

No backend edits. The existing `/api/user/profile` endpoint already returns `grade`; we consume it client-side via SWR.

## Current State

The existing section lives in `dx-web/src/features/web/invite/components/invite-content.tsx` and consists of:

- A hardcoded `rules` array at module scope (5 string items)
- A rendered card using `rounded-[14px] border bg-card p-4 lg:p-6` with:
  - Header: `ScrollText` icon + `活动规则` label
  - Numbered list rendered via `rules.map`
  - An amber info alert at the bottom warning that "佣金奖励仅限首次购买会员时发放"

The amber alert directly contradicts the new "持续续费返佣" rule and must be removed.

## New Content

### Rules list (5 items, renumbered)

1. 邀请好友通过您的专属链接、邀请码或二维码注册斗学账号
2. 好友成功注册并完成首次购买会员即算邀请成功
3. 佣金数额实时反馈，邀请所得收入清晰体现，一目了然
4. 邀请人数不设上限，邀请越多佣金越多
5. 佣金收入按时结算，可随时申请提现

> Note: the original user spec numbered these 1, 2, 4, 5, 6 with a missing `3`. This is treated as a typo and renumbered to 1–5.

### Commission system

Two tiers compared side-by-side.

**Tier 1 — 普通付费会员** (month / season / year paid members):

| Reward category | Value |
|---|---|
| 邀请永久会员 | ¥500 |
| 邀请年度会员 | 30% |
| 邀请季度会员 | 30% |
| 邀请月度会员 | 30% |
| 持续续费返佣 | 10% |

**Tier 2 — 终身会员** (lifetime members):

| Reward category | Value |
|---|---|
| 邀请永久会员 | ¥600 |
| 邀请年度会员 | 50% |
| 邀请季度会员 | 50% |
| 邀请月度会员 | 50% |
| 持续续费返佣 | 20% |

### Invitee exclusive discounts

| Membership | Discount |
|---|---|
| 购买永久会员 | ¥99 off the ¥1999 lifetime price (pays ¥1900) |
| 购买年度会员 | 10% off |
| 购买季度会员 | 10% off |
| 购买月度会员 | 10% off |

Footnote: `未获邀请直接购买者无折扣优惠`

## Architecture

### File structure

| Action | Path | Purpose |
|---|---|---|
| NEW | `dx-web/src/features/web/invite/helpers/invite-rules.ts` | Pure data module with typed constants and formatters |
| NEW | `dx-web/src/features/web/invite/components/invite-rules-section.tsx` | Client component rendering the section with gating |
| MODIFY | `dx-web/src/features/web/invite/components/invite-content.tsx` | Replace inline rules card with `<InviteRulesSection />`, remove old `rules` array, clean up imports, thread `userGrade` prop |
| MODIFY | `dx-web/src/app/(web)/hall/(main)/invite/page.tsx` | Add SWR fetch for `/api/user/profile`, pass `userGrade` down |

### Data module — `invite-rules.ts`

Pure TypeScript, no React or Next.js imports. Exports:

- `inviteRules: string[]` — the 5-item rules list
- `RewardValue` — tagged union: `{ kind: "fixed"; amount: number } | { kind: "percent"; value: number }`
- `InviterTierId` — `"standard" | "lifetime"`
- `CommissionTier` — typed tier config with id, label, sublabel, and rewards map
- `commissionTiers: CommissionTier[]` — the two tiers populated with values
- `InviteeDiscount` — typed discount entry
- `inviteeDiscounts: InviteeDiscount[]` — the four discount entries
- `rewardRowLabels` — `as const` record mapping reward keys to Chinese labels
- `formatRewardValue(value)` — single formatter returning `¥500` for fixed amounts or `30%` for percentages. Reused for both commission rewards and invitee discounts since both use the same `RewardValue` type, and the surrounding section headers (`佣金体系` vs `被邀请者专属折扣`) provide the semantic context.

The tagged-union approach enforces that every reward has an explicit `kind` and forces exhaustive handling via TypeScript's switch narrowing. Adding a third variant in the future will surface as a compile error in the formatter.

### Section component — `invite-rules-section.tsx`

Client component (`"use client"`). Accepts a single prop:

```typescript
type Props = {
  userGrade: UserGrade | null;  // null = still loading
};
```

Internal sub-components (colocated in the same file because they are tightly coupled and have zero reuse potential):

- `SectionHeader` — `ScrollText` icon + `活动规则` label
- `RulesList` — numbered 5-item list (reuses the existing teal-600 numbering style)
- `CommissionTiersBlock` — sub-header with accent bar + 2-col grid of `CommissionTierCard`
- `CommissionTierCard` — single tier card rendering reward rows
- `InviteeDiscountsBlock` — sub-header + list + footnote
- `LockedHint` — dashed-border placeholder shown to free users
- `LoadingPlaceholder` — muted skeleton shown until SWR resolves

### Page integration

In `dx-web/src/app/(web)/hall/(main)/invite/page.tsx`:

- Add a second data fetch using SWR: `const { data, error } = useSWR<ApiProfileData>("/api/user/profile")`
- Compute `userGrade`:
  - If `error` is set → `"free"` (conservative: never leak gated content when profile fetch fails)
  - Else if `data?.grade` is set → cast to `UserGrade`
  - Else → `null` (still loading)
- Pass `userGrade` as a new prop to `<InviteContent />`
- The existing `useEffect`-based fetch for `/api/invite` is untouched — this is NOT a migration to SWR

In `dx-web/src/features/web/invite/components/invite-content.tsx`:

- Remove the `rules` array (old content)
- Remove the rules card JSX block (old card structure)
- Remove imports for `ScrollText` and `Info` (relocated to the new section component)
- Add `userGrade: UserGrade | null` to `InviteContentProps`
- Render `<InviteRulesSection userGrade={userGrade} />` in place of the removed rules card

## Data flow

```
InvitePage (page.tsx)
  │
  ├─ useEffect → apiClient.get<ApiInviteData>("/api/invite")   [unchanged]
  └─ useSWR<ApiProfileData>("/api/user/profile")                 [new]
       │
       └─ userGrade extracted and passed down
            │
            ▼
        <InviteContent userGrade={grade} ... />
                                │
                                ▼
                  <InviteRulesSection userGrade={grade} />
                                │
                                ▼
       if (userGrade === null)          → <LoadingPlaceholder />
       else if (userGrade === "free")   → <LockedHint />
       else                              → <CommissionTiersBlock /> + <InviteeDiscountsBlock />
```

SWR deduplicates `/api/user/profile` calls across components — if the user visits `/hall/me` before `/hall/invite`, the profile is already cached (the me page uses the same SWR key).

## Gating logic

### Free users (`grade === "free"`)

Show only the rules list and a single `LockedHint` placeholder in place of both the commission tiers and the invitee discounts. The hint reads:

```
🔒
佣金体系与会员折扣仅会员可见
升级会员解锁完整邀请奖励
[升级会员] ← button linking to /purchase/membership
```

Styling: `min-h-[280px] rounded-[10px] border border-dashed border-border/60 bg-muted/30 p-6 flex flex-col items-center justify-center gap-3`. The lock icon is `lucide-react`'s `Lock` at `h-6 w-6 text-muted-foreground`. The CTA button uses `bg-teal-600 text-white` to match the page's existing teal accent.

The `min-h-[280px]` roughly matches the combined vertical height of the two hidden sections (commission tiers ~200px + discount list ~100px + gap), keeping layout shift minimal when transitioning between grades.

### Paid users (`grade` in `month | season | year | lifetime`)

Show the full commission tiers block and the invitee discounts block. No upgrade CTA.

### Loading state (`userGrade === null`)

Show `LoadingPlaceholder` — a subtle muted skeleton with the same `min-h-[280px]` footprint as the LockedHint, in the same slot as the commission/discount area. Rules list above is always visible during loading. This avoids flashing full content to a free user before SWR resolves and keeps the layout stable across state transitions.

Styling: `min-h-[280px] rounded-[10px] border border-border/40 bg-muted/10 animate-pulse`. No text — just a muted pulse.

## Styling details

### Inherited from existing page

- Card surfaces: `rounded-[14px] border border-border bg-card`
- Padding: `p-4 lg:p-6`
- Primary accent color: `text-teal-600` / `bg-teal-600`
- Muted text: `text-muted-foreground`
- Section gap: `gap-4 lg:gap-6`

### Section sub-headers (佣金体系, 被邀请者专属折扣)

```
flex items-center gap-2
  span: h-3 w-1 rounded-full bg-teal-600  ← small accent bar
  h3: text-sm font-semibold text-foreground
```

### Commission tier cards

- Standard tier: `rounded-[10px] border border-border bg-card p-4`
- Lifetime tier: `rounded-[10px] border border-teal-500/40 bg-gradient-to-b from-teal-50/40 to-transparent p-4` + small `Crown` icon next to the label
- Grid: `grid grid-cols-1 lg:grid-cols-2 gap-4`
- Reward rows inside each card:
  ```
  flex items-center justify-between py-2 border-b border-border/40 last:border-0
    label: text-sm text-muted-foreground
    value: text-sm font-semibold text-foreground
  ```

### Discount list

- Flat list (no nested card): each row is `flex items-center justify-between py-2 border-b border-border/40 last:border-0`
- Footnote: `text-xs text-muted-foreground pt-2` with the text `未获邀请直接购买者无折扣优惠`

## Lint & type safety

1. `npm run lint` in `dx-web` must pass with zero warnings
2. No `any` types, no `!` non-null assertions, no `@ts-ignore`
3. No unused imports in `invite-content.tsx` (remove `ScrollText`, `Info`, and the `rules` const)
4. No `console.log` (per CLAUDE.md)
5. Exhaustive switches on `RewardValue.kind` — TypeScript enforces completeness via narrowing
6. `next build` (or `tsc --noEmit`) must type-check successfully

## Non-goals (out of scope)

- Migrating `InvitePage` from `useEffect` to SWR for the existing invite data fetch
- Any backend changes to `dx-api`
- Adding i18n (Chinese text remains hardcoded, matching the rest of the page)
- Creating reusable `useCurrentUser()` hook (the SWR call is inlined for simplicity)
- Modifying `src/components/ui/` (forbidden by CLAUDE.md)
- Extracting commission values to `dx-api` constants (frontend-only display)
- Adding new dependencies

## Testing plan

### Manual verification in `npm run dev`

1. **Free user** — open `/hall/invite`
   - Rules list 1–5 visible above
   - Single locked hint card visible where the commission/discount sections would be
   - Lock icon + text + "升级会员" button rendered
   - Clicking "升级会员" navigates to `/purchase/membership`
2. **Paid user (month / season / year)** — open `/hall/invite`
   - Rules list 1–5 visible above
   - Two commission tier cards rendered side-by-side on desktop, stacked on mobile
   - Standard tier has plain border; lifetime tier has teal border + gradient + Crown icon
   - Invitee discounts list rendered below with footnote
3. **Lifetime user** — open `/hall/invite`
   - Same rendering as paid user (the UI doesn't gate differently for lifetime vs other paid tiers)
4. **Responsive behavior** — resize viewport below `lg` breakpoint (~1024px)
   - Commission tier cards stack vertically
   - All text remains readable, no overflow
5. **Regression check** — on the same page, verify untouched elements still work:
   - Banner and share button
   - Stats cards (4-col grid)
   - QR cards
   - Referral table + pagination
   - Copy link button
   - Share snippets modal

### Automated checks

- `cd dx-web && npm run lint` — must pass with zero output
- `cd dx-web && npm run build` — must complete successfully (TypeScript pass included)

### Testing honesty

The implementing agent cannot test all grades without logged-in sessions for each tier. The agent will verify lint, type-check, and at minimum one grade state that the current dev auth provides, and will explicitly report which grades were verified. Full cross-grade visual verification requires either test accounts or post-merge QA by the user.

## Risks & mitigations

| Risk | Mitigation |
|---|---|
| New SWR call adds a fetch on page mount | SWR caches and deduplicates; profile is small (~1 KB); negligible impact |
| Layout shift when profile loads and content swaps in | `LoadingPlaceholder` and `LockedHint` share `min-h-[280px]`, matching the approximate height of the full commission+discount area |
| Free user flashes full content before profile loads | Default `userGrade` to `null`; render `LoadingPlaceholder` until SWR resolves |
| `/api/user/profile` fetch fails with network error | Treat SWR error as `"free"` grade — conservative default never leaks gated content, worst case a paid user sees the locked hint until refresh |
| Adding `userGrade` prop to `InviteContent` breaks existing callers | Only one caller (`page.tsx`); updated in the same change |
| `/api/user/profile` returns an unexpected grade value | The `grade === "free"` check is strict equality, so any unknown value is treated as paid — user sees full content. Acceptable fallback for the happy path. |
| User's eye gets drawn to lock hint and ignores the rules list | Acceptable — the rules are visible above the hint and remain the first thing read |

## Open questions — none

All content and design decisions have been resolved in the brainstorming phase.
