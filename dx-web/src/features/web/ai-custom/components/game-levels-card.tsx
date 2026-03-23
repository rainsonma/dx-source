"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Layers,
  Plus,
  ArrowRight,
  Trash2,
  MoreVertical,
  CirclePlus,
  Loader2,
} from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { AddLevelDialog } from "@/features/web/ai-custom/components/add-level-dialog";
import { useDeleteGameLevel } from "@/features/web/ai-custom/hooks/use-game-actions";

type Level = {
  id: string;
  name: string;
  description: string | null;
  order: number;
  isActive: boolean;
  _count: { items: number };
};

type GameLevelsCardProps = {
  gameId: string;
  levels: Level[];
  totalLevels: number;
  isPublished: boolean;
};

export function GameLevelsCard({
  gameId,
  levels,
  totalLevels,
  isPublished,
}: GameLevelsCardProps) {
  const [addOpen, setAddOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    name: string;
  } | null>(null);

  const deleteLevel = useDeleteGameLevel(gameId);

  function handleDeleteConfirm() {
    if (deleteTarget) {
      deleteLevel.execute(deleteTarget.id);
      setDeleteTarget(null);
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-4 overflow-y-auto rounded-[14px] border border-border bg-card p-4 lg:p-6">
        {/* Header */}
        <div className="flex flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
          <div className="flex items-center gap-2">
            <Layers className="h-[18px] w-[18px] text-teal-600" />
            <span className="text-base font-bold text-foreground">
              课程游戏关卡
            </span>
            <span className="rounded-[10px] bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
              {totalLevels} 关
            </span>
          </div>
          <button
            type="button"
            onClick={() => !isPublished && setAddOpen(true)}
            disabled={isPublished}
            title={isPublished ? "已发布的游戏不可编辑，请先撤回" : undefined}
            className="flex items-center gap-1.5 rounded-[10px] bg-gradient-to-b from-teal-500 to-teal-700 px-3.5 py-2 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <Plus className="h-3.5 w-3.5 text-white" />
            <span className="text-[13px] font-semibold text-white">
              添加课程游戏关卡
            </span>
          </button>
        </div>

        {/* Level items */}
        <div className="flex flex-col gap-4">
          {levels.map((level, index) => (
            <div
              key={level.id}
              className="flex items-center justify-between rounded-xl border border-border bg-muted px-4 py-3.5"
            >
              <div className="flex items-center gap-3.5">
                <div className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-teal-600">
                  <span className="text-sm font-bold text-white">
                    {index + 1}
                  </span>
                </div>
                <div className="flex flex-col gap-0.5">
                  <span className="text-sm font-semibold text-foreground">
                    {level.name}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {level._count.items} 个学习游戏单元
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  aria-label="更多操作"
                  className="text-muted-foreground"
                >
                  <MoreVertical className="h-4 w-4" />
                </button>
                <Link
                  href={`/hall/ai-custom/${gameId}/${level.id}`}
                  className="flex h-7 w-7 items-center justify-center rounded-md bg-teal-100"
                  aria-label="进入关卡"
                >
                  <ArrowRight className="h-3.5 w-3.5 text-teal-600" />
                </Link>
                <button
                  type="button"
                  aria-label="删除关卡"
                  onClick={() =>
                    !isPublished && setDeleteTarget({ id: level.id, name: level.name })
                  }
                  disabled={isPublished}
                  title={isPublished ? "已发布的游戏不可编辑，请先撤回" : undefined}
                  className="flex h-7 w-7 items-center justify-center rounded-md bg-red-100 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <Trash2 className="h-3.5 w-3.5 text-red-500" />
                </button>
              </div>
            </div>
          ))}

          {/* Add placeholder */}
          <button
            type="button"
            onClick={() => !isPublished && setAddOpen(true)}
            disabled={isPublished}
            title={isPublished ? "已发布的游戏不可编辑，请先撤回" : undefined}
            className="flex items-center justify-center gap-2 rounded-xl border border-border bg-muted/50 px-4 py-[18px] disabled:cursor-not-allowed disabled:opacity-50"
          >
            <CirclePlus className="h-5 w-5 text-muted-foreground" />
            <span className="text-sm font-medium text-muted-foreground">
              点击添加新关卡
            </span>
          </button>
        </div>
      </div>

      {/* Add level dialog */}
      <AddLevelDialog
        gameId={gameId}
        open={addOpen}
        onOpenChange={setAddOpen}
      />

      {/* Delete level confirmation */}
      <AlertDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确定删除关卡？</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除关卡「{deleteTarget?.name}」吗？该关卡下的所有内容源数据和学习游戏单元将一并删除，且无法恢复。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {deleteLevel.error && (
            <p className="text-sm text-red-500">{deleteLevel.error}</p>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteLevel.isPending}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              disabled={deleteLevel.isPending}
              className="bg-red-600 hover:bg-red-700"
            >
              {deleteLevel.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              确定删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
