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
