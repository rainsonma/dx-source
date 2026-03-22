# SWR Cache Invalidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace broken `router.refresh()` with SWR cache invalidation so mutations instantly update the UI without manual page refresh.

**Architecture:** Install SWR, create a global fetcher wrapping `apiClient.get()`, add a prefix-based `swrMutate()` helper, then incrementally migrate each feature's data fetching from `useEffect` + `useState` to `useSWR` / `useSWRInfinite`.

**Tech Stack:** SWR, Next.js 16 (React 19), existing `apiClient` from `dx-web/src/lib/api-client.ts`

**Spec:** `docs/superpowers/specs/2026-03-22-swr-cache-invalidation-design.md`

---

## File Map

### New files
| File | Responsibility |
|------|---------------|
| `dx-web/src/lib/swr.ts` | SWR fetcher, `swrMutate` helper, `SWRProvider` component |
| `dx-web/src/components/in/page-spinner.tsx` | Shared loading spinner (sm/md/lg) |
| `dx-web/src/features/web/games/helpers/game-card.ts` | Consolidated `toPublicGameCard` transform |

### Modified files
| File | Change |
|------|--------|
| `dx-web/src/components/in/auth-guard.tsx` | Wrap children with `SWRProvider` |
| `dx-web/src/app/(web)/hall/(main)/ai-custom/page.tsx` | Replace `useEffect` fetch with `useSWR` |
| `dx-web/src/features/web/ai-custom/components/ai-custom-grid.tsx` | Remove `initialGames`/`initialCursor` props, use SWR hooks directly |
| `dx-web/src/features/web/ai-custom/hooks/use-infinite-games.ts` | Rewrite with `useSWRInfinite` |
| `dx-web/src/features/web/ai-custom/hooks/use-create-course-game.ts` | `swrMutate` replaces `router.refresh()` |
| `dx-web/src/features/web/ai-custom/hooks/use-create-game-level.ts` | Same |
| `dx-web/src/features/web/ai-custom/hooks/use-update-course-game.ts` | Same |
| `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts` | Same (4 hooks) |
| `dx-web/src/features/web/ai-custom/components/add-metadata-dialog.tsx` | Same |
| `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx` | Same (3 sites) |
| `dx-web/src/features/web/ai-custom/components/course-detail-content.tsx` | Replace multi-`useEffect` with 3 `useSWR` calls |
| `dx-web/src/app/(web)/hall/(main)/ai-custom/[id]/[levelId]/page.tsx` | Replace `useEffect` fetch with `useSWR` |
| `dx-web/src/features/web/games/hooks/use-infinite-public-games.ts` | Rewrite with `useSWRInfinite` |
| `dx-web/src/features/web/games/components/games-page-content.tsx` | Remove `initialGames`/`initialCursor` props, use SWR hook with local filter state |
| `dx-web/src/app/(web)/hall/(main)/games/page.tsx` | Replace `useEffect` fetch with `useSWR` |
| `dx-web/src/features/web/notice/components/mark-notices-read.tsx` | `swrMutate` replaces `router.refresh()` |

### Deleted code
| File | What |
|------|------|
| `dx-web/src/features/web/ai-custom/actions/course-game.action.ts` | Remove `fetchUserGamesAction` function and its type |
| `dx-web/src/features/web/games/actions/game.action.ts` | Remove `fetchPublicGamesAction` function and its type; keep `PublicGameCard` type |

---

## Task 1: Install SWR and Create Foundation

**Files:**
- Create: `dx-web/src/lib/swr.ts`
- Create: `dx-web/src/components/in/page-spinner.tsx`
- Modify: `dx-web/src/components/in/auth-guard.tsx`

- [ ] **Step 1: Install SWR**

```bash
cd dx-web && npm install swr
```

- [ ] **Step 2: Create `lib/swr.ts`**

Create `dx-web/src/lib/swr.ts`:

```ts
"use client"

import { SWRConfig } from "swr"
import { mutate } from "swr"
import type { ReactNode } from "react"
import { apiClient } from "@/lib/api-client"

export const swrFetcher = (url: string) =>
  apiClient.get(url).then((res) => {
    if (res.code !== 0) throw new Error(res.message)
    return res.data
  })

/**
 * Invalidate all SWR cache entries whose key starts with any of the given prefixes.
 * Call after mutations to trigger refetch.
 */
export async function swrMutate(...keys: string[]) {
  await Promise.all(
    keys.map((key) =>
      mutate(
        (k) => typeof k === "string" && k.startsWith(key),
        undefined,
        { revalidate: true }
      )
    )
  )
}

export function SWRProvider({ children }: { children: ReactNode }) {
  return (
    <SWRConfig value={{ fetcher: swrFetcher }}>
      {children}
    </SWRConfig>
  )
}
```

