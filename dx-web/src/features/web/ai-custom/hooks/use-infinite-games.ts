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
