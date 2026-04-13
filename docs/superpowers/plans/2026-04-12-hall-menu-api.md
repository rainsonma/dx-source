# Hall Menu API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move hardcoded sidebar menus to a Go const file served via API, rename three menu items, and reorder 排行榜单 into the Personal section.

**Architecture:** New `hall_menu.go` const in dx-api defines the full sidebar structure (sections, items with icon/label/subtitle/href). A new `GET /api/hall/menus` endpoint returns it. Frontend sidebar and page top bars consume via SWR hook, resolving icon name strings through a Lucide icon registry.

**Tech Stack:** Go/Goravel (backend const + controller), Next.js/React/SWR/Lucide (frontend)

**Spec:** `docs/superpowers/specs/2026-04-12-hall-menu-api-design.md`

---

### Task 1: Backend — Hall Menu Constants

**Files:**
- Create: `dx-api/app/consts/hall_menu.go`

- [ ] **Step 1: Create the const file**

```go
package consts

// HallMenuItem represents a single sidebar navigation item.
type HallMenuItem struct {
	Icon     string `json:"icon"`
	Label    string `json:"label"`
	Subtitle string `json:"subtitle"`
	Href     string `json:"href"`
}

// HallMenuSection groups related sidebar navigation items.
type HallMenuSection struct {
	Items []HallMenuItem `json:"items"`
}

// HallMenuSections returns the complete sidebar menu structure.
func HallMenuSections() []HallMenuSection {
	return []HallMenuSection{
		{Items: []HallMenuItem{
			{Icon: "LayoutDashboard", Label: "我的主页", Subtitle: "", Href: "/hall"},
			{Icon: "Gamepad2", Label: "学习课程", Subtitle: "选择一个游戏模式，边玩边学英语！", Href: "/hall/games"},
			{Icon: "Gamepad2", Label: "我的游戏", Subtitle: "你玩过的所有课程游戏", Href: "/hall/games/mine"},
			{Icon: "Users", Label: "学习群组", Subtitle: "浏览并加入学习群组，与小伙伴一起进步", Href: "/hall/groups"},
			{Icon: "Star", Label: "我的收藏", Subtitle: "收藏你喜欢的课程游戏和学习内容", Href: "/hall/favorites"},
			{Icon: "Bell", Label: "消息通知", Subtitle: "查看系统通知和公告", Href: "/hall/notices"},
		}},
		{Items: []HallMenuItem{
			{Icon: "Sparkles", Label: "AI 随心学", Subtitle: "AI 驱动的个性化英语练习游戏", Href: "/hall/ai-custom"},
		}},
		{Items: []HallMenuItem{
			{Icon: "BookOpen", Label: "生词本", Subtitle: "记录你遇到的新单词和生词", Href: "/hall/unknown"},
			{Icon: "RotateCcw", Label: "复习本", Subtitle: "需要复习巩固的词汇和知识点", Href: "/hall/review"},
			{Icon: "CheckCircle2", Label: "已掌握", Subtitle: "你已经掌握的词汇和知识点", Href: "/hall/mastered"},
		}},
		{Items: []HallMenuItem{
			{Icon: "MessageCircle", Label: "斗学社", Subtitle: "分享学习心得，与学友互动交流", Href: "/hall/community"},
		}},
		{Items: []HallMenuItem{
			{Icon: "Trophy", Label: "排行榜单", Subtitle: "查看学习排名，与好友一起进步", Href: "/hall/leaderboard"},
			{Icon: "Medal", Label: "个人中心", Subtitle: "管理你的个人资料和账号信息", Href: "/hall/me"},
		}},
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/consts/hall_menu.go
git commit -m "feat(api): add hall menu constants"
```

---

### Task 2: Backend — API Endpoint

**Files:**
- Modify: `dx-api/app/http/controllers/api/hall_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Add GetMenus method to HallController**

In `dx-api/app/http/controllers/api/hall_controller.go`, add after the existing `GetHeatmap` method (before the `currentYear` helper):

```go
// GetMenus returns the sidebar menu structure.
func (c *HallController) GetMenus(ctx contractshttp.Context) contractshttp.Response {
	return helpers.Success(ctx, consts.HallMenuSections())
}
```

The file already imports `dx-api/app/consts` and `dx-api/app/helpers`, so no new imports needed.

- [ ] **Step 2: Add route**

In `dx-api/routes/api.go`, find the hall routes block (lines 172-174):

```go
			hallController := apicontrollers.NewHallController()
			protected.Get("/hall/dashboard", hallController.GetDashboard)
			protected.Get("/hall/heatmap", hallController.GetHeatmap)
