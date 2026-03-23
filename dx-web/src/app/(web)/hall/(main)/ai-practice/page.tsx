import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { AiTopicGrid } from "@/features/web/ai-practice/components/ai-topic-grid";

export default function AiPracticePage() {
  return (
    <div className="flex h-full flex-col gap-6 px-4 py-5 md:px-8 md:py-7">
      <PageTopBar
        title="AI 随心练"
        subtitle="AI 驱动的个性化英语场景练习"
      />
      <AiTopicGrid />
    </div>
  );
}