- [ ] **Step 3: Create `components/in/page-spinner.tsx`**

Create `dx-web/src/components/in/page-spinner.tsx`:

```tsx
import { Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"

type PageSpinnerProps = {
  size?: "sm" | "md" | "lg"
  className?: string
}

const sizeMap = { sm: "h-4 w-4", md: "h-5 w-5", lg: "h-6 w-6" }
const paddingMap = { sm: "py-4", md: "py-12", lg: "py-20" }

export function PageSpinner({ size = "md", className }: PageSpinnerProps) {
  return (
    <div className={cn("flex items-center justify-center", paddingMap[size], className)}>
      <Loader2 className={cn("animate-spin text-muted-foreground", sizeMap[size])} />
    </div>
  )
}
```

- [ ] **Step 4: Wire SWRProvider into AuthGuard**

Modify `dx-web/src/components/in/auth-guard.tsx`:

Add import at top:
```ts
import { SWRProvider } from "@/lib/swr"
```

Change the return on line 39 from:
```tsx
return <>{children}</>;
```
to:
```tsx
return <SWRProvider>{children}</SWRProvider>;
```

- [ ] **Step 5: Verify build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds. No runtime changes yet — SWR is wired but no hooks use it.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/lib/swr.ts dx-web/src/components/in/page-spinner.tsx dx-web/src/components/in/auth-guard.tsx dx-web/package.json dx-web/package-lock.json
git commit -m "feat: add SWR foundation — fetcher, swrMutate helper, SWRProvider, PageSpinner"
```

---

## Task 2: Migrate ai-custom List Page and Infinite Scroll

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/hooks/use-infinite-games.ts`
- Modify: `dx-web/src/features/web/ai-custom/components/ai-custom-grid.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/ai-custom/page.tsx`
- Modify: `dx-web/src/features/web/ai-custom/actions/course-game.action.ts` (delete `fetchUserGamesAction`)

- [ ] **Step 1: Rewrite `use-infinite-games.ts` with `useSWRInfinite`**

Replace entire content of `dx-web/src/features/web/ai-custom/hooks/use-infinite-games.ts`:

```ts
"use client"

import { useRef, useEffect, useCallback } from "react"
import useSWRInfinite from "swr/infinite"

type StatusFilter = "all" | "published" | "withdraw" | "draft"

export function useInfiniteGames(status: StatusFilter = "all") {
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const getKey = (pageIndex: number, previousPageData: any) => {
    if (previousPageData && !previousPageData.hasMore) return null
    const cursor = previousPageData?.nextCursor
    const params = new URLSearchParams()
    if (status !== "all") params.set("status", status)
    if (cursor) params.set("cursor", cursor)
    const qs = params.toString()
    return `/api/course-games${qs ? `?${qs}` : ""}`
  }

  const { data, size, setSize, isLoading, isValidating, mutate } =
    useSWRInfinite(getKey)

  const games = data?.flatMap((page) => page.items) ?? []
  const hasMore = data?.[data.length - 1]?.hasMore ?? false

  const loadMore = useCallback(() => {
    if (!isValidating && hasMore) setSize(size + 1)
  }, [isValidating, hasMore, size, setSize])

  // IntersectionObserver for infinite scroll
  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) loadMore()
      },
      { rootMargin: "200px" }
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore])

  return { games, isLoading, isValidating, hasMore, sentinelRef, mutate }
}
```

- [ ] **Step 2: Rewrite `ai-custom-grid.tsx` to use SWR hooks directly**

Modify `dx-web/src/features/web/ai-custom/components/ai-custom-grid.tsx`.

Replace the entire file. Key changes:
- Remove `initialGames`, `initialCursor` props
- Use `useInfiniteGames(activeFilter)` with `activeFilter` as status
- Use `useSWR('/api/course-games/counts')` for counts
- Use `useSWR('/api/game-categories')` and `useSWR('/api/game-presses')` for selects
- Show `PageSpinner` when `isLoading`
- Show inline spinner when `isValidating`

