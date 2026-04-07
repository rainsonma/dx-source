"use client"

import { use } from "react"
import useSWR from "swr"
import { BreadcrumbTopBar } from "@/features/web/hall/components/breadcrumb-top-bar"
import { LevelUnitsPanel } from "@/features/web/ai-custom-vocab/components/level-units-panel"
import { GAME_STATUSES } from "@/consts/game-status"
import { PageSpinner } from "@/components/in/page-spinner"
import type { GameMode } from "@/consts/game-mode"

export default function CourseGameLevelPage({
  params,
}: {
  params: Promise<{ id: string; levelId: string }>
}) {
  const { id, levelId } = use(params)

  const { data: game, isLoading: gameLoading } = useSWR(`/api/course-games/${id}`)
  type ContentGroupItem = { items: unknown[] | null; contentType: string };
  type ContentGroup = {
    meta: {
      id: string;
      sourceData: string;
      translation: string | null;
      sourceFrom: string;
      sourceType: string;
      isBreakDone: boolean;
      order: number;
    };
    items?: ContentGroupItem[];
  };

  const { data: contentGroups, isLoading: contentLoading } = useSWR<ContentGroup[]>(
    `/api/course-games/${id}/levels/${levelId}/content-items`
  )

  if (gameLoading || contentLoading) return <PageSpinner size="lg" />

  const metas = (contentGroups ?? []).map((group) => ({
    id: group.meta.id,
    sourceData: group.meta.sourceData,
    translation: group.meta.translation ?? null,
    sourceFrom: group.meta.sourceFrom,
    sourceType: group.meta.sourceType,
    isBreakDone: group.meta.isBreakDone,
    isItemDone: group.meta.isBreakDone && (group.items?.length ?? 0) > 0
      && group.items!.every((item) => item.items !== null),
    order: group.meta.order,
    itemCount: group.items?.length ?? 0,
  }))

  const level = game?.levels?.find((l: { id: string }) => l.id === levelId)
  const isPublished = game?.status === GAME_STATUSES.PUBLISHED

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <BreadcrumbTopBar
        backHref={`/hall/ai-custom-vocab/${id}`}
        items={[
          { label: "我创建的词汇游戏", href: "/hall/ai-custom-vocab", maxChars: 10 },
          { label: game?.name ?? id, href: `/hall/ai-custom-vocab/${id}`, maxChars: 5 },
          { label: level?.name ?? levelId, maxChars: 5 },
        ]}
      />

      <LevelUnitsPanel
        gameId={id}
        levelId={levelId}
        gameMode={(game?.mode ?? "vocab-match") as GameMode}
        initialMetas={metas}
        readOnly={isPublished}
      />
    </div>
  )
}
