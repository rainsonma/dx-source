"use client"

import type { ReactNode } from "react"
import { SWRConfig, useSWRConfig } from "swr"
import { mutate } from "swr"
import { apiClient } from "@/lib/api-client"

export const swrFetcher = (url: string) =>
  apiClient.get(url).then((res) => {
    if (res.code !== 0) throw new Error(res.message)
    return res.data
  })

// Module-level reference to SWR cache, captured when SWRProvider mounts.
// Allows swrMutate to iterate cache keys including $inf$ keys that
// global mutate(filterFn) intentionally skips.
let _cache: Map<string, any> | null = null

function CacheCapture() {
  const { cache } = useSWRConfig()
  _cache = cache as Map<string, any>
  return null
}

/**
 * Invalidate all SWR cache entries (including useSWRInfinite) whose
 * internal key contains any of the given prefixes. Calls mutate(exactKey)
 * per match to bypass the $inf$ filter in global mutate(filterFn).
 */
export async function swrMutate(...prefixes: string[]) {
  if (!_cache) return
  const keys: string[] = []
  for (const k of _cache.keys()) {
    if (typeof k === "string" && prefixes.some((p) => k.includes(p))) {
      keys.push(k)
    }
  }
  await Promise.all(keys.map((k) => mutate(k)))
}

export function SWRProvider({ children }: { children: ReactNode }) {
  return (
    <SWRConfig value={{ fetcher: swrFetcher }}>
      <CacheCapture />
      {children}
    </SWRConfig>
  )
}