```tsx
"use client"

import { useState } from "react"
import useSWR from "swr"
import { Puzzle, Plus } from "lucide-react"
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { VisuallyHidden } from "@radix-ui/react-visually-hidden"
import { PageSpinner } from "@/components/in/page-spinner"
import { CreateCourseForm } from "@/features/web/ai-custom/components/create-course-form"
import { GameCardItem } from "@/features/web/ai-custom/components/game-card-item"
import { useInfiniteGames } from "@/features/web/ai-custom/hooks/use-infinite-games"

type GameCounts = { all: number; published: number; withdraw: number; draft: number }
type StatusFilter = "all" | "published" | "withdraw" | "draft"

const filterKeys: { key: StatusFilter; label: string; countKey: keyof GameCounts }[] = [
  { key: "all", label: "全部", countKey: "all" },
  { key: "published", label: "已发布", countKey: "published" },
  { key: "withdraw", label: "已撤回", countKey: "withdraw" },
  { key: "draft", label: "未发布", countKey: "draft" },
]

export function AiCustomGrid() {
  const [open, setOpen] = useState(false)
  const [activeFilter, setActiveFilter] = useState<StatusFilter>("all")

  const { data: categories } = useSWR<any[]>("/api/game-categories")
  const { data: presses } = useSWR<any[]>("/api/game-presses")
  const { data: counts } = useSWR<GameCounts>("/api/course-games/counts")
  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfiniteGames(activeFilter)

  const safeCounts = counts ?? { all: 0, draft: 0, published: 0, withdraw: 0 }

  return (
    <>
      {/* Hero Banner */}
      <div className="flex flex-col gap-4 rounded-2xl bg-gradient-to-tr from-teal-700 via-teal-600 to-purple-400 p-5 pb-6 lg:p-7 lg:pb-8">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-[14px] bg-white/10">
            <Puzzle className="h-7 w-7 text-white" />
          </div>
          <span className="text-xl font-extrabold tracking-tight text-white lg:text-[28px]">
            AI 随心配
          </span>
        </div>
        <p className="text-sm leading-relaxed text-white/80">
          随心定义想要的学习内容，拆解学习单元，通过趣味游戏模式巩固，记忆，理解，轻松掌握学习内容
        </p>
      </div>

      {/* Filter row */}
      <div className="flex w-full flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-wrap items-center gap-2">
          {filterKeys.map((tab) => (
            <button
              key={tab.key}
              type="button"
              onClick={() => setActiveFilter(tab.key)}
              className={`rounded-lg px-3.5 py-1.5 text-[13px] font-semibold ${
                activeFilter === tab.key
                  ? "bg-teal-600 text-white"
                  : "border border-border bg-muted text-muted-foreground"
              }`}
            >
              {tab.label}（{safeCounts[tab.countKey]}）
            </button>
          ))}
        </div>
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="flex items-center gap-1.5 rounded-[10px] bg-gradient-to-b from-teal-600 to-teal-700 px-4 py-2 text-[13px] font-semibold text-white"
        >
          <Plus className="h-3.5 w-3.5" />
          创建我的课程游戏
        </button>
      </div>

      {/* Loading state */}
      {isLoading && <PageSpinner size="lg" />}

      {/* Game grid */}
      {!isLoading && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
          {games.map((game) => (
            <GameCardItem key={game.id} game={game} />
          ))}
        </div>
      )}

      {/* Validating indicator */}
      {isValidating && !isLoading && <PageSpinner size="sm" />}

      {/* Empty state */}
      {!isLoading && !isValidating && games.length === 0 && (
        <div className="flex flex-col items-center gap-2 py-12 text-center">
          <Puzzle className="h-10 w-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">还没有创建任何课程游戏</p>
          <button
            type="button"
            onClick={() => setOpen(true)}
            className="mt-2 rounded-lg bg-teal-600 px-4 py-2 text-sm font-medium text-white"
          >
            创建第一个游戏
          </button>
        </div>
      )}

      {/* Infinite scroll sentinel */}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      {/* Create course game dialog */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent
          aria-describedby={undefined}
          showCloseButton={false}
          className="max-w-[560px] overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
        >
          <VisuallyHidden>
            <DialogTitle>创建我的课程游戏</DialogTitle>
          </VisuallyHidden>
          <CreateCourseForm
            categories={categories ?? []}
            presses={presses ?? []}
            onClose={() => setOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </>
  )
}
```

- [ ] **Step 3: Simplify `ai-custom/page.tsx`**

Replace entire content of `dx-web/src/app/(web)/hall/(main)/ai-custom/page.tsx`:

```tsx
import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { AiCustomGrid } from "@/features/web/ai-custom/components/ai-custom-grid"

export default function AiCustomPage() {
  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="AI 随心配"
        subtitle="AI 驱动的个性化英语练习游戏"
      />
      <AiCustomGrid />
    </div>
  )
}
```

Note: This page is now a **server component** (no `"use client"` needed — all dynamic data is inside `AiCustomGrid`).

- [ ] **Step 4: Delete `fetchUserGamesAction` from `course-game.action.ts`**

