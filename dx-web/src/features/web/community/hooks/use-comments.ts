"use client"

import { useRef, useCallback, useEffect } from "react"
import useSWRInfinite from "swr/infinite"
import type { CursorPaginated } from "@/lib/api-client"
import type { CommentWithReplies } from "../types/post"

export function useComments(postId: string) {
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const getKey = (
    pageIndex: number,
    previousPageData: CursorPaginated<CommentWithReplies> | null
  ) => {
    if (previousPageData && !previousPageData.hasMore) return null
    const params = new URLSearchParams()
    if (previousPageData?.nextCursor) {
      params.set("cursor", previousPageData.nextCursor)
    }
    const qs = params.toString()
    return `/api/posts/${postId}/comments${qs ? `?${qs}` : ""}`
  }

  const { data, size, setSize, isLoading, isValidating, mutate } =
    useSWRInfinite(getKey)

  const comments: CommentWithReplies[] = data?.flatMap(
    (page: CursorPaginated<CommentWithReplies>) => page.items ?? []
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

  return { comments, isLoading, isValidating, hasMore, sentinelRef, mutate }
}
