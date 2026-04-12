"use client"

import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu"
import { GroupListContent } from "@/features/web/groups/components/group-list-content"

export default function GroupsPage() {
  const menu = useHallMenuItem("/hall/groups")

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? ""}
        subtitle={menu?.subtitle ?? ""}
        searchPlaceholder="搜索群组..."
      />
      <GroupListContent />
    </div>
  )
}