Remove lines 58-92 from `dx-web/src/features/web/ai-custom/actions/course-game.action.ts` — the `FetchUserGamesResult` type and `fetchUserGamesAction` function. SWR now fetches directly via the global fetcher.

- [ ] **Step 5: Verify build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds. The ai-custom list page now fetches via SWR.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/ai-custom/hooks/use-infinite-games.ts dx-web/src/features/web/ai-custom/components/ai-custom-grid.tsx dx-web/src/app/\(web\)/hall/\(main\)/ai-custom/page.tsx dx-web/src/features/web/ai-custom/actions/course-game.action.ts
git commit -m "feat: migrate ai-custom list page to SWR with useSWRInfinite"
```

---

## Task 3: Replace `router.refresh()` in ai-custom Mutation Hooks

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/hooks/use-create-course-game.ts`
- Modify: `dx-web/src/features/web/ai-custom/hooks/use-create-game-level.ts`
- Modify: `dx-web/src/features/web/ai-custom/hooks/use-update-course-game.ts`
- Modify: `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts`

- [ ] **Step 1: Fix `use-create-course-game.ts`**

Replace entire content of `dx-web/src/features/web/ai-custom/hooks/use-create-course-game.ts`:

```ts
"use client"

import { useActionState, useState, useEffect, useRef } from "react"
import { swrMutate } from "@/lib/swr"

import {
  createCourseGameAction,
  type CreateCourseGameResult,
} from "@/features/web/ai-custom/actions/course-game.action"

const initialState: CreateCourseGameResult = {}

export function useCreateCourseGame(onSuccess?: () => void) {
  const [coverId, setCoverId] = useState<string | null>(null)
  const onSuccessRef = useRef(onSuccess)
  onSuccessRef.current = onSuccess

  const [state, formAction, isPending] = useActionState(
    createCourseGameAction,
    initialState
  )

  useEffect(() => {
    if (state.success) {
      onSuccessRef.current?.()
      swrMutate("/api/course-games")
    }
  }, [state])

  return {
    state,
    formAction,
    isPending,
    coverId,
    setCoverId,
  }
}
```

- [ ] **Step 2: Fix `use-create-game-level.ts`**

Replace entire content of `dx-web/src/features/web/ai-custom/hooks/use-create-game-level.ts`:

```ts
"use client"

import { useActionState, useEffect, useRef, useCallback } from "react"
import { swrMutate } from "@/lib/swr"

import {
  createGameLevelAction,
  type GameLevelActionResult,
} from "@/features/web/ai-custom/actions/course-game.action"

const initialState: GameLevelActionResult = {}

export function useCreateGameLevel(gameId: string, onSuccess?: () => void) {
  const onSuccessRef = useRef(onSuccess)
  onSuccessRef.current = onSuccess

  const boundAction = useCallback(
    createGameLevelAction.bind(null, gameId),
    [gameId]
  )

  const [state, formAction, isPending] = useActionState(
    boundAction,
    initialState
  )

  useEffect(() => {
    if (state.success) {
      onSuccessRef.current?.()
      swrMutate("/api/course-games")
    }
  }, [state])

  return { state, formAction, isPending }
}
```

- [ ] **Step 3: Fix `use-update-course-game.ts`**

Replace entire content of `dx-web/src/features/web/ai-custom/hooks/use-update-course-game.ts`:

```ts
"use client"

import { useActionState, useState, useEffect, useRef, useCallback } from "react"
import { swrMutate } from "@/lib/swr"

import {
  updateCourseGameAction,
  type UpdateGameResult,
} from "@/features/web/ai-custom/actions/course-game.action"

const initialState: UpdateGameResult = {}

export function useUpdateCourseGame(gameId: string, onSuccess?: () => void) {
  const [coverId, setCoverId] = useState<string | null>(null)
  const onSuccessRef = useRef(onSuccess)
  onSuccessRef.current = onSuccess

  const boundAction = useCallback(
    updateCourseGameAction.bind(null, gameId),
    [gameId]
  )

  const [state, formAction, isPending] = useActionState(
    boundAction,
    initialState
  )

  useEffect(() => {
    if (state.success) {
      onSuccessRef.current?.()
      swrMutate("/api/course-games")
    }
  }, [state])

  return {
    state,
    formAction,
    isPending,
    coverId,
    setCoverId,
  }
}
```

- [ ] **Step 4: Fix `use-game-actions.ts`**

Replace entire content of `dx-web/src/features/web/ai-custom/hooks/use-game-actions.ts`:

