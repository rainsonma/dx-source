# Hall Menu API Design

## Summary

Move sidebar menu definitions from hardcoded frontend to a Go constants file in dx-api, served via API. Rename three menu items and reorder one.

## Text Changes

| Current | New | Notes |
|---------|-----|-------|
| 排行榜 | 排行榜单 | Also moved from section 3 (Learning Tools) to section 5 (Personal) |
| 课程游戏 | 学习课程 | Label + top bar title |
| 游戏群组 | 学习群组 | Label + top bar title + subtitle ("浏览并加入**学习群组**...") |

## Menu Structure (After Changes)

```
Section 1 — Main Navigation:
  我的主页      LayoutDashboard  /hall
  学习课程      Gamepad2         /hall/games
  我的游戏      Gamepad2         /hall/games/mine
  学习群组      Users            /hall/groups
  我的收藏      Star             /hall/favorites
  消息通知      Bell             /hall/notices

Section 2 — AI Learning:
  AI 随心学     Sparkles         /hall/ai-custom

Section 3 — Learning Tools:
  生词本        BookOpen         /hall/unknown
  复习本        RotateCcw        /hall/review
  已掌握        CheckCircle2     /hall/mastered

Section 4 — Community:
  斗学社        MessageCircle    /hall/community

Section 5 — Personal:
  排行榜单      Trophy           /hall/leaderboard    ← moved here
  个人中心      Medal            /hall/me
```

## Backend

### 1. New file: `dx-api/app/consts/hall_menu.go`

Structs:

```go
type HallMenuItem struct {
    Icon     string `json:"icon"`
    Label    string `json:"label"`
    Subtitle string `json:"subtitle"`
    Href     string `json:"href"`
}

type HallMenuSection struct {
    Items []HallMenuItem `json:"items"`
}
```

Function `HallMenuSections() []HallMenuSection` returns the full menu structure with all 5 sections, items, icon names, labels, subtitles, and hrefs as shown above.

Subtitles per item:

| Item | Subtitle |
|------|----------|
| 我的主页 | (empty — home page uses dynamic greeting) |
| 学习课程 | 选择一个游戏模式，边玩边学英语！ |
| 我的游戏 | 你玩过的所有课程游戏 |
| 学习群组 | 浏览并加入学习群组，与小伙伴一起进步 |
| 我的收藏 | 收藏你喜欢的课程游戏和学习内容 |
| 消息通知 | 查看系统通知和公告 |
| AI 随心学 | AI 驱动的个性化英语练习游戏 |
| 生词本 | 记录你遇到的新单词和生词 |
| 复习本 | 需要复习巩固的词汇和知识点 |
| 已掌握 | 你已经掌握的词汇和知识点 |
| 斗学社 | 分享学习心得，与学友互动交流 |
| 排行榜单 | 查看学习排名，与好友一起进步 |
| 个人中心 | 管理你的个人资料和账号信息 |

### 2. New method on `HallController`

`GetMenus` — returns `helpers.Success(ctx, consts.HallMenuSections())`. No DB access, no auth ID needed (but route is still protected since the hall requires JWT).

### 3. New route

```go
protected.Get("/hall/menus", hallController.GetMenus)
```

Added alongside existing `/hall/dashboard` and `/hall/heatmap`.

### API Response Shape

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "items": [
        { "icon": "LayoutDashboard", "label": "我的主页", "subtitle": "", "href": "/hall" },
        { "icon": "Gamepad2", "label": "学习课程", "subtitle": "选择一个游戏模式，边玩边学英语！", "href": "/hall/games" }
      ]
    }
  ]
}
```

## Frontend

### 1. New file: `dx-web/src/features/web/hall/helpers/icon-registry.ts`

Maps icon name strings from the API to Lucide React components:

```ts
import type { LucideIcon } from "lucide-react"
import { LayoutDashboard, Gamepad2, Users, Star, Bell, Sparkles, BookOpen,
         RotateCcw, CheckCircle2, Trophy, Medal, MessageCircle } from "lucide-react"