```

Add the menus route after `/hall/heatmap`:

```go
			hallController := apicontrollers.NewHallController()
			protected.Get("/hall/dashboard", hallController.GetDashboard)
			protected.Get("/hall/heatmap", hallController.GetHeatmap)
			protected.Get("/hall/menus", hallController.GetMenus)
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 4: Verify with go vet**

Run: `cd dx-api && go vet ./...`
Expected: no issues

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/http/controllers/api/hall_controller.go dx-api/routes/api.go
git commit -m "feat(api): add hall menus endpoint"
```

---

### Task 3: Frontend — Types, Icon Registry, Hook, API Client

**Files:**
- Create: `dx-web/src/features/web/hall/types/hall-menu.types.ts`
- Create: `dx-web/src/features/web/hall/helpers/icon-registry.ts`
- Create: `dx-web/src/features/web/hall/hooks/use-hall-menu.ts`
- Modify: `dx-web/src/lib/api-client.ts`

- [ ] **Step 1: Create types file**

Create `dx-web/src/features/web/hall/types/hall-menu.types.ts`:

```ts
export type HallMenuItem = {
  icon: string
  label: string
  subtitle: string
  href: string
}

export type HallMenuSection = {
  items: HallMenuItem[]
}
```

- [ ] **Step 2: Create icon registry**

Create `dx-web/src/features/web/hall/helpers/icon-registry.ts`:

```ts
import type { LucideIcon } from "lucide-react"
import {
  LayoutDashboard,
  Gamepad2,
  Users,
  Star,
  Bell,
  Sparkles,
  BookOpen,
  RotateCcw,
  CheckCircle2,
  Trophy,
  Medal,
  MessageCircle,
} from "lucide-react"

export const iconRegistry: Record<string, LucideIcon> = {
  LayoutDashboard,
  Gamepad2,
  Users,
  Star,
  Bell,
  Sparkles,
  BookOpen,
  RotateCcw,
  CheckCircle2,
  Trophy,
  Medal,
  MessageCircle,
}
```

- [ ] **Step 3: Create hook**

Create `dx-web/src/features/web/hall/hooks/use-hall-menu.ts`:

```ts
import useSWR from "swr"
import type { HallMenuSection } from "@/features/web/hall/types/hall-menu.types"

export function useHallMenu() {
  return useSWR<HallMenuSection[]>("/api/hall/menus")
}

export function useHallMenuItem(href: string) {
  const { data: sections } = useHallMenu()
  if (!sections) return null
  for (const section of sections) {
    const item = section.items.find((i) => i.href === href)
    if (item) return item
  }
  return null
}
```

- [ ] **Step 4: Add getMenus to hallApi in api-client.ts**

In `dx-web/src/lib/api-client.ts`, find the `hallApi` object (line 439-448). Add `getMenus` after `getHeatmap`:

Replace:

```ts
export const hallApi = {
  /** Get aggregated dashboard data */
  async getDashboard() {
    return apiClient.get<unknown>("/api/hall/dashboard");
  },
  /** Get heatmap data for a given year */
  async getHeatmap(year: number) {
    return apiClient.get<unknown>(`/api/hall/heatmap?year=${year}`);
  },
};
```

With:

```ts
export const hallApi = {
  /** Get aggregated dashboard data */
  async getDashboard() {
    return apiClient.get<unknown>("/api/hall/dashboard");
  },
  /** Get heatmap data for a given year */
  async getHeatmap(year: number) {
    return apiClient.get<unknown>(`/api/hall/heatmap?year=${year}`);
  },
  /** Get sidebar menu structure */
  async getMenus() {
    return apiClient.get<unknown>("/api/hall/menus");
  },
};
```

- [ ] **Step 5: Run lint**

Run: `cd dx-web && npx eslint src/features/web/hall/types/hall-menu.types.ts src/features/web/hall/helpers/icon-registry.ts src/features/web/hall/hooks/use-hall-menu.ts src/lib/api-client.ts`
Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/hall/types/hall-menu.types.ts dx-web/src/features/web/hall/helpers/icon-registry.ts dx-web/src/features/web/hall/hooks/use-hall-menu.ts dx-web/src/lib/api-client.ts
git commit -m "feat(web): add hall menu types, icon registry, and hook"
```

---

### Task 4: Frontend — Refactor Sidebar to Use API Data

**Files:**
- Modify: `dx-web/src/features/web/hall/components/hall-sidebar.tsx`

- [ ] **Step 1: Replace the full file**

