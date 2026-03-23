import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { CommunityFeed } from "@/features/web/community/components/community-feed";

export default function CommunityPage() {
  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <PageTopBar
        title="斗学社"
        subtitle="分享学习心得，与学友互动交流"
        searchPlaceholder="搜索帖子..."
      />
      <CommunityFeed />
    </div>
  );
}