```ts
"use client"

import { useState, useTransition } from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { swrMutate } from "@/lib/swr"

import {
  deleteGameAction,
  deleteGameLevelAction,
  publishGameAction,
  withdrawGameAction,
} from "@/features/web/ai-custom/actions/course-game.action"

export function useDeleteGame(gameId: string) {
  const router = useRouter()
  const [isPending, setIsPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function execute() {
    setIsPending(true)
    setError(null)
    const result = await deleteGameAction(gameId)
    if (result.error) {
      setError(result.error)
      setIsPending(false)
    } else {
      await swrMutate("/api/course-games")
      router.replace("/hall/ai-custom")
    }
  }

  return { execute, isPending, error }
}

export function useDeleteGameLevel(gameId: string) {
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)

  function execute(levelId: string) {
    startTransition(async () => {
      const result = await deleteGameLevelAction(gameId, levelId)
      if (result.error) {
        setError(result.error)
      } else {
        await swrMutate("/api/course-games")
      }
    })
  }

  return { execute, isPending, error }
}

export function usePublishGame(gameId: string) {
  const [isPending, startTransition] = useTransition()

  function execute() {
    startTransition(async () => {
      const result = await publishGameAction(gameId)
      if (result.error) {
        toast.error(result.error)
      } else {
        await swrMutate("/api/course-games")
      }
    })
  }

  return { execute, isPending }
}

export function useWithdrawGame(gameId: string) {
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)

  function execute() {
    startTransition(async () => {
      const result = await withdrawGameAction(gameId)
      if (result.error) {
        setError(result.error)
      } else {
        await swrMutate("/api/course-games")
      }
    })
  }

  return { execute, isPending, error }
}
```

- [ ] **Step 5: Verify build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds. All ai-custom mutation hooks now use `swrMutate`.

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/ai-custom/hooks/
git commit -m "fix: replace router.refresh() with swrMutate in ai-custom mutation hooks"
```

---

## Task 4: Replace `router.refresh()` in ai-custom Components

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/add-metadata-dialog.tsx`
- Modify: `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`

- [ ] **Step 1: Fix `add-metadata-dialog.tsx`**

In `dx-web/src/features/web/ai-custom/components/add-metadata-dialog.tsx`:

Replace the import of `useRouter`:
```ts
// Remove:
import { useRouter } from "next/navigation"
// Add:
import { swrMutate } from "@/lib/swr"
```

Remove `const router = useRouter()` on line 53.

Replace `router.refresh()` on line 164 with:
```ts
swrMutate("/api/course-games")
```

- [ ] **Step 2: Fix `level-units-panel.tsx`**

In `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`:

Add import:
```ts
import { swrMutate } from "@/lib/swr"
```

Replace all 3 occurrences of `router.refresh()` (lines 248, 291, 505) with:
```ts
swrMutate("/api/course-games")
```

Also remove the `useRouter` import (line 4) and `const router = useRouter()` declaration — `router` is only used for `.refresh()` in this file.

- [ ] **Step 3: Verify build**

```bash
cd dx-web && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/ai-custom/components/add-metadata-dialog.tsx dx-web/src/features/web/ai-custom/components/level-units-panel.tsx
git commit -m "fix: replace router.refresh() with swrMutate in ai-custom components"
```

---

## Task 5: Migrate ai-custom Detail and Level Pages to SWR

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/course-detail-content.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/ai-custom/[id]/[levelId]/page.tsx`

- [ ] **Step 1: Rewrite `course-detail-content.tsx` with `useSWR`**

Replace the `useEffect` + `useState` data fetching with three `useSWR` calls. The data transformation logic stays in the component.

Replace entire content of `dx-web/src/features/web/ai-custom/components/course-detail-content.tsx`:

```tsx
"use client"

import Link from "next/link"
import useSWR from "swr"
import { ArrowLeft, ChevronRight, ShieldAlert } from "lucide-react"
import { GAME_STATUSES } from "@/consts/game-status"
import { TopActions } from "@/features/web/hall/components/top-actions"
import { GameHeroCard } from "@/features/web/ai-custom/components/game-hero-card"
import { GameLevelsCard } from "@/features/web/ai-custom/components/game-levels-card"
import { GameInfoCard } from "@/features/web/ai-custom/components/game-info-card"
import { PageSpinner } from "@/components/in/page-spinner"