Replace the entire `hall-sidebar.tsx` with the API-driven version. Key changes:
- Remove the hardcoded `navSections` const (lines 33-65)
- Remove unused Lucide icon imports that were only for nav items
- Add `useSWR` import and `useHallMenu` hook
- Add `iconRegistry` import
- Resolve icon names from API data via the registry

New full content of `dx-web/src/features/web/hall/components/hall-sidebar.tsx`:

```tsx
"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  GraduationCap,
  Swords,
  Gift,
  ChevronRight,
  Crown,
  Ticket,
  Menu,
} from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useHallMenu } from "@/features/web/hall/hooks/use-hall-menu";
import { iconRegistry } from "@/features/web/hall/helpers/icon-registry";

function NavItem({
  icon: Icon,
  label,
  href,
  active,
  showDot,
  onClick,
}: {
  icon: React.ElementType;
  label: string;
  href: string;
  active: boolean;
  showDot?: boolean;
  onClick?: () => void;
}) {
  return (
    <Link
      href={href}
      onClick={onClick}
      className={`flex w-full items-center gap-3 rounded-[10px] px-3.5 py-2.5 ${
        active
          ? "bg-teal-600/10 font-semibold text-teal-600"
          : "text-muted-foreground hover:bg-accent"
      }`}
    >
      <Icon className="h-[18px] w-[18px]" />
      <span className="text-[13px]">{label}</span>
      {showDot && (
        <span className="ml-auto h-2 w-2 rounded-full bg-red-500" />
      )}
    </Link>
  );
}

/** CTA card configurations for the sidebar bottom section. */
const ctaItems = [
  {
    icon: Crown,
    label: "续费升级",
    subtitle: "选择会员套餐",
    href: "/purchase/membership",
    iconGradient: "from-teal-400 to-teal-600",
    badge: { text: "VIP", gradient: "from-amber-300 to-yellow-500" },
  },
  {
    icon: Ticket,
    label: "兑换码",
    subtitle: "兑换码兑换会员",
    href: "/hall/redeem",
    iconGradient: "from-violet-400 to-purple-600",
  },
  {
    icon: Gift,
    label: "推广有奖",
    subtitle: "推广、邀请、赚佣金",
    href: "/hall/invite",
    iconGradient: "from-orange-400 to-red-500",
    badge: { text: "HOT", gradient: "from-orange-400 to-red-500" },
  },
];

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
      className="flex w-full items-center justify-between rounded-[10px] border border-border px-3.5 py-3 hover:bg-accent"
    >
      <div className="flex items-center gap-3">
        <div
          className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br ${iconGradient}`}
        >
          <Icon className="h-4 w-4 text-white" />
        </div>
        <div className="flex flex-col gap-0.5">
          <div className="flex items-center gap-1.5">
            <span className="text-[13px] font-medium text-foreground">
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
          <span className="text-[11px] text-muted-foreground">{subtitle}</span>
        </div>
      </div>
      <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
    </Link>
  );
}

function SidebarContent({ onNavigate, hasUnreadNotices }: { onNavigate?: () => void; hasUnreadNotices?: boolean }) {
  const pathname = usePathname();
  const { data: navSections } = useHallMenu();

  return (
    <div className="flex h-full flex-col">
      {/* Header — pinned top */}
      <div className="shrink-0">
        <div className="flex w-full items-center justify-between">
          <Link href="/" className="flex items-center gap-2.5">
            <GraduationCap className="h-7 w-7 text-teal-600" />
            <span className="text-lg font-extrabold text-foreground">斗学</span>
          </Link>
          <Swords className="h-[18px] w-[18px] text-muted-foreground" />
        </div>
      </div>

      {/* Nav — scrollable middle */}
      <nav className="mt-8 flex-1 overflow-y-auto min-h-0 flex flex-col gap-1">
        {navSections?.map((section, si) => (
          <div key={si} className="flex flex-col gap-1">
            {si > 0 && (
              <div className="my-1 h-px w-full bg-border" />
            )}
            {section.items.map((item) => {
              const Icon = iconRegistry[item.icon];
              if (!Icon) return null;
              return (
                <NavItem
                  key={item.href}
                  icon={Icon}
                  label={item.label}
                  href={item.href}
                  active={pathname === item.href}
                  showDot={item.href === "/hall/notices" && hasUnreadNotices}
                  onClick={onNavigate}
                />
              );
            })}
          </div>
        ))}
      </nav>

      {/* Bottom CTAs — pinned bottom */}
      <div className="shrink-0 mt-4 flex flex-col gap-2">
        {ctaItems.map((item) => (
          <CtaCard key={item.label} {...item} onClick={onNavigate} />
        ))}
      </div>
    </div>
  );
}

