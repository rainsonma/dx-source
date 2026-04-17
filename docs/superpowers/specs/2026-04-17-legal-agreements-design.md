# Legal Agreements — Design Spec

- **Date:** 2026-04-17
- **Author:** brainstormed with rainson
- **Status:** Spec — awaiting final user review before writing implementation plan
- **Scope:** dx-web only (no `dx-api` or `deploy` changes)

## 1. Goal

Ship Douxue's own 5 legal documents — 用户协议, 隐私政策, 监护人同意书, 产品服务协议, Cookie 政策 — written for China users, based on (not copied from) the 4 reference `.md` files in `docs/` (which describe 句乐部 / julebu.co, not 斗学). Wire them into:

1. A new set of `/docs/account/<slug>` topic pages.
2. A reusable "full-text" `AgreementDialog` opened from `/auth/signup`, `/auth/signin`, and `/purchase/payment/[orderId]`.
3. The site footer's `服务条款` column (replaces current non-clickable `<span>` mock links).

No new routes, no new API calls, no DB changes.

## 2. Constraints & inputs (decided during brainstorm)

### 2.1 Legal posture — Hybrid with placeholders

Product-level facts (brand, domain, membership tiers, energy beans, community, groups, age rules, session policy, auth methods) are stated concretely. Hard legal-entity facts are rendered as visible `{{placeholders}}`:

```
{{公司名称}}              {{公司注册地址}}
{{统一社会信用代码}}       {{ICP 备案号}}
{{公安备案号}}            {{管辖法院所在地}}
{{数据存储地}}            {{客服邮箱}}
```

All placeholders live in `src/features/com/legal/constants.ts` — one find-and-replace when legal facts are confirmed.

### 2.2 Per-page agreement lists

| Page | Agreements shown |
|---|---|
| `/auth/signup` | 用户协议、隐私政策、监护人同意书、产品服务协议 |
| `/auth/signin` | 用户协议、隐私政策、监护人同意书、Cookie 政策 |
| `/purchase/payment/[orderId]` | 产品服务协议 |
| Footer `服务条款` column | All five |

### 2.3 Modal UX — full content, not summary

Clicking any agreement name opens a scrollable Dialog that contains **the full legal text** (no TOC, no summary). Dialog footer has a "在完整页面查看 →" link to `/docs/account/<slug>` (new tab, preserves form state) and an "我已阅读" close button. "我已阅读" does **not** auto-check the parent consent checkbox — consent remains an explicit form action.

### 2.4 Product-reality guardrails baked into the content

| Area | Decision |
|---|---|
| Payment channels | Vendor-agnostic language ("平台支持的付费渠道") — no WeChat/Alipay name-dropping |
| Auto-renewal | **Omit.** Not implemented; add clauses only when feature ships |
| 推广返利 / 佣金 | Pointer only — "具体规则以平台《推广规则》公示为准" |
| 未成年人 | Full tiered guardian clauses (8+ / 8–16 / 16–18), reference style |
| 数据存储地 | Placeholder |
| 增值税发票 | Omit until the flow is built |
| 客服渠道 | Placeholder `{{客服邮箱}}` + in-app 反馈 link only; no 微信服务号 |
| 虚拟商品 | 能量豆 (not 钻石) |
| 会员档位 | Generic references (月度/季度/年度/终身) — no hardcoded prices |

### 2.5 Brand / domain

- `BRAND = "斗学"`
- `DOMAIN = "douxue.fun"` (confirmed — not `douxue.cc`)
- No hardcoded support email (`bs@douxue.cc` removed from footer; legal docs use the `{{客服邮箱}}` placeholder)

### 2.6 Reference docs — excluded from git

The 4 `.md` files in `docs/` (`用户协议-登录|注册页.md`, `隐私政策 - 登录|注册页.md`, `监护人同意书-登录|注册页.md`, `产品服务协议 - 登录|注册|付费页.md`) are reference-only inputs and have been added to `.gitignore`. They never leave the author's local checkout.

## 3. Module layout

New shared module (used by both `/docs` topic components and the three forms):

```
src/features/com/legal/
├── constants.ts                    BRAND, DOMAIN, PLACEHOLDERS map
├── registry.ts                     LEGAL_AGREEMENTS: LegalAgreement[]
├── types.ts                        LegalAgreement type
├── documents/
│   ├── user-agreement.tsx          用户协议
│   ├── privacy-policy.tsx          隐私政策
│   ├── guardian-consent.tsx        监护人同意书
│   ├── product-service.tsx         产品服务协议
│   └── cookie-policy.tsx           Cookie 政策
└── components/
    ├── agreement-dialog.tsx        shadcn Dialog — sticky header + scroll body + sticky footer
    ├── agreement-link.tsx          Inline `《...》` trigger button (opens dialog)
    ├── agreement-inline-list.tsx   "登录即代表同意 X、Y、Z" hint line for /signin
    └── legal-placeholder-notice.tsx  Top-of-doc callout: 法律信息占位
```

