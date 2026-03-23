"use client"

import Link from "next/link"
import useSWR from "swr"
import { ArrowLeft, ChevronRight, ShieldAlert } from "lucide-react"
import { GAME_STATUSES } from "@/consts/game-status"
import { TopActions } from "@/features/web/hall/components/top-actions"
import { GameHeroCard } from "@/features/web/ai-custom/components/game-hero-card"
import { GameLevelsCard } from "@/features/web/ai-custom/components/game-levels-card"
import { GameInfoCard } from "@/features/web/ai-custom/components/game-info-card"
import { PageSpinner } from "@/components/in/page-spinner"

function mapGameDetail(raw: any, categories: any[], presses: any[]) {
  const mapped = {
    ...raw,
    gameCategoryId: raw.gameCategoryId ?? null,
    gamePressId: raw.gamePressId ?? null,
    coverId: raw.coverId ?? null,
    cover: raw.coverUrl ? { url: raw.coverUrl } : null,
    category: null as { name: string } | null,
    press: null as { name: string } | null,
    levels: (raw.levels ?? []).map((l: any) => ({
      ...l,
      _count: { items: 0 },
    })),
    _count: { levels: raw.levels?.length ?? 0, stats: 0 },
  }

  if (raw.gameCategoryId) {
    const cat = categories.find((c: any) => c.id === raw.gameCategoryId)
    if (cat) mapped.category = { name: cat.name }
  }
  if (raw.gamePressId) {
    const press = presses.find((p: any) => p.id === raw.gamePressId)
    if (press) mapped.press = { name: press.name }
  }

  return mapped
}

export function CourseDetailContent({ id }: { id: string }) {
  const { data: raw, error, isLoading: gameLoading } = useSWR(`/api/course-games/${id}`)
  const { data: categories } = useSWR<any[]>("/api/game-categories")
  const { data: presses } = useSWR<any[]>("/api/game-presses")

  if (gameLoading) return <PageSpinner size="lg" />

  if (error || !raw) {
    return (
      <div className="flex flex-col items-center gap-2 py-20 text-center">
        <p className="text-lg font-semibold text-foreground">游戏不存在</p>
        <Link
          href="/hall/ai-custom"
          className="text-sm text-teal-600 hover:underline"
        >
          返回列表
        </Link>
      </div>
    )
  }

  const game = mapGameDetail(raw, categories ?? [], presses ?? [])

  return (
    <>
      {/* Top bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link
            href="/hall/ai-custom"
            aria-label="返回"
            className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-border bg-card"
          >
            <ArrowLeft className="h-[18px] w-[18px] text-muted-foreground" />
          </Link>
          <div className="flex items-center gap-2">
            <Link
              href="/hall/ai-custom"
              className="text-sm font-medium text-muted-foreground hover:text-foreground"
            >
              我创建的课程游戏
            </Link>
            <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-sm font-semibold text-foreground">
              {game.name}
            </span>
          </div>
        </div>
        <div className="hidden lg:block">
          <TopActions />
        </div>
      </div>

      {/* Published banner */}
      {game.status === GAME_STATUSES.PUBLISHED && (
        <div className="flex shrink-0 items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm font-medium text-amber-700">
          <ShieldAlert className="h-4 w-4 shrink-0" />
          已发布的游戏不可编辑，如需修改请先撤回游戏
        </div>
      )}

      {/* Hero card */}
      <GameHeroCard
        game={game}
        categories={categories ?? []}
        presses={presses ?? []}
      />

      {/* Bottom row */}
      <div className="flex flex-1 flex-col gap-5 overflow-hidden lg:flex-row">
        <GameLevelsCard
          gameId={game.id}
          levels={game.levels}
          totalLevels={game._count.levels}
          isPublished={game.status === GAME_STATUSES.PUBLISHED}
        />
        <GameInfoCard game={game} />
      </div>
    </>
  )
}