export const iconRegistry: Record<string, LucideIcon> = {
  LayoutDashboard, Gamepad2, Users, Star, Bell, Sparkles,
  BookOpen, RotateCcw, CheckCircle2, Trophy, Medal, MessageCircle,
}
```

### 2. New file: `dx-web/src/features/web/hall/types/hall-menu.types.ts`

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

### 3. New hook: `dx-web/src/features/web/hall/hooks/use-hall-menu.ts`

- `useHallMenu()` — wraps `useSWR<HallMenuSection[]>("/api/hall/menus")`
- `useHallMenuItem(href: string)` — calls `useHallMenu()`, finds item by href match

### 4. Modified: `hall-sidebar.tsx`

- Remove hardcoded `navSections` const
- Call `useHallMenu()` to get sections from API
- Resolve icon names via `iconRegistry[item.icon]`
- Render sections and items as before

### 5. Modified page files (use `useHallMenuItem` hook)

Pages that correspond to sidebar items replace hardcoded title/subtitle with the hook:

| Page | Change |
|------|--------|
| `games/page.tsx` | `useHallMenuItem("/hall/games")` for GreetingTopBar title/subtitle |
| `groups/page.tsx` | `useHallMenuItem("/hall/groups")` for PageTopBar title/subtitle; add `"use client"` |
| `leaderboard/page.tsx` | `useHallMenuItem("/hall/leaderboard")` for PageTopBar title/subtitle |
| `me/page.tsx` | `useHallMenuItem("/hall/me")` for PageTopBar title/subtitle |
| `games/mine/page.tsx` | `useHallMenuItem("/hall/games/mine")` for PageTopBar title/subtitle |
| `favorites/page.tsx` | `useHallMenuItem("/hall/favorites")` for PageTopBar title/subtitle |
| `notices/page.tsx` | `useHallMenuItem("/hall/notices")` for PageTopBar title/subtitle |
| `unknown/page.tsx` | `useHallMenuItem("/hall/unknown")` for PageTopBar title/subtitle |
| `review/page.tsx` | `useHallMenuItem("/hall/review")` for PageTopBar title/subtitle |
| `mastered/page.tsx` | `useHallMenuItem("/hall/mastered")` for PageTopBar title/subtitle |
| `community/page.tsx` | `useHallMenuItem("/hall/community")` for PageTopBar title/subtitle; add `"use client"` |
| `ai-custom/page.tsx` | `useHallMenuItem("/hall/ai-custom")` for PageTopBar title/subtitle; add `"use client"` |

**Not changed** (not sidebar items):
- `(home)/page.tsx` — keeps dynamic greeting
- `invite/page.tsx` — CTA, not a nav item
- `redeem/page.tsx` — CTA, not a nav item
- `ai-practice/page.tsx` — hidden/disabled

### 6. Frontend API client

Add to `hallApi` in `api-client.ts`:

```ts
getMenus: () => apiClient.get<HallMenuSection[]>("/api/hall/menus"),
```

## Files Changed

**New files:**
- `dx-api/app/consts/hall_menu.go`
- `dx-web/src/features/web/hall/helpers/icon-registry.ts`
- `dx-web/src/features/web/hall/types/hall-menu.types.ts`
- `dx-web/src/features/web/hall/hooks/use-hall-menu.ts`

**Modified files:**
- `dx-api/app/http/controllers/api/hall_controller.go` — add `GetMenus`
- `dx-api/routes/api.go` — add menu route
- `dx-web/src/lib/api-client.ts` — add `getMenus` to `hallApi`
- `dx-web/src/features/web/hall/components/hall-sidebar.tsx` — API-driven
- 12 page files under `dx-web/src/app/(web)/hall/(main)/` — use `useHallMenuItem` hook
