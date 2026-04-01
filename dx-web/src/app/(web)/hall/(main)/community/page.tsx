import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { CommunityFeed } from "@/features/web/community/components/community-feed"

export default function CommunityPage() {
  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="斗学社"
        subtitle="分享学习心得，与学友互动交流"
      />
      <CommunityFeed />
    </div>
  )
}
