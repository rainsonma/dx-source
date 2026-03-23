# Sidebar CTA Enhancement Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enhance the hall sidebar's three bottom CTA cards with subtitles, colored icon backgrounds, gradient badges (HOT/VIP), proper linking, and auth-protect the membership page.

**Architecture:** Pure UI changes to `hall-sidebar.tsx` (data-driven CTA config + component extraction) and a server-side auth guard on the membership page.

**Tech Stack:** Next.js, React, Tailwind CSS, NextAuth (`auth()`), Lucide icons

---

### Task 1: Add Logo Link to Home

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx:104-108`

**Step 1: Wrap logo in Link**

In `SidebarContent`, replace the logo `<div>` wrapper with a `<Link>`:

```tsx
<Link href="/" className="flex items-center gap-2.5">
  <GraduationCap className="h-7 w-7 text-teal-600" />
  <span className="text-lg font-extrabold text-slate-900">斗学</span>
</Link>
```

**Step 2: Verify**

Run: `npm run build`
Expected: No errors. Logo in sidebar is now a clickable link to `/`.

**Step 3: Commit**

```
feat: link sidebar logo to home landing page
```

---

### Task 2: Enhance Bottom CTA Cards

**Files:**
- Modify: `src/features/web/hall/components/hall-sidebar.tsx:139-173`

**Step 1: Define CTA config array**

Above `SidebarContent`, add a data-driven config for the 3 CTAs:

```tsx
const ctaItems = [
  {
    icon: Gift,
    label: "推广有奖",
    subtitle: "推广、邀请、赚佣金",
    href: "/hall/invite",
    iconGradient: "from-orange-400 to-red-500",
    badge: { text: "HOT", gradient: "from-orange-400 to-red-500" },
  },
  {
    icon: Ticket,
    label: "兑换码",
    subtitle: "兑换码兑换会员",
    href: "/hall/redeem",
    iconGradient: "from-violet-400 to-purple-600",
  },
  {
    icon: ArrowUpCircle,
    label: "续费升级",
    subtitle: "选择会员套餐",
    href: "/auth/membership",
    iconGradient: "from-amber-300 to-yellow-500",
    badge: { text: "VIP", gradient: "from-amber-300 to-yellow-500" },
  },
];
```

**Step 2: Create CtaCard component**

Add a local `CtaCard` component inside the file:

```tsx
function CtaCard({
  icon: Icon,
  label,
  subtitle,
  href,
  iconGradient,
  badge,
  onClick,
}: {
  icon: React.ElementType;
  label: string;
  subtitle: string;
  href: string;
  iconGradient: string;
  badge?: { text: string; gradient: string };
  onClick?: () => void;
}) {
  return (
    <Link
      href={href}
      onClick={onClick}
      className="flex w-full items-center justify-between rounded-[10px] border border-slate-200 px-3.5 py-3 hover:bg-slate-50"
    >
      <div className="flex items-center gap-3">
        <div
          className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br ${iconGradient}`}
        >
          <Icon className="h-4 w-4 text-white" />
        </div>
        <div className="flex flex-col gap-0.5">
          <div className="flex items-center gap-1.5">
            <span className="text-[13px] font-medium text-slate-700">
              {label}
            </span>
            {badge && (
              <span
                className={`rounded-full bg-gradient-to-r px-1.5 py-0.5 text-[10px] font-semibold text-white ${badge.gradient}`}
              >
                {badge.text}
              </span>
            )}
          </div>
          <span className="text-[11px] text-slate-400">{subtitle}</span>
        </div>
      </div>
      <ChevronRight className="h-4 w-4 shrink-0 text-slate-400" />
    </Link>
  );
}
```

**Step 3: Replace bottom CTAs with mapped CtaCards**

Replace the entire `{/* Bottom CTAs */}` section with:

```tsx
{/* Bottom CTAs */}
<div className="flex flex-col gap-2">
  {ctaItems.map((item) => (
    <CtaCard key={item.label} {...item} onClick={onNavigate} />
  ))}
</div>
```

**Step 4: Verify**

Run: `npm run build`
Expected: No errors. Three CTA cards render with colored icon backgrounds, subtitles, and gradient badges.

**Step 5: Commit**

```
feat: enhance sidebar CTA cards with subtitles, icon backgrounds, and badges
```

---

### Task 3: Auth-Protect Membership Page

**Files:**
- Modify: `src/app/(web)/auth/membership/page.tsx`

**Step 1: Add auth guard**

Convert to async server component with `auth()` check:

```tsx
import { redirect } from "next/navigation";
import { CircleCheck } from "lucide-react";
import { auth } from "@/lib/auth";
import { PricingGrid } from "@/features/web/auth/components/pricing-grid";
import { TestimonialsGrid } from "@/features/web/auth/components/testimonials-grid";
import { FaqSection } from "@/features/web/auth/components/faq-section";

export default async function MembershipPage() {
  const session = await auth();
  if (!session?.user) redirect("/auth/signin");

  return (
    // ... existing JSX unchanged
  );
}
```

**Step 2: Verify**

Run: `npm run build`
Expected: No errors. Unauthenticated users visiting `/auth/membership` are redirected to `/auth/signin`.

**Step 3: Commit**

```
feat: auth-protect membership page
```
