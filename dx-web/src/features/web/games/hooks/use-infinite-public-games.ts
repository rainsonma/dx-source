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

  const getKey = (pageIndex: number, previousPageData: { hasMore: boolean; nextCursor?: string; items?: unknown[] } | null) => {
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
