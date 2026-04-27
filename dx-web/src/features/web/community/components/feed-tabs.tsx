"use client"

import { TabPill } from "@/components/in/tab-pill"
import type { FeedTab } from "../types/post"

const TABS: { key: FeedTab; label: string }[] = [
  { key: "latest", label: "最新" },
  { key: "hot", label: "热门" },
  { key: "following", label: "关注" },
  { key: "bookmarked", label: "书签" },
]

interface FeedTabsProps {
  active: FeedTab
  onChange: (tab: FeedTab) => void
}

export function FeedTabs({ active, onChange }: FeedTabsProps) {
  return (
    <div className="flex items-center gap-2">
      {TABS.map((t) => (
        <TabPill
          key={t.key}
          label={t.label}
          active={active === t.key}
          onClick={() => onChange(t.key)}
        />
      ))}
    </div>
  )
}