export function HallSidebar({ hasUnreadNotices }: { hasUnreadNotices?: boolean }) {
  return (
    <aside className="hidden md:flex h-full w-[260px] shrink-0 flex-col border-r border-border bg-card px-5 py-6">
      <SidebarContent hasUnreadNotices={hasUnreadNotices} />
    </aside>
  );
}

export function MobileSidebarTrigger({ hasUnreadNotices }: { hasUnreadNotices?: boolean }) {
  return (
    <Sheet>
      <SheetTrigger asChild>
        <button
          type="button"
          className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground hover:bg-accent"
        >
          <Menu className="h-5 w-5" />
        </button>
      </SheetTrigger>
      <SheetContent side="left" className="w-[260px] p-5" showCloseButton={false}>
        <SidebarContent hasUnreadNotices={hasUnreadNotices} />
      </SheetContent>
    </Sheet>
  );
}
```

- [ ] **Step 2: Run lint on the sidebar**

Run: `cd dx-web && npx eslint src/features/web/hall/components/hall-sidebar.tsx`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/hall/components/hall-sidebar.tsx
git commit -m "refactor(web): make hall sidebar API-driven"
```

---

### Task 5: Frontend — Update Page Top Bars to Use Hook

**Files:**
- Modify: `dx-web/src/app/(web)/hall/(main)/games/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/groups/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/leaderboard/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/me/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/games/mine/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/favorites/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/notices/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/unknown/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/review/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/mastered/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/community/page.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/ai-custom/page.tsx`

Each page follows the same pattern: import `useHallMenuItem`, call it with the page's href, and pass the returned label/subtitle to the top bar component. Pages that are currently server components need `"use client"` added.

- [ ] **Step 1: Update games/page.tsx**

This page already has `"use client"` and uses `GreetingTopBar`. Add the hook import and replace hardcoded strings.

Replace:

```tsx
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar"
```

With:

```tsx
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
```

Replace:

```tsx
      <GreetingTopBar
        title="课程游戏"
        subtitle="选择一个游戏模式，边玩边学英语！"
      />
```

With:

```tsx
      <GreetingTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
```

And add the hook call at the top of the component function body (first line inside `HallGamesPage`):

```tsx
  const menu = useHallMenuItem("/hall/games")
```

- [ ] **Step 2: Update groups/page.tsx**

This page is currently a server component. Add `"use client"` and the hook.

Replace full file content:

```tsx
"use client"

import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
import { GroupListContent } from "@/features/web/groups/components/group-list-content"

export default function GroupsPage() {
  const menu = useHallMenuItem("/hall/groups")

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索群组..."
      />
      <GroupListContent />
    </div>
  )
}
```

- [ ] **Step 3: Update leaderboard/page.tsx**

Already has `"use client"`. Add hook import and replace strings.

Replace full file content:

```tsx
"use client";

import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
import { LeaderboardContent } from "@/features/web/leaderboard/components/leaderboard-content";

export default function LeaderboardPage() {
  const menu = useHallMenuItem("/hall/leaderboard");

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索用户..."
      />
      <LeaderboardContent />
    </div>
  );
}
```

- [ ] **Step 4: Update me/page.tsx**

Already has `"use client"`. Add hook import and replace strings. Keep all existing imports and logic.

Add import:

```tsx
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
```

Add hook call as first line in the component (before the `useSWR` call):

```tsx
  const menu = useHallMenuItem("/hall/me");
```

Replace:

```tsx
      <PageTopBar
        title="个人中心"
        subtitle="管理你的个人资料和账号信息"
      />
```

With:

```tsx
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
```

- [ ] **Step 5: Update games/mine/page.tsx**

Already has `"use client"`. Add hook import and replace strings.

Add import:

```tsx
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
```

Add hook call as first line in `MyGamesPage` (before `useState`):

```tsx
  const menu = useHallMenuItem("/hall/games/mine");
```

Replace:

```tsx
      <PageTopBar
        title="我的游戏"
        subtitle="你玩过的所有课程游戏"
        searchPlaceholder="搜索游戏..."
      />
```

With:

```tsx
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索游戏..."
      />
```

- [ ] **Step 6: Update favorites/page.tsx**

Already has `"use client"`. Add hook import and replace strings.

Add import:

```tsx
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
```

Add hook call as first line in `FavoritesPage` (before `useState`):

```tsx
  const menu = useHallMenuItem("/hall/favorites");
```

Replace:

```tsx
      <PageTopBar
        title="我的收藏"
        subtitle="收藏你喜欢的课程游戏和学习内容"
        searchPlaceholder="搜索收藏..."
      />
```

