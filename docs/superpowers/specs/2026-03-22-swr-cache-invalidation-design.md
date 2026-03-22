# SWR Cache Invalidation Design

## Problem

All dynamic pages in dx-web are `"use client"` components that fetch data via `apiClient` in `useEffect`. After mutations (create, update, delete), `router.refresh()` is called across 7 files (11 active call sites) to refresh the data — but this is a no-op because `router.refresh()` only re-runs server components, and there are no server components fetching data.

Result: users must manually refresh the browser to see changes (e.g., a newly created course game doesn't appear in the list).

## Decision

Adopt SWR (~4KB gzip) as the client-side data fetching and cache layer. Replace manual `useEffect` + `useState` fetching with `useSWR` hooks, and replace all `router.refresh()` calls with SWR cache invalidation via `mutate()`.

## Design

### 1. SWR Foundation

**Fetcher** — wraps `apiClient.get`, unwraps the response envelope:

```ts
// lib/swr.ts
import { apiClient } from '@/lib/api-client'
import { SWRConfig } from 'swr'

export const swrFetcher = (url: string) =>
  apiClient.get(url).then(res => {
    if (res.code !== 0) throw new Error(res.message)
    return res.data
  })
```

**Provider** — global SWR config placed inside `AuthGuard` (which is already a `"use client"` component wrapping all `/hall/*` children):

```tsx
// Inside AuthGuard's render, wrap children with SWRConfig
<SWRConfig value={{ fetcher: swrFetcher }}>
  {children}
</SWRConfig>
```

With the global fetcher, components use SWR without specifying a fetcher each time:

```ts
const { data, isLoading } = useSWR('/api/course-games/counts')
```

**Auth error handling:** The fetcher delegates to `apiClient.get()` which already handles 401 responses internally (token refresh, redirect to signin). SWR's error state only fires for non-auth API errors (e.g., validation failures, server errors).

**SWR key convention:** Keys must always be the exact API path including query parameters (e.g., `/api/course-games?status=draft&cursor=abc`). This enables prefix-based invalidation via `swrMutate('/api/course-games')` matching all variants.

**Default behaviors kept:** `revalidateOnFocus` and `revalidateOnReconnect` are enabled by default — users returning to the tab or reconnecting see fresh data automatically. These can be disabled per-hook if needed for static-ish data.

### 2. Cache Invalidation Helper

A centralized helper that invalidates all SWR keys matching a prefix. Replaces every `router.refresh()` call in the codebase.

```ts
// lib/swr.ts
import { mutate } from 'swr'

export async function swrMutate(...keys: string[]) {
  await Promise.all(keys.map(key => mutate(
    k => typeof k === 'string' && k.startsWith(key),
    undefined,
    { revalidate: true }
  )))
}
```

Usage after any mutation:

```ts
await swrMutate('/api/course-games')
```

This invalidates all keys starting with `/api/course-games` — the list, counts, detail pages, and infinite scroll pages — in one call.

### 3. Infinite Scroll with `useSWRInfinite`

Replaces custom `useInfiniteGames` and `useInfinitePublicGames` hooks.

```ts
import useSWRInfinite from 'swr/infinite'

export function useInfiniteGames(status?: string) {
  const getKey = (pageIndex: number, previousPageData: any) => {
    if (previousPageData && !previousPageData.hasMore) return null
    const cursor = previousPageData?.nextCursor
    const params = new URLSearchParams()
    if (status) params.set('status', status)
    if (cursor) params.set('cursor', cursor)
    return `/api/course-games?${params}`
  }

  const { data, size, setSize, isLoading, isValidating, mutate } =
    useSWRInfinite(getKey)

  const games = data?.flatMap(page => page.items) ?? []
  const hasMore = data?.[data.length - 1]?.hasMore ?? false
  const loadMore = () => setSize(size + 1)

  return { games, isLoading, isValidating, hasMore, loadMore, mutate }
}
```

IntersectionObserver sentinel remains for triggering `loadMore()`.

Filter changes reset the hook by changing the `status` parameter, which changes the SWR key and triggers a fresh fetch.

### 4. Mutation Pattern

All mutation hooks replace `router.refresh()` with `swrMutate()`:

```ts
// BEFORE
useEffect(() => {
  if (state.success) {
    onSuccessRef.current?.()
    router.refresh()
  }
}, [state, router])

// AFTER
useEffect(() => {
  if (state.success) {
    onSuccessRef.current?.()
    swrMutate('/api/course-games')
  }
}, [state])
```

**Delete + navigate pattern:** `useDeleteGame` calls `router.replace('/hall/ai-custom')` after deletion. Call `swrMutate` before navigating so the list cache is already invalidated when the list page renders:

```ts
// BEFORE
await deleteGameAction(gameId)
router.replace('/hall/ai-custom')

// AFTER
await deleteGameAction(gameId)
await swrMutate('/api/course-games')
router.replace('/hall/ai-custom')
```

**Key invalidation mapping:**

| Mutation | Invalidated keys |
|----------|-----------------|
| Create / Delete game | `/api/course-games` |
| Update / Publish / Withdraw game | `/api/course-games` |
| Create / Delete level | `/api/course-games` |
| Save metadata, content items | `/api/course-games` |
| Mark notices read | `/api/notices` |

### 5. Loading Spinner

A shared `PageSpinner` component with context-appropriate sizing:

```tsx
// components/in/page-spinner.tsx
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

type PageSpinnerProps = {
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

export function PageSpinner({ size = 'md', className }: PageSpinnerProps) {
  const sizeMap = { sm: 'h-4 w-4', md: 'h-5 w-5', lg: 'h-6 w-6' }
  const paddingMap = { sm: 'py-4', md: 'py-12', lg: 'py-20' }
  return (
    <div className={cn('flex items-center justify-center', paddingMap[size], className)}>
      <Loader2 className={cn('animate-spin text-muted-foreground', sizeMap[size])} />
    </div>
  )
}
```

Usage contexts:
- **Full page first load** (`isLoading`, no data yet): `<PageSpinner size="lg" />`
- **Infinite scroll bottom**: `<PageSpinner size="md" />`
- **Inline refetch indicator** (`isValidating`, data already shown): `<PageSpinner size="sm" />`

### 6. Two Loading States

SWR distinguishes:
- **`isLoading`** — first load, no cached data. Show full spinner, hide content.
- **`isValidating`** — refetching while cached data is displayed. Show inline indicator, keep content visible.

This gives instant feedback on mutations (cached data stays visible) while the refetch happens in the background.

## Scope of Changes

### New files
- `dx-web/src/lib/swr.ts` — fetcher, `swrMutate` helper, SWR config export
- `dx-web/src/components/in/page-spinner.tsx` — shared loading spinner

### New dependency
- `swr` (~4KB gzip)

### Modified files

**ai-custom (primary fix):**
- `app/(web)/hall/(main)/ai-custom/page.tsx` — `useEffect` fetching replaced with `useSWR`
- `app/(web)/hall/(main)/ai-custom/[id]/page.tsx` — same (detail page)
- `app/(web)/hall/(main)/ai-custom/[id]/[levelId]/page.tsx` — same (level page)
- `features/web/ai-custom/hooks/use-infinite-games.ts` — rewritten with `useSWRInfinite`
- `features/web/ai-custom/hooks/use-create-course-game.ts` — `swrMutate` replaces `router.refresh()`
- `features/web/ai-custom/hooks/use-create-game-level.ts` — same
- `features/web/ai-custom/hooks/use-update-course-game.ts` — same
- `features/web/ai-custom/hooks/use-game-actions.ts` — same (delete + navigate, publish, withdraw, deleteLevel)
- `features/web/ai-custom/components/add-metadata-dialog.tsx` — same
- `features/web/ai-custom/components/level-units-panel.tsx` — same (3 occurrences)
- `features/web/ai-custom/components/ai-custom-grid.tsx` — remove `initialGames` prop pattern, use SWR loading states
- `features/web/ai-custom/components/course-detail-content.tsx` — multi-fetch pattern (game detail + categories + presses) replaced with three separate `useSWR` calls combined in the component

**Other pages with infinite scroll:**
- `features/web/games/hooks/use-infinite-public-games.ts` — rewritten with `useSWRInfinite`
- `app/(web)/hall/(main)/games/page.tsx` — replace `useEffect` with `useSWR`

**Notices:**
- `features/web/notice/components/mark-notices-read.tsx` — `swrMutate` replaces `router.refresh()`

**Data transformation note:** Pages that transform API responses (e.g., public games `toPublicGameCard`) keep their transformation in the component layer after receiving raw data from SWR. The fetcher remains generic.

### Deleted code (replaced by SWR)
- `fetchUserGamesAction` in `course-game.action.ts` — SWR fetches directly
- `fetchPublicGamesAction` in `game.action.ts` — SWR fetches directly (keep `toPublicGameCard` transform, consolidate the duplicated copy from `games/page.tsx` into a shared helper)

### Untouched
- All server component pages (home, auth, landing)
- `api-client.ts` — stays as-is, SWR wraps it
- Backend (dx-api) — no changes
- SSE streaming helpers — not a fetch pattern

## Migration Strategy

Incremental, one feature at a time:

1. **Phase 1:** Install SWR, create `lib/swr.ts`, add `PageSpinner`, wire SWR provider
2. **Phase 2:** Migrate ai-custom (fixes the reported bug)
3. **Phase 3:** Migrate games (public games infinite scroll)
4. **Phase 4:** Migrate notices
5. **Phase 5:** Migrate remaining pages as needed

Each phase is independently deployable.
