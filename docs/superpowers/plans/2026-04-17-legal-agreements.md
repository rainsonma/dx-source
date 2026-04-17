# Legal Agreements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship 5 Chinese legal agreements (用户协议, 隐私政策, 监护人同意书, 产品服务协议, Cookie 政策) for 斗学 as a reusable module; wire them into /docs/account/\*, into scrollable dialogs on /auth/signup, /auth/signin, /purchase/payment/\[orderId\], and into the footer 服务条款 column.

**Architecture:** New `src/features/com/legal/` module (types + constants + registry + 5 document components + 4 presentational components). The same document component is rendered both inline inside `/docs/account/<slug>` topic pages AND inside `AgreementDialog` triggered from form pages — one source, two surfaces. Hard legal-entity facts stay as visible `{{placeholders}}` until legal review.

**Tech Stack:** Next.js 16 (app router), React 19, Tailwind v4, shadcn/ui Dialog (Radix under the hood), TypeScript, existing `DocSection` / `DocCallout` / `DocSteps` primitives from `@/features/web/docs/primitives/*`.

**Spec reference:** `docs/superpowers/specs/2026-04-17-legal-agreements-design.md`

**Testing conventions for this project:** dx-web has NO unit-test framework (no Jest/Vitest). Verification in this plan uses:
- `npx tsc --noEmit` — type check (must be clean)
- `npm run lint` — ESLint (zero warnings, zero errors)
- `npm run build` — Next build (at milestone tasks only; slow)
- Browser checks on `npm run dev` (`localhost:3000`) at milestone tasks

Tasks commit after each focused change. Every task ends with a commit.

---

## File Structure

### New files (all under `dx-web/src/features/com/legal/`)

| File | Responsibility |
|---|---|
| `types.ts` | `LegalAgreementSlug` union + `LegalAgreement` type |
| `constants.ts` | `BRAND`, `DOMAIN`, `PLACEHOLDERS` map |
| `registry.ts` | `LEGAL_AGREEMENTS: LegalAgreement[]` + slug helpers |
| `documents/user-agreement.tsx` | 用户协议 doc — full content |
| `documents/privacy-policy.tsx` | 隐私政策 doc — full content |
| `documents/guardian-consent.tsx` | 监护人同意书 doc — full content |
| `documents/product-service.tsx` | 产品服务协议 doc — full content |
| `documents/cookie-policy.tsx` | Cookie 政策 doc — full content |
| `components/legal-placeholder-notice.tsx` | Top-of-doc amber 法律信息占位 callout |
| `components/agreement-dialog.tsx` | Scrollable modal: sticky header + scroll body + sticky footer |
| `components/agreement-link.tsx` | Inline `《...》` button that opens dialog |
| `components/agreement-inline-list.tsx` | "登录即代表同意《X》《Y》《Z》" hint line |

### Modified files

| File | Change |
|---|---|
| `dx-web/src/features/web/docs/types.ts` | Add `groupLabel?: string` to `DocTopic` |
| `dx-web/src/features/web/docs/registry.ts` | Import 5 legal doc components; append 5 topics under `account` category |
| `dx-web/src/features/web/docs/components/docs-sidebar.tsx` | Render `groupLabel` divider |
| `dx-web/src/features/web/docs/components/docs-category-index.tsx` | Render `groupLabel` section break between cards |
| `dx-web/src/features/web/auth/components/sign-up-form.tsx` | Replace 4 static `<span>`s with `AgreementLink` |
| `dx-web/src/features/web/auth/components/sign-in-form.tsx` | Add `AgreementInlineList` hint below card footer |
| `dx-web/src/features/web/purchase/components/order-payment.tsx` | Replace fake `<span>` with `AgreementLink slug="product-service"` |
| `dx-web/src/components/in/footer.tsx` | Real `<Link>` legal anchors; remove `bs@douxue.cc`; update copyright to `douxue.fun` |

No backend (`dx-api/`) changes. No `deploy/` changes. No changes to `dx-web/src/components/ui/` (shadcn-managed).

---

## Task 0: Baseline verification

**Files:** none (just verification)

- [ ] **Step 1: Confirm branch and clean tree**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git status
git log -1 --oneline
```

Expected: on `main`, clean working tree, HEAD at `37296ed docs: add legal agreements design spec` (or newer — any commit after the spec is fine as long as tree is clean).

- [ ] **Step 2: Baseline tsc/lint/build pass**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both exit 0 with no errors. (Skip `npm run build` at baseline — it's slow and will be run at milestone tasks only.)

If either fails at baseline, stop and report. Do not start the feature on a broken tree.

---

## Task 1: Add `groupLabel?` field to `DocTopic`

**Files:**
- Modify: `dx-web/src/features/web/docs/types.ts`

- [ ] **Step 1: Edit the type**

Current file has 4 fields on `DocTopic`. Add the 5th (optional):

```ts
// dx-web/src/features/web/docs/types.ts
import type { ComponentType } from "react";
import type { LucideIcon } from "lucide-react";

export type DocTopic = {
  slug: string;
  title: string;
  description: string;
  Component: ComponentType;
  groupLabel?: string;
};

export type DocCategory = {
  slug: string;
  title: string;
  description: string;
  icon: LucideIcon;
  accentClass: string;
  topics: DocTopic[];
};

export type TopicRef = {
  category: DocCategory;
  topic: DocTopic;
};
```

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
```

Expected: pass (field is optional, so no existing call sites break).

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/docs/types.ts
git commit -m "feat(web): add optional groupLabel field to DocTopic"
```

---

## Task 2: Render `groupLabel` divider in sidebar

**Files:**
- Modify: `dx-web/src/features/web/docs/components/docs-sidebar.tsx`

`docs-sidebar-drawer.tsx` composes `<DocsSidebar />` internally — no separate change needed for mobile.

- [ ] **Step 1: Edit the sidebar render**

Replace the topic loop inside each category (the part that renders `<Link key={topic.slug}…/>` for each topic) so it emits a divider heading when `topic.groupLabel` is set:

```tsx
// dx-web/src/features/web/docs/components/docs-sidebar.tsx
"use client";

import { Fragment } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { DOC_CATEGORIES } from "@/features/web/docs/registry";

