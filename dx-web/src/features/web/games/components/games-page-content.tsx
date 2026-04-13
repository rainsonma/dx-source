"use client"

import { useState } from "react"
import { Gamepad2 } from "lucide-react"
import { PageSpinner } from "@/components/in/page-spinner"
import { FilterSection } from "@/features/web/games/components/filter-section"
import { GameCard } from "@/features/web/games/components/game-card"
import { useInfinitePublicGames } from "@/features/web/games/hooks/use-infinite-public-games"

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean }
type Filters = { categoryIds?: string[]; pressId?: string; mode?: string }

type GamesPageContentProps = {
  categories: CategoryOption[]
}

export function GamesPageContent({ categories }: GamesPageContentProps) {
  const [filters, setFilters] = useState<Filters>({})
  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfinitePublicGames(filters)

  return (
    <>
      <FilterSection
        categories={categories}
        presses={[]}
        filters={filters}
        onFiltersChange={setFilters}
        showPresses={false}
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
