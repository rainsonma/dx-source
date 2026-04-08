"use client"

import { useState, useEffect } from "react"
import useSWR from "swr"
import { Puzzle, Plus } from "lucide-react"
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { VisuallyHidden } from "@radix-ui/react-visually-hidden"
import { PageSpinner } from "@/components/in/page-spinner"
import { CreateCourseForm } from "@/features/web/ai-custom/components/create-course-form"
import { GameCardItem } from "@/features/web/ai-custom/components/game-card-item"
import { useInfiniteGames } from "@/features/web/ai-custom/hooks/use-infinite-games"
import { apiClient } from "@/lib/api-client"
import { isVipActive } from "@/lib/vip"
import type { UserGrade } from "@/consts/user-grade"
import { UpgradeDialog } from "@/features/web/games/components/upgrade-dialog"

type GameCounts = { all: number; published: number; withdraw: number; draft: number }
type StatusFilter = "all" | "published" | "withdraw" | "draft"

const filterKeys: { key: StatusFilter; label: string; countKey: keyof GameCounts }[] = [
  { key: "all", label: "全部", countKey: "all" },
  { key: "published", label: "已发布", countKey: "published" },
  { key: "withdraw", label: "已撤回", countKey: "withdraw" },
  { key: "draft", label: "未发布", countKey: "draft" },
]

export function AiCustomGrid() {
  const [open, setOpen] = useState(false)
  const [activeFilter, setActiveFilter] = useState<StatusFilter>("all")
  const [isVip, setIsVip] = useState(false)
  const [upgradeOpen, setUpgradeOpen] = useState(false)

  useEffect(() => {
    apiClient.get<{ grade: string; vip_due_at: string | null }>("/api/user/profile").then((res) => {
      if (res.code === 0 && res.data) {
        setIsVip(isVipActive(res.data.grade as UserGrade, res.data.vip_due_at))
      }
    })
  }, [])

  const { data: categories } = useSWR<{ id: string; name: string; depth: number; isLeaf: boolean }[]>("/api/game-categories")
  const { data: presses } = useSWR<{ id: string; name: string }[]>("/api/game-presses")
  const { data: counts } = useSWR<GameCounts>("/api/course-games/counts")
  const { games, isLoading, isValidating, hasMore, sentinelRef } =
    useInfiniteGames(activeFilter)

  const safeCounts = counts ?? { all: 0, draft: 0, published: 0, withdraw: 0 }

  return (
    <>
      {/* Hero Banner */}
      <div className="flex flex-col gap-4 rounded-2xl bg-gradient-to-tr from-teal-700 via-teal-600 to-purple-400 p-5 pb-6 lg:p-7 lg:pb-8">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-[14px] bg-white/10">
            <Puzzle className="h-7 w-7 text-white" />
          </div>
          <span className="text-xl font-extrabold tracking-tight text-white lg:text-[28px]">
            AI 随心配
          </span>
        </div>
        <p className="text-sm leading-relaxed text-white/80">
          随心定义想要的学习内容，拆解学习单元，通过趣味游戏模式巩固，记忆，理解，轻松掌握学习内容
        </p>
      </div>

      {/* Filter row */}
      <div className="flex w-full flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-wrap items-center gap-2">
          {filterKeys.map((tab) => (
            <button
              key={tab.key}
              type="button"
              onClick={() => setActiveFilter(tab.key)}
              className={`rounded-lg px-3.5 py-1.5 text-[13px] font-semibold ${
                activeFilter === tab.key
                  ? "bg-teal-600 text-white"
                  : "border border-border bg-muted text-muted-foreground"
              }`}
            >
              {tab.label}（{safeCounts[tab.countKey]}）
            </button>
          ))}
        </div>
        <button
          type="button"
          onClick={() => isVip ? setOpen(true) : setUpgradeOpen(true)}
          className="flex items-center gap-1.5 rounded-[10px] bg-gradient-to-b from-teal-600 to-teal-700 px-4 py-2 text-[13px] font-semibold text-white"
        >
          <Plus className="h-3.5 w-3.5" />
          创建我的课程游戏
        </button>
      </div>

      {/* Loading state */}
      {isLoading && <PageSpinner size="lg" />}

      {/* Game grid */}
      {!isLoading && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
          {games.map((game) => (
            <GameCardItem
              key={game.id}
              game={game}
              asDiv={!isVip}
              onClick={!isVip ? () => setUpgradeOpen(true) : undefined}
            />
          ))}
        </div>
      )}

      {/* Validating indicator */}
      {isValidating && !isLoading && <PageSpinner size="sm" />}

      {/* Empty state */}
      {!isLoading && !isValidating && games.length === 0 && (
        <div className="flex flex-col items-center gap-2 py-12 text-center">
          <Puzzle className="h-10 w-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">还没有创建任何课程游戏</p>
          <button
            type="button"
            onClick={() => isVip ? setOpen(true) : setUpgradeOpen(true)}
            className="mt-2 rounded-lg bg-teal-600 px-4 py-2 text-sm font-medium text-white"
          >
            创建第一个游戏
          </button>
        </div>
      )}

      {/* Infinite scroll sentinel */}
      {hasMore && <div ref={sentinelRef} className="h-1" />}

      {/* Create course game dialog */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent
          aria-describedby={undefined}
          showCloseButton={false}
          className="sm:max-w-[672px] overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
        >
          <VisuallyHidden>
            <DialogTitle>创建我的课程游戏</DialogTitle>
          </VisuallyHidden>
          <CreateCourseForm
            categories={categories ?? []}
            presses={presses ?? []}
            onClose={() => setOpen(false)}
          />
        </DialogContent>
      </Dialog>

      <UpgradeDialog
        open={upgradeOpen}
        onOpenChange={setUpgradeOpen}
        title="会员专属功能"
        message="升级会员即可使用 AI 随心配，创建专属学习课程"
      />
    </>
  )
}