export function DocsSidebar() {
  const pathname = usePathname();

  return (
    <div className="flex flex-col gap-1">
      <Link
        href="/docs"
        className="mb-3 text-sm font-extrabold tracking-tight text-slate-900 hover:text-teal-600"
      >
        斗学文档
      </Link>
      {DOC_CATEGORIES.map((category, ci) => {
        const catHref = `/docs/${category.slug}`;
        const isActiveCat = pathname?.startsWith(catHref) ?? false;
        const CatIcon = category.icon;
        return (
          <div key={category.slug}>
            {ci > 0 && <div className="my-2 h-px w-full bg-slate-200" />}
            <Link
              href={catHref}
              className={`flex items-center gap-2 rounded-md px-3 py-1 text-sm font-semibold ${
                isActiveCat ? "text-teal-600" : "text-slate-900"
              }`}
            >
              <CatIcon
                className={`h-4 w-4 ${
                  isActiveCat ? "text-teal-600" : "text-slate-400"
                }`}
                aria-hidden="true"
              />
              {category.title}
            </Link>
            <div className="mt-1 flex flex-col gap-1">
              {category.topics.map((topic) => {
                const topicHref = `/docs/${category.slug}/${topic.slug}`;
                const isActiveTopic = pathname === topicHref;
                return (
                  <Fragment key={topic.slug}>
                    {topic.groupLabel && (
                      <div className="mt-3 mb-1 px-3 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                        {topic.groupLabel}
                      </div>
                    )}
                    <Link
                      href={topicHref}
                      className={`flex items-center gap-2 rounded-md px-3 py-1 text-left text-sm ${
                        isActiveTopic
                          ? "bg-teal-50 font-medium text-teal-600"
                          : "text-slate-500 hover:text-slate-700"
                      }`}
                    >
                      <div
                        className={`h-5 w-[3px] rounded-sm ${
                          isActiveTopic ? "bg-teal-600" : "bg-transparent"
                        }`}
                        aria-hidden="true"
                      />
                      {topic.title}
                    </Link>
                  </Fragment>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass. No `groupLabel` usage anywhere yet, so output is visually unchanged.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/docs/components/docs-sidebar.tsx
git commit -m "feat(web): render groupLabel divider in docs sidebar"
```

---

## Task 3: Render `groupLabel` section break in category index

**Files:**
- Modify: `dx-web/src/features/web/docs/components/docs-category-index.tsx`

- [ ] **Step 1: Edit the index grid**

Wrap the topic card render in a `Fragment` so a `groupLabel` can insert a labeled divider between cards:

```tsx
// dx-web/src/features/web/docs/components/docs-category-index.tsx
import { Fragment } from "react";
import Link from "next/link";
import { ChevronRight } from "lucide-react";
import { DocsBreadcrumb } from "./docs-breadcrumb";
import type { DocCategory } from "@/features/web/docs/types";

type Props = { category: DocCategory };

export function DocsCategoryIndex({ category }: Props) {
  const Icon = category.icon;
  return (
    <>
      <DocsBreadcrumb category={category} />
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-3">
          <div
            className={`flex h-12 w-12 items-center justify-center rounded-lg border ${category.accentClass}`}
          >
            <Icon className="h-6 w-6" aria-hidden="true" />
          </div>
          <h1 className="text-2xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            {category.title}
          </h1>
        </div>
        <p className="text-base leading-relaxed text-slate-500">
          {category.description}
        </p>
      </div>
      <div className="h-px w-full bg-slate-200" />
      <div className="flex flex-col gap-3">
        {category.topics.map((topic, i) => (
          <Fragment key={topic.slug}>
            {topic.groupLabel && (
              <div className="mt-4 flex items-center gap-3">
                <div className="h-px flex-1 bg-slate-200" />
                <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                  {topic.groupLabel}
                </span>
                <div className="h-px flex-1 bg-slate-200" />
              </div>
            )}
            <Link
              href={`/docs/${category.slug}/${topic.slug}`}
              className="group flex items-center gap-4 rounded-[10px] border border-slate-200 bg-white p-5 hover:border-slate-300"
            >
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-100 text-sm font-bold text-slate-600 group-hover:bg-teal-50 group-hover:text-teal-600">
                {i + 1}
              </div>
              <div className="flex flex-1 flex-col gap-1">
                <span className="text-[15px] font-semibold text-slate-900">
                  {topic.title}
                </span>
                <span className="text-[13px] text-slate-500">
                  {topic.description}
                </span>
              </div>
              <ChevronRight
                className="h-4 w-4 shrink-0 text-slate-400 transition-transform group-hover:translate-x-1"
                aria-hidden="true"
              />
            </Link>
          </Fragment>
        ))}
      </div>
    </>
  );
}
```

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/docs/components/docs-category-index.tsx
git commit -m "feat(web): render groupLabel divider in docs category index"
```

---

## Task 4: Create legal module scaffold — types + constants

**Files:**
- Create: `dx-web/src/features/com/legal/types.ts`
- Create: `dx-web/src/features/com/legal/constants.ts`

- [ ] **Step 1: Create `types.ts`**

```ts
// dx-web/src/features/com/legal/types.ts
import type { ComponentType } from "react";

export type LegalAgreementSlug =
  | "user-agreement"
  | "privacy-policy"
  | "guardian-consent"
  | "product-service"
  | "cookie-policy";

export type LegalAgreement = {
  slug: LegalAgreementSlug;
  title: string;
  shortTitle: string;
  description: string;
  effectiveDate: string;
  lastUpdated: string;
  Component: ComponentType;
};
```

- [ ] **Step 2: Create `constants.ts`**

```ts
// dx-web/src/features/com/legal/constants.ts
export const BRAND = "斗学";
export const DOMAIN = "douxue.fun";

export const PLACEHOLDERS = {
  companyName: "{{公司名称}}",
  companyAddr: "{{公司注册地址}}",
  uscc: "{{统一社会信用代码}}",
  icpNumber: "{{ICP 备案号}}",
  pscRecordNo: "{{公安备案号}}",
  courtLocation: "{{管辖法院所在地}}",
  dataStorage: "{{数据存储地}}",
  supportEmail: "{{客服邮箱}}",
} as const;

export const EFFECTIVE_DATE = "2026-04-17";
export const LAST_UPDATED = "2026-04-17";
```

- [ ] **Step 3: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
```

Expected: pass.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/com/legal/types.ts dx-web/src/features/com/legal/constants.ts
git commit -m "feat(web): scaffold legal module types and constants"
```

---

## Task 5: Create `LegalPlaceholderNotice` component

**Files:**
- Create: `dx-web/src/features/com/legal/components/legal-placeholder-notice.tsx`

- [ ] **Step 1: Write the component**

Reuses existing `DocCallout` primitive. Lists placeholder names visibly so reviewers know exactly what must be replaced.

```tsx
// dx-web/src/features/com/legal/components/legal-placeholder-notice.tsx
import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { PLACEHOLDERS } from "@/features/com/legal/constants";

type Props = { fields: (keyof typeof PLACEHOLDERS)[] };

export function LegalPlaceholderNotice({ fields }: Props) {
  return (
    <DocCallout variant="tip" title="法律信息占位">
      <p>
        本页以下 <code>{"{{...}}"}</code> 标记的字段需在法律团队审阅后替换为正式内容，当前版本用于产品功能验证，<strong>非最终法律文本</strong>。
      </p>
      <ul className="mt-2 flex flex-wrap gap-2">
        {fields.map((k) => (
          <li
            key={k}
            className="rounded bg-amber-50 px-2 py-0.5 font-mono text-[12px] text-amber-700"
          >
            {PLACEHOLDERS[k]}
          </li>
        ))}
      </ul>
    </DocCallout>
  );
}
```

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/components/legal-placeholder-notice.tsx
git commit -m "feat(web): add LegalPlaceholderNotice component"
```

---

## Task 6: Create `AgreementDialog` component

**Files:**
- Create: `dx-web/src/features/com/legal/components/agreement-dialog.tsx`

Built on shadcn `Dialog`. Sticky header + scrollable body + sticky footer. Reads the agreement entry from the registry by slug.

- [ ] **Step 1: Write the component**

```tsx
// dx-web/src/features/com/legal/components/agreement-dialog.tsx
"use client";

import Link from "next/link";
import type { ReactNode } from "react";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import { getAgreementBySlug } from "@/features/com/legal/registry";
import type { LegalAgreementSlug } from "@/features/com/legal/types";

type Props = {
  slug: LegalAgreementSlug;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  trigger?: ReactNode;
};

export function AgreementDialog({ slug, open, onOpenChange }: Props) {
  const agreement = getAgreementBySlug(slug);
  const Body = agreement.Component;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="flex h-[85vh] max-h-[85vh] w-full max-w-[720px] flex-col gap-0 overflow-hidden p-0 sm:max-w-[720px]"
        showCloseButton
      >
        <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-6 py-4">
          <div className="flex flex-col gap-1">
            <DialogTitle className="text-lg font-bold text-slate-900">
              {agreement.title}
            </DialogTitle>
            <DialogDescription className="text-xs text-slate-500">
              生效日期：{agreement.effectiveDate} · 最近更新：{agreement.lastUpdated}
            </DialogDescription>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto px-6 py-5">
          <div className="flex flex-col gap-6 text-[14px] leading-[1.75] text-slate-700">
            <Body />
          </div>
        </div>
        <div className="flex items-center justify-between gap-3 border-t border-slate-200 bg-slate-50 px-6 py-3">
          <Link
            href={`/docs/account/${agreement.slug}`}
            target="_blank"
            rel="noreferrer"
            className="text-sm font-medium text-teal-600 hover:text-teal-700"
          >
            在完整页面查看 →
          </Link>
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="h-9 rounded-md bg-teal-600 px-4 text-sm font-semibold text-white hover:bg-teal-700"
          >
            我已阅读
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Verify**

At this step, `getAgreementBySlug` does not exist yet — `tsc` will error. **That's expected — the registry ships in Task 12.** Skip tsc verification here; run lint only:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/components/agreement-dialog.tsx
```

Expected: lint passes for this file (import resolution is not a lint concern). The `getAgreementBySlug` missing-import error will resolve after Task 12.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/components/agreement-dialog.tsx
git commit -m "feat(web): add AgreementDialog component"
```

---

## Task 7: Create `AgreementLink` component

**Files:**
- Create: `dx-web/src/features/com/legal/components/agreement-link.tsx`

- [ ] **Step 1: Write the component**

```tsx
// dx-web/src/features/com/legal/components/agreement-link.tsx
"use client";

import { useState } from "react";
import { AgreementDialog } from "./agreement-dialog";
import { getAgreementBySlug } from "@/features/com/legal/registry";
import type { LegalAgreementSlug } from "@/features/com/legal/types";

type Props = {
  slug: LegalAgreementSlug;
  className?: string;
};

export function AgreementLink({ slug, className }: Props) {
  const [open, setOpen] = useState(false);
  const agreement = getAgreementBySlug(slug);
  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className={
          className ??
          "text-teal-600 underline-offset-2 hover:text-teal-700 hover:underline"
        }
      >
        {agreement.shortTitle}
      </button>
      <AgreementDialog slug={slug} open={open} onOpenChange={setOpen} />
    </>
  );
}
```

- [ ] **Step 2: Verify**

Same note as Task 6 — `getAgreementBySlug` not yet defined. Skip tsc; run lint:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/components/agreement-link.tsx
```

Expected: pass.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/components/agreement-link.tsx
git commit -m "feat(web): add AgreementLink component"
```

---

## Task 8: Create `AgreementInlineList` component

**Files:**
- Create: `dx-web/src/features/com/legal/components/agreement-inline-list.tsx`

- [ ] **Step 1: Write the component**

```tsx
// dx-web/src/features/com/legal/components/agreement-inline-list.tsx
import { Fragment } from "react";
import { AgreementLink } from "./agreement-link";
import type { LegalAgreementSlug } from "@/features/com/legal/types";

type Props = {
  prefix: string;
  slugs: LegalAgreementSlug[];
  className?: string;
};

export function AgreementInlineList({ prefix, slugs, className }: Props) {
  return (
    <p className={className ?? "text-xs text-slate-400"}>
      {prefix}
      {slugs.map((slug, i) => (
        <Fragment key={slug}>
          <AgreementLink slug={slug} className="text-slate-500 hover:text-slate-700" />
          {i < slugs.length - 1 && "、"}
        </Fragment>
      ))}
    </p>
  );
}
```

- [ ] **Step 2: Verify (lint only for now)**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/components/agreement-inline-list.tsx
```

Expected: pass.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/components/agreement-inline-list.tsx
git commit -m "feat(web): add AgreementInlineList component"
```

---

## Task 9: Write 用户协议 document

**Files:**
- Create: `dx-web/src/features/com/legal/documents/user-agreement.tsx`

The document is a React component that renders the full legal text using existing doc primitives (`DocSection`, `DocCallout`, `LegalPlaceholderNotice`). Structure is fixed by the spec; prose is China-law-compliant, 斗学-specific, vendor-agnostic on payments, with guardian clauses.

### Required structure (non-negotiable)

Top matter:
1. `<LegalPlaceholderNotice fields={["companyName", "companyAddr", "courtLocation", "supportEmail"]} />`
2. Metadata line: `生效日期：2026-04-17 · 最近更新：2026-04-17`
3. Preamble: 斗学 positioning (游戏化英语学习 / 社区社交 / AI 辅助); 法律法规引用 (《网络安全法》《个人信息保护法》《未成年人网络保护条例》《网络信息内容生态治理规定》); minor must have guardian consent.

Sections (in order, each as `<DocSection id="..." title="...">`):

| id | title |
|---|---|
| `clause-1` | 第1条 定义 |
| `clause-2` | 第2条 用户注册与账号管理 |
| `clause-3` | 第3条 用户的权利与义务 |
| `clause-4` | 第4条 本平台的权利与义务 |
| `clause-5` | 第5条 知识产权 |
| `clause-6` | 第6条 免责声明 |
| `clause-7` | 第7条 协议的修改与终止 |
| `clause-8` | 第8条 个人信息保护专项条款 |
| `clause-9` | 第9条 账号注销 |
| `clause-10` | 第10条 法律适用与争议解决 |
| `clause-11` | 第11条 其他 |

### Required clauses verbatim (legal-critical; must appear exactly as written)

**In 第2条:**
- `<DocCallout variant="warning" title="账号禁止共享使用">` body: "用户的账号仅限用户本人独立使用，严禁将账号提供给他人使用，严禁多人共享同一账号。本平台实施账号同一时间仅限单设备登录的技术限制措施，若检测到异常登录行为，本平台有权要求身份验证、临时冻结账号，或对经核实确认存在账号共享、倒卖会员权益等违规行为的账号立即暂停或永久封禁，已支付费用不予退还。"
- 未成年人年龄分层：8 周岁以下不得注册；8 周岁以上未成年人需监护人同意；16 周岁以上不满 18 周岁仍需监护人同意方可使用付费功能。

**In 第5条:**
- UGC 授权条款：全球范围内、非独占、免费、可转授权的许可。

**In 第10条:**
- "本协议适用中华人民共和国法律。因本协议产生的争议，双方应首先友好协商解决；协商不成的，任何一方有权向本平台运营方所在地 `{{管辖法院所在地}}` 有管辖权的人民法院提起诉讼。"

### Product-reality guardrails (DO NOT violate)

- Brand: use `BRAND` constant or literally `斗学`; domain `douxue.fun`.
- Virtual items: 能量豆 (not 钻石).
- Membership tiers: generic (月度/季度/年度/终身会员) — NO prices.
- Payment channels: "平台支持的付费渠道" — NO WeChat/Alipay vendor mentions.
- NO auto-renewal mentions.
- NO 增值税 / 发票 mentions.
- NO 微信服务号 mentions as a channel; only `{{客服邮箱}}` for formal contact.
- Referral: single sentence pointer — "推广返利的具体规则以平台《推广规则》公示为准". No rates, no tax clauses.
- Data storage location: `{{数据存储地}}` placeholder.
- Operating entity: `{{公司名称}}` placeholder throughout; first occurrence may say "本平台运营方（`{{公司名称}}`）".

### Reference usage (styling)

- Wrap 免责/限制责任 clauses in `<strong>`.
- Use `<ol>` / `<ul>` for enumerations.
- Cross-link sibling docs inline: `<AgreementLink slug="privacy-policy" />`.

- [ ] **Step 1: Create the file with full content**

Write `dx-web/src/features/com/legal/documents/user-agreement.tsx`. The scaffold:

```tsx
// dx-web/src/features/com/legal/documents/user-agreement.tsx
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { LegalPlaceholderNotice } from "@/features/com/legal/components/legal-placeholder-notice";
import { AgreementLink } from "@/features/com/legal/components/agreement-link";
import {
  BRAND,
  DOMAIN,
  EFFECTIVE_DATE,
  LAST_UPDATED,
  PLACEHOLDERS,
} from "@/features/com/legal/constants";

function P({ children }: { children: React.ReactNode }) {
  return <span className="rounded bg-amber-50 px-1 py-0.5 font-mono text-[13px] text-amber-700">{children}</span>;
}

export function UserAgreementDoc() {
  return (
    <>
      <LegalPlaceholderNotice
        fields={["companyName", "companyAddr", "courtLocation", "supportEmail"]}
      />
      <p className="text-xs text-slate-500">
        生效日期：{EFFECTIVE_DATE} · 最近更新：{LAST_UPDATED}
      </p>

      {/* Preamble */}
      <p>
        欢迎您注册并使用{BRAND}（域名：{DOMAIN}）提供的产品和服务。{BRAND}系专注于英语学习的在线服务平台，通过游戏化学习机制（任务、等级、排行榜、虚拟道具等）、社区功能（发帖、评论、小组、关注）及 AI 辅助能力，为用户提供有趣、高效的英语学习体验。{BRAND}运营主体为 <P>{PLACEHOLDERS.companyName}</P>（以下简称 &ldquo;本平台&rdquo; 或 &ldquo;运营方&rdquo;）。
      </p>
      <p>
        您通过点击 &ldquo;同意&rdquo; 按钮、注册账号或使用本平台服务，即视为您已充分阅读、理解并接受本协议全部条款（
        <strong>特别是加粗标注的免除 / 限制本平台责任、加重您义务、涉及您重大权益的条款</strong>
        ），并同意受本协议约束。
      </p>
      <p>
        若您为未成年人（不满 18 周岁），需在法定监护人陪同下阅读本协议，并取得监护人明确同意（勾选《监护人同意书》）后，方可注册或使用本平台服务。
      </p>
      <p>
        本协议依据《中华人民共和国网络安全法》《中华人民共和国个人信息保护法》《中华人民共和国未成年人网络保护条例》《网络信息内容生态治理规定》等法律法规制定。若您不同意本协议任何条款，请勿注册或使用本平台服务。
      </p>

      {/* === 后续 11 条内容在执行阶段按规定结构和必填条款补齐 === */}
    </>
  );
}
```

**Now expand the body.** Write all 11 sections with full content following the required structure, the required verbatim clauses (in Task-9 preamble text above), and the product-reality guardrails. Minimum 2500 Chinese characters total across sections. Each `DocSection` body uses `<p>`, `<ol>`, `<ul>`, `<strong>`, `<DocCallout>` as needed.

Key clauses to implement inside sections:

- **clause-1 (定义):** 用户 / 本平台 / 服务 / 虚拟物品（能量豆/积分/等级）/ 游戏化作弊行为 / 付费服务 / 小组功能 / UGC 内容。
- **clause-2 (注册与账号):** 真实信息 · 微信第三方账户 · 账号保管 · 禁止共享 warning callout (verbatim text from structure block above) · 单设备登录 · 未成年人 8/16/18 分层。
- **clause-3 (用户权利与义务):** 使用服务 / 反馈 / 注销权；内容合规义务 8 项违禁；小组内容独立责任 / 组长管理义务 / 禁止拉人头导流。
- **clause-4 (平台权利与义务):** 服务调整权 (提前 1 日通告) · 违规内容处置 · 个人信息保护义务 · 技术保障义务 (不保证完全无故障)。
- **clause-5 (知识产权):** UGC 归属用户 + 全球非独占免费可转授权；平台知识产权 (商标、域名、课程、算法、游戏化机制) 禁止反编译盗用；侵权通知后 3 个工作日应对期。
- **clause-6 (免责):** 学习效果免责 · 不可抗力免责 · 虚拟物品免责 · 小组内容免责 (wrap in `<strong>` for 免责 sentences)。
- **clause-7 (修改终止):** 修改公告 3 日生效；违规终止不退款；续费 / 自动续费话题**一律不提**。
- **clause-8 (个人信息保护专项):** 跨引用：详见 <AgreementLink slug="privacy-policy" />；收集范围最小化；未成年人信息不用于营销；撤回同意路径 = 个人主页-隐私设置。
- **clause-9 (账号注销):** 正常状态 · 无未到期权益 · 路径：个人主页→账号安全→注销；或客服协助（引用 <P>{PLACEHOLDERS.supportEmail}</P>）。
- **clause-10 (法律适用与争议):** Verbatim clause from structure block above.
- **clause-11 (其他):** 条款独立性 · 通知送达方式 (公告 / 邮箱 / 短信) · 本协议最终解释权归 <P>{PLACEHOLDERS.companyName}</P>。

- [ ] **Step 2: Verify lint on new file**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/documents/user-agreement.tsx
```

Expected: pass. Common lint failure mode: `react/no-unescaped-entities` — use `&ldquo;`, `&rdquo;`, `&amp;`, `&lt;`, `&gt;`, `{"{"}`, `{"}"}` as needed.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/documents/user-agreement.tsx
git commit -m "feat(web): write 用户协议 document"
```

---

## Task 10: Write 隐私政策 document

**Files:**
- Create: `dx-web/src/features/com/legal/documents/privacy-policy.tsx`

### Required structure

Top matter: `<LegalPlaceholderNotice fields={["companyName", "dataStorage", "supportEmail"]} />` + metadata line + preamble.

Preamble must include: 《个人信息保护法》《未成年人网络保护条例》法律法规引用; 合法/正当/必要/诚信原则; minor requires guardian consent.

Sections:

| id | title |
|---|---|
| `clause-1` | 第1条 定义 |
| `clause-2` | 第2条 个人信息的收集与范围 |
| `clause-3` | 第3条 个人信息的使用规则 |
| `clause-4` | 第4条 个人信息的共享、转让与披露 |
| `clause-5` | 第5条 个人信息的存储与安全保护 |
| `clause-6` | 第6条 用户的个人信息权利及行使方式 |
| `clause-7` | 第7条 未成年人个人信息的特别保护 |
| `clause-8` | 第8条 隐私政策的修改与通知 |
| `clause-9` | 第9条 联系我们 |
| `clause-10` | 第10条 其他 |

### Required verbatim clauses

**In clause-2:** A table of 信息类别 / 具体内容 / 收集目的. Render via an HTML `<table>` (or `<DocKeyValue>` primitive — check `dx-web/src/features/web/docs/primitives/doc-key-value.tsx` for signature). Categories (at minimum):
1. 账号注册与身份验证信息 — 用户名、邮箱、微信 openid/unionid、密码（加密存储）；监护人联系方式 (仅未成年人场景)。
2. 学习相关信息 — 学习时长、课程进度、任务打卡、生词/掌握记录、UGC 内容、排行榜数据。
3. 第三方授权信息 — 微信快捷登录时的公开信息，不收集密码和好友列表。
4. 支付相关信息 — 订单号、金额、支付渠道名称；NOT 银行卡号、支付密码。
5. 设备与访问信息 — 设备型号、OS 版本、浏览器类型、IP、登录时间、访问路径。

**In clause-3:** A subsection 3.1.4 titled `发送通知（含短信通知）`. Must state: 必选通知 (验证码、账号安全、服务变更) + 可选通知 (学习提醒、活动优惠). Verbatim: "<strong>您注册并使用本平台服务，即表示您已阅读、理解并同意我们按上述方式向您发送通知。</strong>" 可选通知可在 个人主页-隐私设置 关闭。

**In clause-3.3:** 禁止性使用 — 不做 "大数据杀熟"；不向学习记录无关的商品广告；不将未成年人信息用于商业用途。

**In clause-4:** Three subsections (共享 / 转让 / 披露). 转让需提前 7 日通告；披露仅限法定机关文书或维护用户/公共利益。

**In clause-5:**
- 5.1.1 存储地点：用户个人信息存储于中华人民共和国境内的服务器（机房位于 <P>{PLACEHOLDERS.dataStorage}</P>），未跨境存储。
- 5.1.2 存储期限：账号存续期间存储；注销后 30 个工作日内删除或匿名化；法定保留期限例外 (支付记录 5 年)。

**In clause-6:** A table of 权利类型 / 具体内容 / 行使路径 / 响应时间. Rights: 查询 / 更正 / 删除 / 撤回同意 / 注销。

**In clause-7:** 未成年人额外保护 — 8 周岁以下不得注册；不向未成年人推送商业广告；不收集生物识别信息和行踪轨迹；向监护人提供查询/删除/异议权。

**In clause-9:** 仅使用 <P>{PLACEHOLDERS.supportEmail}</P> 作为联系渠道；10 个工作日内响应；若对处理结果不满可向国家互联网信息办公室投诉。

### Product-reality guardrails (same as Task 9)

Virtual items = 能量豆. No 微信服务号. No specific payment vendor. No 微信公众号 retrieval path.

- [ ] **Step 1: Create the file**

Same scaffold pattern as Task 9 (imports, `P` helper for placeholders, top matter, then sections). Write all 10 sections with full content, minimum 3000 Chinese characters.

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/documents/privacy-policy.tsx
```

Expected: pass.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/documents/privacy-policy.tsx
git commit -m "feat(web): write 隐私政策 document"
```

---

## Task 11: Write 监护人同意书 document

**Files:**
- Create: `dx-web/src/features/com/legal/documents/guardian-consent.tsx`

### Required structure

Top matter: `<LegalPlaceholderNotice fields={["companyName", "supportEmail"]} />` + metadata line + addressing paragraph ("尊敬的家长 / 监护人：").

Preamble must invoke 《未成年人保护法》《个人信息保护法》.

Sections:

| id | title |
|---|---|
| `clause-1` | 第1条 服务说明 |
| `clause-2` | 第2条 未成年人信息收集与使用 |
| `clause-3` | 第3条 监护人的权利 |
| `clause-4` | 第4条 监护人的责任 |
| `clause-5` | 第5条 内容安全保障 |
| `clause-6` | 第6条 使用时长管理 |
| `clause-7` | 第7条 付费服务说明 |
| `clause-8` | 第8条 信息安全保护 |
| `clause-9` | 第9条 同意书的变更 |
| `clause-10` | 第10条 联系我们 |
| `clause-11` | 第11条 监护人声明 |

### Required verbatim clauses

**In clause-2.3:** 不会将未成年人信息用于商业营销；未经监护人同意不向第三方提供；最小化收集原则。

**In clause-6:** 合理安排学习时间；每 45-60 分钟休息 10-15 分钟。

**In clause-7:** 付费服务包括 **会员服务和能量豆充值**（斗学现实 — 不要写 "高级课程"/"辅导服务"）；未成年人购买付费服务需经监护人同意并由监护人完成支付；建议定期检查账户消费记录。

**In clause-11:** 勾选即确认：(1) 您是合法监护人；(2) 已阅读全部内容；(3) 同意孩子注册使用 斗学 (`douxue.fun`)；(4) 同意按本同意书收集使用保护孩子信息；(5) 履行监督指导责任。

### Guardrails

- Contact channel: <P>{PLACEHOLDERS.supportEmail}</P> only.
- Brand: 斗学, 域名 `douxue.fun`.
- Virtual items: 能量豆.
- NO WeChat vendor names.

- [ ] **Step 1: Create the file**

Same scaffold. Minimum 1500 Chinese characters.

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/documents/guardian-consent.tsx
```

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/documents/guardian-consent.tsx
git commit -m "feat(web): write 监护人同意书 document"
```

---

## Task 12: Write 产品服务协议 document

**Files:**
- Create: `dx-web/src/features/com/legal/documents/product-service.tsx`

### Required structure

Top matter: `<LegalPlaceholderNotice fields={["companyName", "companyAddr", "courtLocation", "supportEmail"]} />` + metadata line + preamble.

Preamble cites 《消费者权益保护法》第二十五条 (数字商品) alongside the usual laws.

Sections:

| id | title |
|---|---|
| `clause-1` | 第1条 定义 |
| `clause-2` | 第2条 协议生效与效力 |
| `clause-3` | 第3条 会员服务内容与权益 |
| `clause-4` | 第4条 会员订阅与支付规则 |
| `clause-5` | 第5条 会员账号使用规范 |
| `clause-6` | 第6条 会员服务的暂停与终止 |
| `clause-7` | 第7条 退订与退款规则 |
| `clause-8` | 第8条 知识产权 |
| `clause-9` | 第9条 免责声明 |
| `clause-10` | 第10条 账号注销 |
| `clause-11` | 第11条 协议的修改与通知 |
| `clause-12` | 第12条 联系我们 |
| `clause-13` | 第13条 法律适用与争议解决 |
| `clause-14` | 第14条 其他 |

### Required verbatim clauses

**In clause-2.1 (三种生效触发):**
1. 点击 "同意协议并支付" 按钮；
2. **付费行为：您完成会员费用支付，无论是否点击协议确认按钮，付费行为本身即视为您已充分阅读、理解并完全接受本协议全部条款**（wrapped in `<strong>`）；
3. 实际使用会员服务即视为认可本协议效力。

**In clause-3 (会员权益):** 五档会员泛称（月度/季度/年度/终身会员等）。权益包括：
- 解锁所有标注 "会员专享" 的关卡和内容；
- 专属客服通道；
- 能量豆月度赠送（generic amount — 不写具体数字）；
- 排行榜专属标识；
- 会员专属学习社群。
- 推广返利：一句指向 — "推广返利相关规则以平台《推广规则》实时公示为准"。NO 30%, NO 90 天, NO 劳务报酬/代扣。

**In clause-4:** 订阅流程四步：注册/绑定 → 选择档位 → 确认订单 → 完成支付。费用为含税价格。付费渠道使用 "平台支持的付费渠道"（NO WeChat/Alipay by name）。**NO 自动续费 subsection。**

**In clause-5 (账号规范) — ALL in `<DocCallout variant="warning">`:**
- 账号归属：会员账号归本平台所有，用户享有使用权，禁止转借/出租/出售/共享；违反立即封禁且不退费；
- 单设备登录：异常登录本平台有权冻结；
- 未成年人：8 周岁以下不得订阅；8-16 需监护人协助并勾选 <AgreementLink slug="guardian-consent" />；16-18 需监护人同意后订阅。

**In clause-7 (退款) — critical legal clauses, use `<strong>`:**
- 7.2 **会员服务属于《消费者权益保护法》第二十五条规定的 "在线下载的数字化商品"，具有无形性、即时交付性、不可回收性等特点，除本协议明确约定外，订阅后不支持无理由退款。**
- 7.3 例外可退款：
  - 年度/终身会员 10 个自然日内无理由全额退款（不适用于月度、季度、能量豆充值）；
  - 未成年人误订阅（监护人 7 日内提交证明）；
  - 系统故障重复扣费。
- 7.4 不可退款：月度/季度会员一经购买不退；年度/终身超过 10 天不退；自身原因不退；违规封禁不退。
- 7.5 退款流程：通过 <P>{PLACEHOLDERS.supportEmail}</P> 提交；10 个工作日内反馈；原路退回。

**In clause-13 (法律适用):** verbatim — "本协议适用中华人民共和国法律（不含香港、澳门、台湾）。因本协议产生的争议，双方应首先友好协商解决；协商不成的，任何一方有权向本平台运营方所在地 <P>{PLACEHOLDERS.courtLocation}</P> 有管辖权的人民法院提起诉讼。"

### Guardrails

Same as all prior docs. **Absolutely no auto-renewal, no vendor-specific payment, no 30%/tax/劳务报酬 referral clauses, no 发票/增值税, no 微信服务号.**

- [ ] **Step 1: Create the file**

Minimum 3000 Chinese characters. Same scaffold pattern.

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/documents/product-service.tsx
```

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/documents/product-service.tsx
git commit -m "feat(web): write 产品服务协议 document"
```

---

## Task 13: Write Cookie 政策 document

**Files:**
- Create: `dx-web/src/features/com/legal/documents/cookie-policy.tsx`

### Required structure

Top matter: `<LegalPlaceholderNotice fields={["companyName", "supportEmail"]} />` + metadata line + preamble.

Sections:

| id | title |
|---|---|
| `clause-1` | 第1条 什么是 Cookie 及类似技术 |
| `clause-2` | 第2条 我们如何使用 Cookie |
| `clause-3` | 第3条 Cookie 的分类 |
| `clause-4` | 第4条 第三方 Cookie 与 SDK |
| `clause-5` | 第5条 您的选择与管理方式 |
| `clause-6` | 第6条 本政策的更新 |
| `clause-7` | 第7条 联系我们 |

### Required content

**clause-1:** 解释 Cookie / LocalStorage / SessionStorage 的基本概念和用途。

**clause-2 (使用目的):** 分类说明：
1. 必要性 Cookie — 登录状态（`dx_token` 等）、CSRF 防护、会话保持；
2. 功能性 Cookie — 用户偏好（如侧边栏折叠、主题设置）；
3. 分析性 Cookie — 学习行为、页面访问路径，用于产品优化。

**clause-3 (分类 table):** Render an HTML `<table>` with columns: 类别 / 示例 / 目的 / 存留时长 / 是否可关闭。Include at least:
- `dx_token` — 登录凭证 — Session / 7 days — 必要（不可关闭）
- 功能偏好类 — UI 状态 — 30 days — 可选（可关闭）
- 分析类 — 访问路径 — 30 days — 可选（可关闭）

**clause-4:** 说明如使用第三方服务（如微信开放平台 OAuth），会由这些服务方在您的浏览器中设置其 Cookie，受其隐私政策约束；我们不主动读取跨域 Cookie。明确 NOT 当前使用广告联盟 / 像素追踪类 Cookie。

**clause-5:** 用户可在浏览器设置中禁用 Cookie，但会导致登录、学习进度同步等核心功能不可用。提供主流浏览器（Chrome / Safari / Edge / Firefox）Cookie 管理路径的泛指引（不需要具体点击步骤 — 一句话指引即可）。

**clause-6:** 政策更新通过 <AgreementLink slug="privacy-policy" /> 同步公告；继续使用视为接受。

**clause-7:** <P>{PLACEHOLDERS.supportEmail}</P>。

### Guardrails

No 微信服务号. No specific analytics vendor (如 Google Analytics / 百度统计) unless the user confirms use. Keep vendor-neutral.

- [ ] **Step 1: Create the file**

Minimum 1500 Chinese characters. Same scaffold pattern.

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint -- --max-warnings 0 src/features/com/legal/documents/cookie-policy.tsx
```

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/documents/cookie-policy.tsx
git commit -m "feat(web): write Cookie 政策 document"
```

---

## Task 14: Populate `LEGAL_AGREEMENTS` registry

**Files:**
- Create: `dx-web/src/features/com/legal/registry.ts`

- [ ] **Step 1: Write the registry**

```ts
// dx-web/src/features/com/legal/registry.ts
import { UserAgreementDoc } from "./documents/user-agreement";
import { PrivacyPolicyDoc } from "./documents/privacy-policy";
import { GuardianConsentDoc } from "./documents/guardian-consent";
import { ProductServiceDoc } from "./documents/product-service";
import { CookiePolicyDoc } from "./documents/cookie-policy";
import { EFFECTIVE_DATE, LAST_UPDATED } from "./constants";
import type { LegalAgreement, LegalAgreementSlug } from "./types";

export {
  UserAgreementDoc,
  PrivacyPolicyDoc,
  GuardianConsentDoc,
  ProductServiceDoc,
  CookiePolicyDoc,
};

export const LEGAL_AGREEMENTS: LegalAgreement[] = [
  {
    slug: "user-agreement",
    title: "用户协议",
    shortTitle: "《用户协议》",
    description:
      "斗学账号注册、账号管理、用户权责、知识产权、免责与注销等完整条款。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: UserAgreementDoc,
  },
  {
    slug: "privacy-policy",
    title: "隐私政策",
    shortTitle: "《隐私政策》",
    description:
      "我们如何收集、使用、存储、共享、保护您的个人信息，以及您的权利行使方式。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: PrivacyPolicyDoc,
  },
  {
    slug: "guardian-consent",
    title: "监护人同意书",
    shortTitle: "《监护人同意书》",
    description:
      "未成年人使用斗学前，监护人需知情并同意的相关条款与权责说明。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: GuardianConsentDoc,
  },
  {
    slug: "product-service",
    title: "产品服务协议",
    shortTitle: "《产品服务协议》",
    description:
      "会员订阅与支付、服务暂停与终止、退款规则、知识产权及争议解决等条款。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: ProductServiceDoc,
  },
  {
    slug: "cookie-policy",
    title: "Cookie 政策",
    shortTitle: "《Cookie 政策》",
    description:
      "我们在斗学网站使用的 Cookie 及类似技术的类型、目的与您的选择。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: CookiePolicyDoc,
  },
];

export function getAgreementBySlug(slug: LegalAgreementSlug): LegalAgreement {
  const found = LEGAL_AGREEMENTS.find((a) => a.slug === slug);
  if (!found) {
    throw new Error(`[legal] unknown agreement slug: ${slug}`);
  }
  return found;
}
```

- [ ] **Step 2: Verify tsc + lint now run clean across the full module**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: both pass. At this point `AgreementDialog` and `AgreementLink` have their missing imports resolved.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/com/legal/registry.ts
git commit -m "feat(web): populate legal agreements registry"
```

---

## Task 15: Register 5 legal topics under `account` category

**Files:**
- Modify: `dx-web/src/features/web/docs/registry.ts`

- [ ] **Step 1: Add imports**

At the top of the file, after the existing topic imports (after line 63 `import Faq from "./topics/account/faq";`), add:

```ts
import {
  UserAgreementDoc,
  PrivacyPolicyDoc,
  GuardianConsentDoc,
  ProductServiceDoc,
  CookiePolicyDoc,
} from "@/features/com/legal/registry";
```

- [ ] **Step 2: Append 5 legal topics to the `account` category**

Locate the `account` category object (currently ends with the `faq` topic around line 483). Append these 5 topics **after** the `faq` entry, before the closing `]` of the `topics` array:

```ts
      {
        slug: "user-agreement",
        title: "用户协议",
        description:
          "斗学账号注册、账号管理、用户权责、知识产权、免责与注销等完整条款。",
        Component: UserAgreementDoc,
        groupLabel: "法律条款",
      },
      {
        slug: "privacy-policy",
        title: "隐私政策",
        description:
          "我们如何收集、使用、存储、共享、保护您的个人信息，以及您的权利行使方式。",
        Component: PrivacyPolicyDoc,
      },
      {
        slug: "guardian-consent",
        title: "监护人同意书",
        description:
          "未成年人使用斗学前，监护人需知情并同意的相关条款与权责说明。",
        Component: GuardianConsentDoc,
      },
      {
        slug: "product-service",
        title: "产品服务协议",
        description:
          "会员订阅与支付、服务暂停与终止、退款规则、知识产权及争议解决等条款。",
        Component: ProductServiceDoc,
      },
      {
        slug: "cookie-policy",
        title: "Cookie 政策",
        description:
          "我们在斗学网站使用的 Cookie 及类似技术的类型、目的与您的选择。",
        Component: CookiePolicyDoc,
      },
```

- [ ] **Step 3: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: pass. If `generateStaticParams` in the topic page fails, revisit — it reads `DOC_CATEGORIES`.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/docs/registry.ts
git commit -m "feat(web): register 5 legal topics under 账户与帮助 category"
```

---

## Task 16: Milestone verification — /docs routes live

**Files:** none (verification only)

- [ ] **Step 1: Build**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run build
```

Expected: succeeds. Inspect output — route list must include:
- `/docs/account/user-agreement`
- `/docs/account/privacy-policy`
- `/docs/account/guardian-consent`
- `/docs/account/product-service`
- `/docs/account/cookie-policy`

- [ ] **Step 2: Run dev server**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run dev
```

Wait for "ready" message.

- [ ] **Step 3: Browser functional checks**

In a browser at `localhost:3000`:

1. Visit `/docs/account` → verify the 10 topic cards render; a `法律条款` horizontal divider with label appears between 常见问题 (5th card) and 用户协议 (6th card).
2. Sidebar (desktop viewport ≥ 1024px): scroll 账户与帮助; the `法律条款` small-caps label appears above 用户协议.
3. Visit `/docs/account/user-agreement` → full doc renders with placeholder notice at top, amber placeholder chips inline, all 11 sections, prev = 常见问题, next = 隐私政策.
4. Tab prev/next through all 5 legal topics → continuous chain, Cookie 政策 has no next.
5. Mobile drawer (< 1024px viewport): open 目录; same 法律条款 label appears.

If any check fails, fix the underlying code before proceeding. Kill the dev server when done: `Ctrl-C`.

- [ ] **Step 4: No commit (verification-only task)**

---

## Task 17: Wire `AgreementLink` into SignUpForm

**Files:**
- Modify: `dx-web/src/features/web/auth/components/sign-up-form.tsx`

- [ ] **Step 1: Add import**

At the top of the file, after the existing imports (after the `useSignup` import line), add:

```tsx
import { AgreementLink } from "@/features/com/legal/components/agreement-link";
```

- [ ] **Step 2: Replace the 4 static spans (lines 202–208)**

Current code:
```tsx
<span className="text-xs text-slate-700">
  我已阅读并同意{" "}
  <span className="text-teal-600">用户协议</span>、
  <span className="text-teal-600">隐私政策</span>、
  <span className="text-teal-600">监护人同意书</span>、
  <span className="text-teal-600">产品服务协议</span>
</span>
```

New code:
```tsx
<span className="text-xs text-slate-700">
  我已阅读并同意{" "}
  <AgreementLink slug="user-agreement" />、
  <AgreementLink slug="privacy-policy" />、
  <AgreementLink slug="guardian-consent" />、
  <AgreementLink slug="product-service" />
</span>
```

- [ ] **Step 3: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: pass.

- [ ] **Step 4: Browser check**

Dev server running (`npm run dev`). Visit `/auth/signup`. Click 《用户协议》 — dialog opens with the full user agreement, header shows title + dates, body scrolls, footer has "在完整页面查看 →" and "我已阅读". Clicking close, Esc key, outside click, and "我已阅读" all close the dialog. Repeat click-check for 《隐私政策》, 《监护人同意书》, 《产品服务协议》. Form state (email, code, etc.) preserved across dialog open/close cycles. "我已阅读" does NOT auto-check the form's agreement checkbox.

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/auth/components/sign-up-form.tsx
git commit -m "feat(web): wire AgreementLink into signup form"
```

---

## Task 18: Wire `AgreementInlineList` into SignInForm

**Files:**
- Modify: `dx-web/src/features/web/auth/components/sign-in-form.tsx`

- [ ] **Step 1: Add import**

At the top of the file, after the existing imports (after the `useSignIn` import line), add:

```tsx
import { AgreementInlineList } from "@/features/com/legal/components/agreement-inline-list";
```

- [ ] **Step 2: Insert the hint line**

Locate the footer block that currently looks like (around line 264–280):

```tsx
          {/* Separator */}
          <div className="h-px bg-slate-200" />

          {/* Footer */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1 text-sm">
              <span className="text-slate-400">没有账号？</span>
              <Link
                href="/auth/signin"
```

Insert the inline list between the separator and the footer row. Rewrite that segment as:

```tsx
          {/* Separator */}
          <div className="h-px bg-slate-200" />

          {/* Agreement hint */}
          <AgreementInlineList
            prefix="登录即代表同意 "
            slugs={[
              "user-agreement",
              "privacy-policy",
              "guardian-consent",
              "cookie-policy",
            ]}
            className="text-xs text-slate-400"
          />

          {/* Footer */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1 text-sm">
              <span className="text-slate-400">没有账号？</span>
              <Link
                href="/auth/signin"
```

- [ ] **Step 3: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: pass.

- [ ] **Step 4: Browser check**

Visit `/auth/signin`. Below the separator and above "没有账号？" row, a muted text line reads "登录即代表同意 《用户协议》、《隐私政策》、《监护人同意书》、《Cookie 政策》". Each clickable, opens the correct dialog.

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/auth/components/sign-in-form.tsx
git commit -m "feat(web): wire AgreementInlineList into signin form"
```

---

## Task 19: Wire `AgreementLink` into OrderPayment

**Files:**
- Modify: `dx-web/src/features/web/purchase/components/order-payment.tsx`

- [ ] **Step 1: Add import**

At the top of the file, after the existing imports, add:

```tsx
import { AgreementLink } from "@/features/com/legal/components/agreement-link";
```

- [ ] **Step 2: Replace the fake agreement span (around line 163)**

Current code:
```tsx
          <div className="flex flex-col gap-1">
            <span className="text-sm text-slate-700">
              我已阅读并同意以下协议
            </span>
            <span className="text-xs text-teal-600">《斗学服务协议》</span>
          </div>
```

New code:
```tsx
          <div className="flex flex-col gap-1">
            <span className="text-sm text-slate-700">
              我已阅读并同意以下协议
            </span>
            <span className="text-xs">
              <AgreementLink slug="product-service" />
            </span>
          </div>
```

- [ ] **Step 3: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: pass.

- [ ] **Step 4: Browser check**

Create a pending order via `/purchase/membership` → pick any tier. Payment page loads. Click 《产品服务协议》 — dialog opens with full product service doc. Close by × / Esc / outside / "我已阅读". Parent checkbox state unchanged.

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/purchase/components/order-payment.tsx
git commit -m "feat(web): wire AgreementLink into payment form"
```

---

## Task 20: Wire real legal links into Footer, remove email, fix domain

**Files:**
- Modify: `dx-web/src/components/in/footer.tsx`

- [ ] **Step 1: Rewrite the file**

Replacing it wholesale is cleaner than patching individual lines:

```tsx
// dx-web/src/components/in/footer.tsx
import Link from "next/link";
import { GraduationCap } from "lucide-react";

const LEGAL_LINKS: { label: string; href: string }[] = [
  { label: "用户协议", href: "/docs/account/user-agreement" },
  { label: "隐私政策", href: "/docs/account/privacy-policy" },
  { label: "监护人同意书", href: "/docs/account/guardian-consent" },
  { label: "产品服务协议", href: "/docs/account/product-service" },
  { label: "Cookie 政策", href: "/docs/account/cookie-policy" },
];

const footerColumns = [
  {
    title: "斗学产品",
    links: [
      "渐进学习法",
      "AI 智能定制",
      "多重游戏模式",
      "丰富课程体系",
      "社群小组",
    ],
  },
  {
    title: "斗学团队",
    links: ["关于我们", "建议反馈", "内容投稿", "商务合作"],
  },
];

export function Footer() {
  return (
    <footer
      id="contact"
      className="scroll-mt-20 w-full border-t border-slate-200 bg-slate-50"
    >
      <div className="mx-auto flex max-w-[1280px] flex-col gap-12 px-[120px] pb-10 pt-[60px]">
        <div className="flex w-full flex-col gap-10 xl:flex-row xl:justify-between">
          {/* Brand */}
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2.5">
              <GraduationCap className="h-7 w-7 text-teal-600" />
              <span className="text-lg font-extrabold text-slate-900">斗学</span>
            </div>
            <p className="max-w-[280px] text-sm leading-[1.5] text-slate-500">
              玩游戏，学英语，AI 智能辅助，斗学重新定义英语学习体验，让进步自然发生...
            </p>
          </div>

          {/* Link columns */}
          <div className="grid grid-cols-1 gap-10 md:grid-cols-2 lg:grid-cols-3 xl:flex xl:gap-16">
            {/* 服务条款 — real links */}
            <div className="flex flex-col gap-4">
              <h4 className="text-[13px] font-semibold tracking-[1px] text-slate-900">
                服务条款
              </h4>
              {LEGAL_LINKS.map((l) => (
                <Link
                  key={l.href}
                  href={l.href}
                  className="text-sm text-slate-500 hover:text-slate-700"
                >
                  {l.label}
                </Link>
              ))}
            </div>

            {/* Other columns — unchanged mock spans */}
            {footerColumns.map((col) => (
              <div key={col.title} className="flex flex-col gap-4">
                <h4 className="text-[13px] font-semibold tracking-[1px] text-slate-900">
                  {col.title}
                </h4>
                {col.links.map((link) => (
                  <span
                    key={link}
                    className="cursor-pointer text-sm text-slate-500 hover:text-slate-700"
                  >
                    {link}
                  </span>
                ))}
              </div>
            ))}
          </div>

          {/* Contact column */}
          <div className="flex flex-col items-start gap-4 xl:items-end">
            <h4 className="text-[13px] font-semibold tracking-[1px] text-slate-900">
              联系我们
            </h4>
            <div className="flex h-[140px] w-[140px] items-center justify-center rounded-lg bg-slate-100">
              <span className="text-xs text-slate-400">微信二维码</span>
            </div>
            <span className="text-xs text-slate-400">微信扫一扫联系小助手</span>
          </div>
        </div>

        <div className="h-px w-full bg-slate-200" />

        <div className="flex w-full flex-col items-center gap-2">
          <span className="text-[13px] text-slate-400">
            © 2026 douxue.fun 版权所有
          </span>
          <span className="text-[13px] text-slate-400">
            京公网安备 xxxxxxxxxxxxxx 号  京 ICP 备 xxxxxxxxxx 号
          </span>
        </div>
      </div>
    </footer>
  );
}
```

- [ ] **Step 2: Verify**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
```

Expected: pass.

- [ ] **Step 3: Browser check**

Reload any footer-including page (e.g., `/docs`, `/`, `/auth/signup`). Footer 服务条款 column has 5 working links. Clicking 用户协议 navigates to `/docs/account/user-agreement`. 斗学团队 column has 4 items (no `bs@douxue.cc`). Copyright reads `© 2026 douxue.fun 版权所有`.

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/components/in/footer.tsx
git commit -m "feat(web): link real legal docs from footer, drop email, fix domain"
```

---

## Task 21: Final verification

**Files:** none (verification only)

- [ ] **Step 1: Type check + lint + build**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npx tsc --noEmit
npm run lint
npm run build
```

Expected: all three succeed with zero errors, zero warnings. Build output confirms 5 new static legal routes.

- [ ] **Step 2: Functional regression checklist**

Dev server (`npm run dev`), walk through each row in order. If any fails, stop and fix.

| # | Flow | Expected |
|---|---|---|
| 1 | Visit `/docs/account` | 10 topics; 法律条款 divider before 用户协议 |
| 2 | Visit `/docs/account/user-agreement` | Placeholder notice at top; full doc renders; 11 clauses; prev = 常见问题, next = 隐私政策 |
| 3 | Walk all 5 legal topic URLs via prev/next | Continuous chain; last has no next |
| 4 | `/auth/signup`: click each of the 4 AgreementLinks | Correct dialog each time; scrollable body; close works (× / Esc / outside / "我已阅读") |
| 5 | Dialog → "在完整页面查看 →" | New tab → correct /docs/account/<slug>; signup form state preserved |
| 6 | Signup with unchecked agreement | Submit button disabled (existing behavior unchanged) |
| 7 | `/auth/signin`: inline hint line visible | 4 clickable agreement names; correct dialogs |
| 8 | `/purchase/payment/[orderId]` (any pending order) | Only 《产品服务协议》 link; dialog opens |
| 9 | Mobile viewport (375px) via devtools | Dialog full-width; sticky header + footer reachable; body scrolls |
| 10 | Footer on any page | 5 working legal anchors; `bs@douxue.cc` absent; copyright `douxue.fun` |
| 11 | Visit `/docs/learning-modes/overview` (unrelated category) | Sidebar / layout unchanged; no group divider |
| 12 | Visit `/hall` (existing protected page) | Loads normally; no regressions |

- [ ] **Step 3: Regression spot checks**

| Check | Expected |
|---|---|
| All 11 non-`account` `/docs` categories render identically to pre-feature | PASS |
| Signup `agreed` state machine unchanged | PASS |
| Payment `agreed` state machine unchanged | PASS |
| 斗学产品 / 斗学团队 footer columns visually identical except email removed | PASS |

- [ ] **Step 4: Final commit (only if any cleanup was needed during verification)**

If any cleanup commits were made, they're already done. Otherwise, verification produces no commit — just a thumbs-up.

- [ ] **Step 5: Merge to main locally (per project convention)**

If the work was done on a feature branch, merge to `main` locally. Per `feedback_git_workflow.md`: never push feature branches to remote; push only `main`.

```bash
# Only if on a feature branch:
git checkout main
git merge --no-ff <feature-branch>
git push origin main
```

If already on `main`, just push:

```bash
git push origin main
```

---

## Self-Review

**Spec coverage** — every spec section maps to a task:

| Spec section | Implementing task(s) |
|---|---|
| §3 Module layout | Tasks 4, 5, 6, 7, 8, 14 |
| §4 Data model | Task 4 (types, constants), Task 14 (registry) |
| §5 Content conventions | Tasks 9–13 (5 docs); `LegalPlaceholderNotice` in Task 5 |
| §6 /docs integration | Task 1 (`groupLabel` type), Tasks 2–3 (render), Task 15 (registry), Task 16 (verify) |
| §7 AgreementDialog UX | Task 6 (dialog), Task 7 (link), Task 8 (inline list) |
| §8.1 Signup wiring | Task 17 |
| §8.2 Signin wiring | Task 18 |
| §8.3 Payment wiring | Task 19 |
| §8.4 Footer wiring | Task 20 |
| §9 Touched files | All file paths match this plan's "File Structure" section |
| §10 Verification gates | Task 16 (milestone), Task 21 (final) |
| §11 Out of scope | Honored: no cookie banner, no real-name enforcement, no i18n, no placeholder fill-in |
| §12 Commit strategy | Implemented per-task |
| §13 Open risks | Legal-entity placeholders visible in docs (Task 5 + Tasks 9–13); content drift noted; `{{` grep-searchable |

**Placeholder scan:** no "TBD", "implement later", "appropriate error handling", or "similar to Task N" references — each task restates what it needs.

**Type consistency:** `LegalAgreementSlug`, `LegalAgreement`, `getAgreementBySlug`, `LEGAL_AGREEMENTS`, `UserAgreementDoc` (and siblings) — names match across Tasks 4, 6, 7, 8, 14, 15.

**Deferred-verification convention:** Tasks 6, 7, 8 (which import from the not-yet-existing `registry.ts`) explicitly defer `tsc --noEmit` and run only `lint` on the new file. Task 14 re-runs full tsc and confirms resolution. This is called out in each task's verify step.