`src/features/com/` is the established "shared between web/adm" bucket (see `src/features/com/` in the existing tree). Placing legal under `com/` keeps the option open for admin consoles to surface the same documents later without duplication.

## 4. Data model

```ts
// src/features/com/legal/types.ts
import type { ComponentType } from "react";

export type LegalAgreementSlug =
  | "user-agreement"
  | "privacy-policy"
  | "guardian-consent"
  | "product-service"
  | "cookie-policy";

export type LegalAgreement = {
  slug: LegalAgreementSlug;
  title: string;           // 用户协议
  shortTitle: string;      // 《用户协议》 — used by AgreementLink
  description: string;     // used on /docs category index
  effectiveDate: string;   // YYYY-MM-DD
  lastUpdated: string;     // YYYY-MM-DD
  Component: ComponentType;
};
```

```ts
// src/features/com/legal/constants.ts
export const BRAND = "斗学";
export const DOMAIN = "douxue.fun";

export const PLACEHOLDERS = {
  companyName:   "{{公司名称}}",
  companyAddr:   "{{公司注册地址}}",
  uscc:          "{{统一社会信用代码}}",
  icpNumber:     "{{ICP 备案号}}",
  pscRecordNo:   "{{公安备案号}}",
  courtLocation: "{{管辖法院所在地}}",
  dataStorage:   "{{数据存储地}}",
  supportEmail:  "{{客服邮箱}}",
} as const;
```

## 5. Content conventions (every doc)

