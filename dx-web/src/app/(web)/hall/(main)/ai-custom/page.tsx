import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { AiCustomGrid } from "@/features/web/ai-custom/components/ai-custom-grid"

export default function AiCustomPage() {
  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="AI 随心配"
        subtitle="AI 驱动的个性化英语练习游戏"
      />
      <AiCustomGrid />
    </div>
  )
}