With:

```tsx
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索收藏..."
      />
```

- [ ] **Step 7: Update notices/page.tsx**

Already has `"use client"`. Add hook import and replace strings.

Add import:

```tsx
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
```

Add hook call as first line in `NoticesPage` (before `useState`):

```tsx
  const menu = useHallMenuItem("/hall/notices");
```

Replace:

```tsx
      <PageTopBar
        title="消息通知"
        subtitle="查看系统通知和公告"
      />
```

With:

```tsx
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
```

- [ ] **Step 8: Update unknown/page.tsx**

Already has `"use client"`. Add hook import and replace strings.

Add import:

```tsx
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
```

Add hook call as first line in `UnknownPage` (before `useState`):

```tsx
  const menu = useHallMenuItem("/hall/unknown");
```

Replace:

```tsx
      <PageTopBar
        title="生词本"
        subtitle="记录你遇到的新单词和生词"
        searchPlaceholder="搜索生词..."
      />
```

With:

```tsx
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索生词..."
      />
```

- [ ] **Step 9: Update review/page.tsx**

Already has `"use client"`. Add hook import and replace strings.

Add import:

```tsx
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
```

Add hook call as first line in `ReviewPage` (before `useState`):

```tsx
  const menu = useHallMenuItem("/hall/review");
```

Replace:

```tsx
      <PageTopBar
        title="复习本"
        subtitle="需要复习巩固的词汇和知识点"
        searchPlaceholder="搜索复习内容..."
      />
```

With:

```tsx
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索复习内容..."
      />
```

- [ ] **Step 10: Update mastered/page.tsx**

Already has `"use client"`. Add hook import and replace strings.

Add import:

```tsx
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
```

Add hook call as first line in `MasterPage` (before `useState`):

```tsx
  const menu = useHallMenuItem("/hall/mastered");
```

Replace:

```tsx
      <PageTopBar
        title="已掌握"
        subtitle="你已经掌握的词汇和知识点"
        searchPlaceholder="搜索已掌握内容..."
      />
```

With:

```tsx
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索已掌握内容..."
      />
```

- [ ] **Step 11: Update community/page.tsx**

Currently a server component — add `"use client"`.

Replace full file content:

```tsx
"use client"

import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
import { CommunityFeed } from "@/features/web/community/components/community-feed"

export default function CommunityPage() {
  const menu = useHallMenuItem("/hall/community")

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
      <CommunityFeed />
    </div>
  )
}
```

- [ ] **Step 12: Update ai-custom/page.tsx**

Currently a server component — add `"use client"`.

Replace full file content:

```tsx
"use client"

import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
import { AiCustomGrid } from "@/features/web/ai-custom/components/ai-custom-grid"

export default function AiCustomPage() {
  const menu = useHallMenuItem("/hall/ai-custom")

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
      <AiCustomGrid />
    </div>
  )
}
```

- [ ] **Step 13: Run lint on all modified page files**

Run: `cd dx-web && npx eslint "src/app/(web)/hall/(main)/**/page.tsx"`
Expected: no errors

- [ ] **Step 14: Run full build**

Run: `cd dx-web && npm run build`
Expected: build succeeds with no errors

- [ ] **Step 15: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/app/\(web\)/hall/\(main\)/
git commit -m "refactor(web): use hall menu hook in page top bars"
```

---

### Task 6: Verification

- [ ] **Step 1: Start dx-api and verify endpoint**

Run: `cd dx-api && go run .`

Then in another terminal:
```bash
curl -s http://localhost:3001/api/hall/menus -H "Authorization: Bearer <token>" | jq '.data | length'
```
Expected: `5` (five sections)

- [ ] **Step 2: Start dx-web and verify sidebar**

Run: `cd dx-web && npm run dev`

Open http://localhost:3000/hall in browser and verify:
- Sidebar shows all menu items with correct new labels (学习课程, 学习群组, 排行榜单)
- 排行榜单 appears in section 5 (above 个人中心), not in section 3
- All icons render correctly
- Click each menu item — page top bar shows matching title/subtitle
- Mobile sidebar (narrow viewport) works the same
- Notification dot on 消息通知 still works

- [ ] **Step 3: Verify no regressions**

- Navigate to every hall page and confirm no blank titles
- Confirm games page still shows GreetingTopBar style (not PageTopBar)
- Confirm home page still shows dynamic greeting
- Confirm invite, redeem pages still show their hardcoded titles
- Confirm the 3 server-to-client conversions (groups, community, ai-custom) render correctly
