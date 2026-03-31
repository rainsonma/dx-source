import { CommunityFeed } from "@/features/web/community/components/community-feed"

export default function CommunityPage() {
  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold text-foreground">斗学社</h1>
        <p className="text-sm text-muted-foreground">
          分享学习心得，与学友互动交流
        </p>
      </div>
      <CommunityFeed />
    </div>
  )
}
