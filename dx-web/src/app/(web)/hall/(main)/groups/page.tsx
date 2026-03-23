import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { GroupListContent } from "@/features/web/groups/components/group-list-content";

export default function GroupsPage() {
  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="学习群"
        subtitle="浏览并加入学习群组，与小伙伴一起进步"
        searchPlaceholder="搜索群组..."
      />
      <GroupListContent />
    </div>
  );
}
