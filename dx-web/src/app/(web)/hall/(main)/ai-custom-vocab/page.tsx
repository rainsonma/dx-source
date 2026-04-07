import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { AiCustomVocabGrid } from "@/features/web/ai-custom-vocab/components/ai-custom-vocab-grid"

export default function AiCustomVocabPage() {
  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="AI 词汇工坊"
        subtitle="AI 驱动的个性化词汇练习游戏"
      />
      <AiCustomVocabGrid />
    </div>
  )
}
