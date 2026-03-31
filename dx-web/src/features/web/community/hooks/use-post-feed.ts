"use client"

import { useRef, useCallback, useEffect } from "react"
import useSWRInfinite from "swr/infinite"
import type { CursorPaginated } from "@/lib/api-client"
import type { Post, FeedTab } from "../types/post"

export function usePostFeed(tab: FeedTab) {
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const getKey = (
    pageIndex: number,
    previousPageData: CursorPaginated<Post> | null
  ) => {
    if (previousPageData && !previousPageData.hasMore) return null
    const params = new URLSearchParams()
    params.set("tab", tab)
    if (previousPageData?.nextCursor) {
      params.set("cursor", previousPageData.nextCursor)
    }
    return `/api/posts?${params}`
  }

  const { data, size, setSize, isLoading, isValidating, mutate } =
    useSWRInfinite(getKey)

  const posts: Post[] = data?.flatMap((page: CursorPaginated<Post>) => page.items ?? []) ?? []
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

  return { posts, isLoading, isValidating, hasMore, sentinelRef, mutate }
}