function mapGameDetail(raw: any, categories: any[], presses: any[]) {
  const mapped = {
    ...raw,
    gameCategoryId: raw.gameCategoryId ?? null,
    gamePressId: raw.gamePressId ?? null,
    coverId: raw.coverId ?? null,
    cover: raw.coverUrl ? { url: raw.coverUrl } : null,
    category: null as { name: string } | null,
    press: null as { name: string } | null,
    levels: (raw.levels ?? []).map((l: any) => ({
      ...l,
      _count: { items: 0 },
    })),
    _count: { levels: raw.levels?.length ?? 0, stats: 0 },
  }

  if (raw.gameCategoryId) {
    const cat = categories.find((c: any) => c.id === raw.gameCategoryId)
    if (cat) mapped.category = { name: cat.name }
  }
  if (raw.gamePressId) {
    const press = presses.find((p: any) => p.id === raw.gamePressId)
    if (press) mapped.press = { name: press.name }
  }

  return mapped
}

export function CourseDetailContent({ id }: { id: string }) {
  const { data: raw, error, isLoading: gameLoading } = useSWR(`/api/course-games/${id}`)
  const { data: categories } = useSWR<any[]>("/api/game-categories")
  const { data: presses } = useSWR<any[]>("/api/game-presses")

  if (gameLoading) return <PageSpinner size="lg" />

  if (error || !raw) {
    return (
      <div className="flex flex-col items-center gap-2 py-20 text-center">
        <p className="text-lg font-semibold text-foreground">游戏不存在</p>
        <Link
          href="/hall/ai-custom"
          className="text-sm text-teal-600 hover:underline"
        >
          返回列表
        </Link>
      </div>
    )
  }

  const game = mapGameDetail(raw, categories ?? [], presses ?? [])

  return (
    <>
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link
            href="/hall/ai-custom"
            aria-label="返回"
            className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-border bg-card"
          >
            <ArrowLeft className="h-[18px] w-[18px] text-muted-foreground" />
          </Link>
          <div className="flex items-center gap-2">
            <Link
              href="/hall/ai-custom"
              className="text-sm font-medium text-muted-foreground hover:text-foreground"
            >
              我创建的课程游戏
            </Link>
            <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-sm font-semibold text-foreground">
              {game.name}
            </span>
          </div>
        </div>
        <div className="hidden lg:block">
          <TopActions />
        </div>
      </div>

      {/* Published banner */}
      {game.status === GAME_STATUSES.PUBLISHED && (
        <div className="flex shrink-0 items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm font-medium text-amber-700">
          <ShieldAlert className="h-4 w-4 shrink-0" />
          已发布的游戏不可编辑，如需修改请先撤回游戏
        </div>
      )}

      {/* Hero card */}
      <GameHeroCard
        game={game}
        categories={categories ?? []}
        presses={presses ?? []}
      />

      {/* Bottom row */}
      <div className="flex flex-1 flex-col gap-5 overflow-hidden lg:flex-row">
        <GameLevelsCard
          gameId={game.id}
          levels={game.levels}
          totalLevels={game._count.levels}
          isPublished={game.status === GAME_STATUSES.PUBLISHED}
        />
        <GameInfoCard game={game} />
      </div>
    </>
  )
}
```

- [ ] **Step 2: Rewrite level page with `useSWR`**

Replace entire content of `dx-web/src/app/(web)/hall/(main)/ai-custom/[id]/[levelId]/page.tsx`:

```tsx
"use client"

import { use } from "react"
import useSWR from "swr"
import { BreadcrumbTopBar } from "@/features/web/hall/components/breadcrumb-top-bar"
import { LevelUnitsPanel } from "@/features/web/ai-custom/components/level-units-panel"
import { GAME_STATUSES } from "@/consts/game-status"
import { PageSpinner } from "@/components/in/page-spinner"