1. `<LegalPlaceholderNotice />` at the very top — amber `DocCallout` title **"法律信息占位"**, body lists which `{{...}}` tokens appear, ending with *"本页 `{{…}}` 标记的字段需在法律团队审阅后替换为正式内容，当前版本用于产品功能验证，非最终法律文本。"*
2. Metadata line: `生效日期：2026-04-17 · 最近更新：2026-04-17`
3. Preamble paragraph (adapted to 斗学's real product scope).
4. Numbered 章节 rendered via `<DocSection id="..." title="第X条 ..." />` (matches existing `/docs` convention).
5. Emphasis via `<strong>` for 免责 / 限制责任 / 加重义务 clauses (matches reference convention and existing docs).
6. Warnings via `<DocCallout variant="warning">`.
7. Ordered enumerations via `<ol>` / `<ul>`.
8. Cross-references to sibling legal docs use `<AgreementLink slug="..." />` so internal mentions are hot-linked inside the modal too.

**Placeholder visual treatment:**
```tsx
<span className="rounded bg-amber-50 px-1 py-0.5 font-mono text-[13px] text-amber-700">
  {PLACEHOLDERS.companyName}
</span>
```

**Approximate lengths (中文字数):**
- 用户协议 ~3000
- 隐私政策 ~3500
- 监护人同意书 ~1800
- 产品服务协议 ~3200
- Cookie 政策 ~1800

## 6. `/docs` integration

### 6.1 Placement

New topics go inside the **existing** `account` (账户与帮助) category, appended after the current 5 topics. A visual group marker `法律条款` separates them.

Final sidebar order for 账户与帮助:
```
个人资料 · 账号安全 · 通知中心 · 提交反馈 · 常见问题
── 法律条款 ──
用户协议 · 隐私政策 · 监护人同意书 · 产品服务协议 · Cookie 政策
```

### 6.2 `DocTopic.groupLabel` — additive, opt-in field

One tiny type change enables the in-category divider without restructuring any existing data:

```ts
// src/features/web/docs/types.ts (diff)
 export type DocTopic = {
   slug: string;
   title: string;
   description: string;
   Component: ComponentType;
+  groupLabel?: string; // when set, sidebar renders a divider heading above this topic
 };
```

Only `user-agreement` gets `groupLabel: "法律条款"`. All 11 other categories and their topics are untouched.

### 6.3 Sidebar / drawer / category-index render change

`docs-sidebar.tsx`, `docs-sidebar-drawer.tsx`, and `docs-category-index.tsx` each gain one extra branch in the topic loop:

```tsx
{topic.groupLabel && (
  <div className="mt-3 mb-1 px-3 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
    {topic.groupLabel}
  </div>
)}
```

### 6.4 Registry wiring

```ts
// src/features/web/docs/registry.ts — inside the `account` category, appended after Faq:
{ slug: "user-agreement",    title: "用户协议",     description: "...",
  Component: UserAgreementDoc,    groupLabel: "法律条款" },
{ slug: "privacy-policy",    title: "隐私政策",     description: "...",
  Component: PrivacyPolicyDoc },
{ slug: "guardian-consent",  title: "监护人同意书", description: "...",
  Component: GuardianConsentDoc },
{ slug: "product-service",   title: "产品服务协议", description: "...",
  Component: ProductServiceDoc },
{ slug: "cookie-policy",     title: "Cookie 政策",  description: "...",
  Component: CookiePolicyDoc },
```

Each `LEGAL_AGREEMENTS` entry in `features/com/legal/registry.ts` exports its `Component` directly; `features/web/docs/registry.ts` imports and registers that same component — one source, two consumers. No wrapper files.

```ts
// features/com/legal/registry.ts
import { UserAgreementDoc } from "./documents/user-agreement";
// ...
export const LEGAL_AGREEMENTS: LegalAgreement[] = [
  { slug: "user-agreement", title: "用户协议", shortTitle: "《用户协议》",
    description: "...", effectiveDate: "2026-04-17", lastUpdated: "2026-04-17",
    Component: UserAgreementDoc },
  // ...4 more
];
```

```ts
// features/web/docs/registry.ts — imports components from the legal registry
import {
  UserAgreementDoc,
  PrivacyPolicyDoc,
  GuardianConsentDoc,
  ProductServiceDoc,
  CookiePolicyDoc,
} from "@/features/com/legal/registry";
```

### 6.5 URLs

- `/docs/account/user-agreement`
- `/docs/account/privacy-policy`
- `/docs/account/guardian-consent`
- `/docs/account/product-service`
- `/docs/account/cookie-policy`

`generateStaticParams` picks these up automatically via the existing `DOC_CATEGORIES` traversal.

### 6.6 Prev/next traversal

`findTopic` uses flat-index traversal across all categories. The 5 legal topics chain naturally: `常见问题 → 用户协议 → 隐私政策 → 监护人同意书 → 产品服务协议 → Cookie 政策 → (end)`.

## 7. `AgreementDialog` UX

Built on shadcn `Dialog` (already in `src/components/ui/dialog.tsx`). Focus trap, `Esc` close, `aria-labelledby`, body scroll-lock — all free.

```tsx
<Dialog open={open} onOpenChange={setOpen}>
  <DialogContent className="flex h-[85vh] max-h-[85vh] w-full max-w-[720px] flex-col gap-0 p-0">
    <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-6 py-4">
      <div className="flex flex-col gap-1">
        <DialogTitle className="text-lg font-bold text-slate-900">{agreement.title}</DialogTitle>
        <DialogDescription className="text-xs text-slate-500">
          生效日期：{agreement.effectiveDate} · 最近更新：{agreement.lastUpdated}
        </DialogDescription>
      </div>
      {/* shadcn close ✕ */}
    </div>
    <div className="flex-1 overflow-y-auto px-6 py-5">
      <div className="flex flex-col gap-6 text-[14px] leading-[1.75] text-slate-700">
        <agreement.Component />
      </div>
    </div>
    <div className="flex items-center justify-between gap-3 border-t border-slate-200 bg-slate-50 px-6 py-3">
      <Link href={`/docs/account/${agreement.slug}`} target="_blank"
            className="text-sm font-medium text-teal-600 hover:text-teal-700">
        在完整页面查看 →
      </Link>
      <button type="button" onClick={() => setOpen(false)}
              className="h-9 rounded-md bg-teal-600 px-4 text-sm font-semibold text-white hover:bg-teal-700">
        我已阅读
      </button>
    </div>
  </DialogContent>
</Dialog>
```

**Key behaviors:**
- Body scrolls; header + footer stay visible.
- "我已阅读" closes the dialog only. It does **not** auto-check the parent consent checkbox (reading ≠ agreeing).
- "在完整页面查看 →" opens `/docs/account/{slug}` in a new tab, preserving form state.
- Triggered per-link: each `AgreementLink` owns its own `open` state. Dialog body only renders when open (Radix lazy-renders children).

### 7.1 `AgreementLink`

```tsx
<AgreementLink slug="user-agreement" />
// Renders:
// <button type="button" className="...">《用户协议》</button>
```

- `<button>`, not `<a>` — no navigable URL for a modal, and matches the recent fix `3f7fea6` that replaced nested anchors.
- `type="button"` so it never submits the containing form.
- Inherits text sizing from parent so it slots cleanly into sentences.

### 7.2 `AgreementInlineList`

For `/signin` (no consent checkbox — just an informational footer line):

```tsx
<AgreementInlineList
  prefix="登录即代表同意"
  slugs={["user-agreement", "privacy-policy", "guardian-consent", "cookie-policy"]}
  className="text-xs text-slate-400"
/>
// 登录即代表同意《用户协议》、《隐私政策》、《监护人同意书》、《Cookie 政策》
```

## 8. Per-page wiring

### 8.1 `SignUpForm` (`src/features/web/auth/components/sign-up-form.tsx`, lines 202–208)

Replace the 4 static `<span>`s inside the agreement checkbox label:

```tsx
<span className="text-xs text-slate-700">
  我已阅读并同意{" "}
  <AgreementLink slug="user-agreement" />、
  <AgreementLink slug="privacy-policy" />、
  <AgreementLink slug="guardian-consent" />、
  <AgreementLink slug="product-service" />
</span>
```

Everything else on the form stays — `agreed` checkbox, submit gate, etc.

### 8.2 `SignInForm` (`src/features/web/auth/components/sign-in-form.tsx`)

Add a hint line below the "没有账号？立即注册" row, inside the form card footer area:

```tsx
<div className="h-px bg-slate-200" />
<AgreementInlineList
  prefix="登录即代表同意"
  slugs={["user-agreement", "privacy-policy", "guardian-consent", "cookie-policy"]}
  className="text-xs text-slate-400"
/>
<div className="flex items-center justify-between">
  {/* existing 没有账号？立即注册 row */}
</div>
```

No checkbox added — signin implies an existing account with prior consent.

### 8.3 `OrderPayment` (`src/features/web/purchase/components/order-payment.tsx`, line 163)

Replace the fake `《斗学服务协议》` span with the single required link:

```tsx
<div className="flex flex-col gap-1">
  <span className="text-sm text-slate-700">我已阅读并同意以下协议</span>
  <AgreementLink slug="product-service" />
</div>
```

### 8.4 `Footer` (`src/components/in/footer.tsx`)

Three edits:

1. Extract a new `LEGAL_LINKS` constant and render the 服务条款 column out of the `.map()` with real `<Link href>` anchors:

```tsx
const LEGAL_LINKS: { label: string; href: string }[] = [
  { label: "用户协议",     href: "/docs/account/user-agreement" },
  { label: "隐私政策",     href: "/docs/account/privacy-policy" },
  { label: "监护人同意书", href: "/docs/account/guardian-consent" },
  { label: "产品服务协议", href: "/docs/account/product-service" },
  { label: "Cookie 政策",  href: "/docs/account/cookie-policy" },
];
```

2. Remove `"bs@douxue.cc"` from the 斗学团队 column's `links` array.

3. Update copyright line from `douxue.cc` to `douxue.fun`.

The other two columns (斗学产品, 斗学团队) keep their current mock `<span>` behavior — explicitly out of scope.

## 9. Touched files

| File | Change |
|---|---|
| `src/features/com/legal/constants.ts` | **NEW** |
| `src/features/com/legal/types.ts` | **NEW** |
| `src/features/com/legal/registry.ts` | **NEW** |
| `src/features/com/legal/documents/user-agreement.tsx` | **NEW** |
| `src/features/com/legal/documents/privacy-policy.tsx` | **NEW** |
| `src/features/com/legal/documents/guardian-consent.tsx` | **NEW** |
| `src/features/com/legal/documents/product-service.tsx` | **NEW** |
| `src/features/com/legal/documents/cookie-policy.tsx` | **NEW** |
| `src/features/com/legal/components/agreement-dialog.tsx` | **NEW** |
| `src/features/com/legal/components/agreement-link.tsx` | **NEW** |
| `src/features/com/legal/components/agreement-inline-list.tsx` | **NEW** |
| `src/features/com/legal/components/legal-placeholder-notice.tsx` | **NEW** |
| `src/features/web/docs/types.ts` | Add `groupLabel?: string` |
| `src/features/web/docs/registry.ts` | Append 5 topics under `account` |
| `src/features/web/docs/components/docs-sidebar.tsx` | Render `groupLabel` divider |
| `src/features/web/docs/components/docs-sidebar-drawer.tsx` | Same divider |
| `src/features/web/docs/components/docs-category-index.tsx` | Same divider |
| `src/features/web/auth/components/sign-up-form.tsx` | Replace 4 spans with `AgreementLink` |
| `src/features/web/auth/components/sign-in-form.tsx` | Add `AgreementInlineList` |
| `src/features/web/purchase/components/order-payment.tsx` | Replace span with `AgreementLink` |
| `src/components/in/footer.tsx` | Real 服务条款 links, drop email, fix domain |
| `.gitignore` | Ignore the 4 reference `.md` files in `docs/` |

Zero files in `dx-api/`, `deploy/`, `components/ui/` are touched.

## 10. Verification gates

### 10.1 Correctness (in order)

1. `cd dx-web && npx tsc --noEmit` — zero errors.
2. `cd dx-web && npm run lint` — zero warnings, zero errors. (Expect to satisfy `react/no-unescaped-entities` via `&ldquo;/&rdquo;/&amp;` escapes — same as existing `/docs` topic files.)
3. `cd dx-web && npm run build` — success. Static params include the 5 new legal topic routes.

### 10.2 Functional (localhost:3000 via `npm run dev`)

| Flow | Expected |
|---|---|
| `/docs/account` | Category index shows all 10 topics; 法律条款 divider above 用户协议 |
| `/docs/account/user-agreement` | Full doc renders; sidebar shows divider; prev = 常见问题, next = 隐私政策 |
| Walk all 5 legal URLs via prev/next | Continuous chain; final doc has no `next` |
| `/auth/signup` → 《用户协议》 | Dialog opens; scrollable body; header+footer sticky |
| Dialog → "在完整页面查看 →" | New tab → `/docs/account/user-agreement`; signup form state preserved |
| Signup uncheck agreement | Submit button disabled (existing behavior unchanged) |
| `/auth/signin` | Hint line visible; 4 clickable agreement names open correct dialogs |
| `/purchase/payment/[orderId]` (any PENDING order) | Only 《产品服务协议》 link; dialog opens |
| Footer anywhere | 服务条款 column has 5 working anchors; `bs@douxue.cc` gone; copyright reads `douxue.fun` |
| Mobile viewport (375px) | Dialog full-width; sticky header + footer reachable |
| Esc / outside click / ✕ | All close the dialog |

If any row fails, fix before the next row.

### 10.3 Regression safety

- All 11 non-`account` categories render identically (`groupLabel` is opt-in).
- Signup `agreed` state machine unchanged.
- Payment `agreed` state machine unchanged.
- Footer 斗学产品 and 斗学团队 columns visually identical except the removed email item.
- No new routes, no new API calls, no DB changes, no `dx-api` edits.

## 11. Out of scope (explicit)

- Making 斗学产品 / 斗学团队 footer columns clickable.
- Cookie consent banner on first visit (the policy exists, UI doesn't).
- Real-name verification / age-gate enforcement / guardian signature capture.
- English translations.
- Filling in the `{{placeholders}}` with real legal-entity values (that's a find-and-replace after legal review).

## 12. Commit strategy

Single feature branch; each commit builds + lints in isolation:

1. `feat(web): add legal document registry and dialog infrastructure` — `features/com/legal/**` skeleton, types, constants, empty documents, `AgreementDialog` / `AgreementLink` / `AgreementInlineList` / `LegalPlaceholderNotice`. No wiring yet.
2. `feat(web): write 5 legal documents for 斗学` — the 5 document TSX files.
3. `feat(web): add groupLabel divider to docs sidebar` — `types.ts` + sidebar / drawer / category-index render.
4. `feat(web): register 5 legal topics under 账户与帮助` — `features/web/docs/registry.ts` edits.
5. `feat(web): wire agreement dialogs into signup/signin/payment` — 3 form edits.
6. `feat(web): link real legal docs from footer, update domain` — `footer.tsx` edits.
7. `chore: ignore reference legal source docs` — `.gitignore` edit (can be first commit; listed here for completeness).

## 13. Open risks

- **Legal liability of shipped placeholders.** The `法律信息占位` callout and the `{{placeholder}}` chips make the unfinished state visible, but the docs are still legally displayable. Product/legal owner must review before a public launch. This spec does not claim the docs are legally complete.
- **Content drift vs. product reality.** Future product features (real payment, auto-renewal, guardian verification UI) must trigger a matching doc update. There's no automated link between code and legal text — this is a human-review process.
- **Placeholder discoverability at replace time.** A single search for `{{` across `features/com/legal/` reveals every occurrence. `constants.ts` is the single source — don't inline placeholder strings in document files.
