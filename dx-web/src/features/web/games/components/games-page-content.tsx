"use client"

import { useMemo, useState } from "react"
import { Gamepad2 } from "lucide-react"
import { PageSpinner } from "@/components/in/page-spinner"
import { FilterSection } from "@/features/web/games/components/filter-section"
import { GameCard } from "@/features/web/games/components/game-card"
import { useInfinitePublicGames } from "@/features/web/games/hooks/use-infinite-public-games"

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean }
type PressOption = { id: string; name: string }
type Filters = { categoryIds?: string[]; pressId?: string; mode?: string }

type GamesPageContentProps = {
  categories: CategoryOption[]
  presses: PressOption[]
}

const SYNC_CATEGORY_NAME = "同步练习"

export function GamesPageContent({ categories, presses }: GamesPageContentProps) {
  const [filters, setFilters] = useState<Filters>({})
  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfinitePublicGames(filters)

  const syncRootId = useMemo(
    () => categories.find((c) => c.name === SYNC_CATEGORY_NAME && c.depth === 0)?.id,
    [categories],
  )

  const syncChildIds = useMemo(() => {
    if (!syncRootId) return new Set<string>()
    const idx = categories.findIndex((c) => c.id === syncRootId)
    const ids = new Set<string>([syncRootId])
    for (let i = idx + 1; i < categories.length; i++) {
      if (categories[i].depth === 0) break
      ids.add(categories[i].id)
    }
    return ids
  }, [categories, syncRootId])

  const showPresses = useMemo(() => {
    if (!syncRootId || !filters.categoryIds?.length) return false
    return filters.categoryIds.some((id) => syncChildIds.has(id))
  }, [syncRootId, filters.categoryIds, syncChildIds])

  function handleFiltersChange(next: Filters) {
    if (next.pressId && syncRootId) {
      const nextInSync = next.categoryIds?.some((id) => syncChildIds.has(id)) ?? false
      if (!nextInSync) {
        setFilters({ ...next, pressId: undefined })
        return
      }
    }
    setFilters(next)
  }

  return (
    <>
      <FilterSection
        categories={categories}
        presses={presses}
        filters={filters}
        onFiltersChange={handleFiltersChange}
        showPresses={showPresses}
      />

      {isLoading && <PageSpinner size="lg" />}

      {!isLoading && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
          {games.map((game) => (
            <GameCard key={game.id} game={game} />
          ))}
        </div>
      )}

      {isValidating && !isLoading && <PageSpinner size="sm" />}

      {!isLoading && !isValidating && games.length === 0 && (
        <div className="flex flex-col items-center gap-2 py-12 text-center">
          <Gamepad2 className="h-10 w-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">暂无游戏</p>
        </div>
      )}

      {hasMore && <div ref={sentinelRef} className="h-1" />}
    </>
  )
}