export default function CourseGameLevelPage({
  params,
}: {
  params: Promise<{ id: string; levelId: string }>
}) {
  const { id, levelId } = use(params)

  const { data: game, isLoading: gameLoading } = useSWR(`/api/course-games/${id}`)
  const { data: contentGroups, isLoading: contentLoading } = useSWR<any[]>(
    `/api/course-games/${id}/levels/${levelId}/content-items`
  )

  if (gameLoading || contentLoading) return <PageSpinner size="lg" />

  const metas = (contentGroups ?? []).map((group: any) => ({
    id: group.meta.id,
    sourceData: group.meta.sourceData,
    translation: group.meta.translation ?? null,
    sourceFrom: group.meta.sourceFrom,
    sourceType: group.meta.sourceType,
    isBreakDone: group.meta.isBreakDone,
    isItemDone: group.meta.isBreakDone && (group.items?.length ?? 0) > 0,
    order: group.meta.order,
    itemCount: group.items?.length ?? 0,
  }))

  const level = game?.levels?.find((l: any) => l.id === levelId)
  const isPublished = game?.status === GAME_STATUSES.PUBLISHED

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <BreadcrumbTopBar
        backHref={`/hall/ai-custom/${id}`}
        items={[
          { label: "我创建的课程游戏", href: "/hall/ai-custom", maxChars: 10 },
          { label: game?.name ?? id, href: `/hall/ai-custom/${id}`, maxChars: 5 },
          { label: level?.name ?? levelId, maxChars: 5 },
        ]}
      />

      <LevelUnitsPanel gameId={id} levelId={levelId} initialMetas={metas} readOnly={isPublished} />
    </div>
  )
}
```

- [ ] **Step 3: Verify build**

```bash
cd dx-web && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/ai-custom/components/course-detail-content.tsx dx-web/src/app/\(web\)/hall/\(main\)/ai-custom/\[id\]/\[levelId\]/page.tsx
git commit -m "feat: migrate ai-custom detail and level pages to SWR"
```

---

## Task 6: Migrate Public Games Page to SWR

**Files:**
- Create: `dx-web/src/features/web/games/helpers/game-card.ts`
- Modify: `dx-web/src/features/web/games/hooks/use-infinite-public-games.ts`
- Modify: `dx-web/src/app/(web)/hall/(main)/games/page.tsx`
- Modify: `dx-web/src/features/web/games/actions/game.action.ts`

- [ ] **Step 1: Create consolidated `toPublicGameCard` helper**

Create `dx-web/src/features/web/games/helpers/game-card.ts`:

```ts
import type { PublicGameCard } from "@/features/web/games/actions/game.action"

/** Map Go API flat GameCardData to the nested PublicGameCard shape */
export function toPublicGameCard(item: any): PublicGameCard {
  return {
    id: item.id,
    name: item.name,
    description: item.description ?? null,
    mode: item.mode,
    createdAt: new Date(item.createdAt),
    cover: item.coverUrl ? { url: item.coverUrl } : null,
    user: item.author ? { username: item.author } : null,
    category: item.categoryName ? { name: item.categoryName } : null,
    _count: { levels: item.levelCount ?? 0 },
  }
}
```

- [ ] **Step 2: Rewrite `use-infinite-public-games.ts` with `useSWRInfinite`**

Replace entire content of `dx-web/src/features/web/games/hooks/use-infinite-public-games.ts`:

```ts
"use client"

import { useRef, useEffect, useCallback } from "react"
import useSWRInfinite from "swr/infinite"
import { toPublicGameCard } from "@/features/web/games/helpers/game-card"
import type { PublicGameCard } from "@/features/web/games/actions/game.action"

type Filters = {
  categoryIds?: string[]
  pressId?: string
  mode?: string
}

export function useInfinitePublicGames(filters: Filters = {}) {
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const getKey = (pageIndex: number, previousPageData: any) => {
    if (previousPageData && !previousPageData.hasMore) return null
    const cursor = previousPageData?.nextCursor
    const params = new URLSearchParams()
    if (cursor) params.set("cursor", cursor)
    if (filters.categoryIds?.length) params.set("categoryIds", filters.categoryIds.join(","))
    if (filters.pressId) params.set("pressId", filters.pressId)
    if (filters.mode) params.set("mode", filters.mode)
    const qs = params.toString()
    return `/api/games${qs ? `?${qs}` : ""}`
  }

  const { data, size, setSize, isLoading, isValidating, mutate } =
    useSWRInfinite(getKey)

  const games: PublicGameCard[] = data?.flatMap((page) =>
    (page.items ?? []).map(toPublicGameCard)
  ) ?? []
  const hasMore = data?.[data.length - 1]?.hasMore ?? false

  const loadMore = useCallback(() => {
    if (!isValidating && hasMore) setSize(size + 1)
  }, [isValidating, hasMore, size, setSize])

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) loadMore()
      },
      { rootMargin: "200px" }
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore])

  return { games, isLoading, isValidating, hasMore, sentinelRef, mutate }
}
```

- [ ] **Step 3: Simplify `games/page.tsx`**

Replace entire content of `dx-web/src/app/(web)/hall/(main)/games/page.tsx`:

```tsx
"use client"

import useSWR from "swr"
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar"
import { GamesPageContent } from "@/features/web/games/components/games-page-content"
import { PageSpinner } from "@/components/in/page-spinner"

