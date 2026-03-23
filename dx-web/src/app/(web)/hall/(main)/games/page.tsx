"use client"

import useSWR from "swr"
import { GreetingTopBar } from "@/features/web/hall/components/greeting-top-bar"
import { GamesPageContent } from "@/features/web/games/components/games-page-content"
import { PageSpinner } from "@/components/in/page-spinner"

export default function HallGamesPage() {
  const { data: categories, isLoading: catLoading } = useSWR<any[]>("/api/game-categories")
  const { data: presses, isLoading: pressLoading } = useSWR<any[]>("/api/game-presses")

  const isLoading = catLoading || pressLoading

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <GreetingTopBar
        title="课程游戏"
        subtitle="选择一个游戏模式，边玩边学英语！"
      />
      {isLoading ? (
        <PageSpinner size="lg" />
      ) : (
        <GamesPageContent
          categories={categories ?? []}
          presses={presses ?? []}
        />
      )}
    </div>
  )
}
