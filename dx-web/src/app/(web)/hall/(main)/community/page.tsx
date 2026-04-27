"use client"

import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
import { useWS } from "@/providers/websocket-provider"
import { CommunityFeed } from "@/features/web/community/components/community-feed"

export default function CommunityPage() {
  const menu = useHallMenuItem("/hall/community")
  const { userId } = useWS()

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
      />
      <CommunityFeed currentUserId={userId ?? undefined} />
    </div>
  )
}