export default function HallGamesPage() {
  const { data: categories, isLoading: catLoading } = useSWR<any[]>("/api/game-categories")
  const { data: presses, isLoading: pressLoading } = useSWR<any[]>("/api/game-presses")

  const isLoading = catLoading || pressLoading

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <GreetingTopBar
        title="课程游戏"
        subtitle="选择一个游戏模式，边玩边学英语！"
      />
      {isLoading ? (
        <PageSpinner size="lg" />
      ) : (
        <GamesPageContent
          categories={categories ?? []}
          presses={presses ?? []}
        />
      )}
    </div>
  )
}
```

- [ ] **Step 4: Rewrite `GamesPageContent` to use SWR hook directly**

Replace entire content of `dx-web/src/features/web/games/components/games-page-content.tsx`:

```tsx
"use client"

import { useState } from "react"
import { Gamepad2 } from "lucide-react"
import { PageSpinner } from "@/components/in/page-spinner"
import { FilterSection } from "@/features/web/games/components/filter-section"
import { GameCard } from "@/features/web/games/components/game-card"
import { useInfinitePublicGames } from "@/features/web/games/hooks/use-infinite-public-games"

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean }
type PressOption = { id: string; name: string }
type Filters = { categoryIds?: string[]; pressId?: string; mode?: string }

type GamesPageContentProps = {
  categories: CategoryOption[]
  presses: PressOption[]
}

export function GamesPageContent({ categories, presses }: GamesPageContentProps) {
  const [filters, setFilters] = useState<Filters>({})
  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfinitePublicGames(filters)

  return (
    <>
      <FilterSection
        categories={categories}
        presses={presses}
        filters={filters}
        onFiltersChange={setFilters}
      />

      {isLoading && <PageSpinner size="lg" />}

      {!isLoading && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
          {games.map((game) => (
            <GameCard key={game.id} game={game} />
          ))}
        </div>
      )}

      {isValidating && !isLoading && <PageSpinner size="sm" />}

      {!isLoading && !isValidating && games.length === 0 && (
        <div className="flex flex-col items-center gap-2 py-12 text-center">
          <Gamepad2 className="h-10 w-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">暂无游戏</p>
        </div>
      )}

      {hasMore && <div ref={sentinelRef} className="h-1" />}
    </>
  )
}
```

- [ ] **Step 5: Clean up `game.action.ts`**

In `dx-web/src/features/web/games/actions/game.action.ts`:
- Remove `fetchPublicGamesAction` function and its `FetchPublicGamesResult` type
- Remove the local `toPublicGameCard` function — now in `helpers/game-card.ts`
- Keep `PublicGameCard` type export

- [ ] **Step 6: Verify build**

```bash
cd dx-web && npm run build
```

- [ ] **Step 7: Commit**

```bash
git add dx-web/src/features/web/games/ dx-web/src/app/\(web\)/hall/\(main\)/games/page.tsx
git commit -m "feat: migrate public games page to SWR with useSWRInfinite"
```

---

## Task 7: Migrate Notices

**Files:**
- Modify: `dx-web/src/features/web/notice/components/mark-notices-read.tsx`

- [ ] **Step 1: Fix `mark-notices-read.tsx`**

Replace entire content of `dx-web/src/features/web/notice/components/mark-notices-read.tsx`:

```tsx
"use client"

import { useEffect } from "react"
import { swrMutate } from "@/lib/swr"
import { markNoticesReadAction } from "@/features/web/notice/actions/notice.action"

/** Marks notices as read on mount and invalidates SWR cache */
export function MarkNoticesRead() {
  useEffect(() => {
    markNoticesReadAction().then(() => swrMutate("/api/notices"))
  }, [])

  return null
}
```

- [ ] **Step 2: Verify build**

```bash
cd dx-web && npm run build
```

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/notice/components/mark-notices-read.tsx
git commit -m "fix: replace router.refresh() with swrMutate in notices"
```

---

## Task 8: Final Verification

- [ ] **Step 1: Grep for remaining `router.refresh()` calls**

```bash
cd dx-web && grep -r "router.refresh()" src/ --include="*.ts" --include="*.tsx"
```

Expected: Zero results (or only comments).

- [ ] **Step 2: Run full build**

```bash
cd dx-web && npm run build
```

Expected: Build succeeds with no errors.

- [ ] **Step 3: Run lint**

```bash
cd dx-web && npm run lint
```

Expected: No new lint errors.

- [ ] **Step 4: Manual smoke test**

Start the dev server and verify:
1. Ai-custom list page loads with spinner, then shows games
2. Create a new course game — it appears in the list immediately (no page refresh)
3. Delete a game — list updates immediately
4. Publish/withdraw — status updates immediately
5. Games page loads with spinner, infinite scroll works
6. Filter changes on both pages work correctly

- [ ] **Step 5: Commit (if any final fixes needed)**

```bash
git add -A
git commit -m "chore: final verification — zero router.refresh() remaining"
```
