"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Send,
  Undo2,
  Pencil,
  Trash2,
  Loader2,
  Play,
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
import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";
import { GAME_STATUSES } from "@/consts/game-status";
import { EditGameDialog } from "@/features/web/ai-custom/components/edit-game-dialog";
import {
  useDeleteGame,
  usePublishGame,
  useWithdrawGame,
} from "@/features/web/ai-custom/hooks/use-game-actions";

type CategoryOption = {
  id: string;
  name: string;
  depth: number;
  isLeaf: boolean;
};
type SelectOption = { id: string; name: string };

type GameHeroCardProps = {
  game: {
    id: string;
    name: string;
    description: string | null;
    mode: string;
    status: string;
    createdAt: Date;
    cover: { url: string } | null;
    category: { name: string } | null;
    press: { name: string } | null;
    gameCategoryId: string | null;
    gamePressId: string | null;
    isPrivate: boolean;
    _count: { levels: number; stats: number };
  };
  categories: CategoryOption[];
  presses: SelectOption[];
};

const TAG_STYLES: Record<string, string> = {
  mode: "bg-teal-600/10 text-teal-600",
  press: "bg-purple-600/10 text-purple-600",
  category: "bg-blue-500/10 text-blue-500",
};

function formatDate(date: Date) {
  return new Date(date).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}

const GRADIENT_COVERS = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

function getGradient(id: string) {
  let hash = 0;
  for (const ch of id) hash = (hash * 31 + ch.charCodeAt(0)) | 0;
  return GRADIENT_COVERS[Math.abs(hash) % GRADIENT_COVERS.length];
}

export function GameHeroCard({
  game,
  categories,
  presses,
}: GameHeroCardProps) {
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [publishOpen, setPublishOpen] = useState(false);
  const [withdrawOpen, setWithdrawOpen] = useState(false);

  const deleteAction = useDeleteGame(game.id);
  const publishAction = usePublishGame(game.id);
  const withdrawAction = useWithdrawGame(game.id);

  const modeLabel =
    GAME_MODE_LABELS[game.mode as GameMode] ?? game.mode;
  const isPublished = game.status === GAME_STATUSES.PUBLISHED;
  const canPublish = !isPublished;
  const canDelete = !isPublished;

  const tags = [
    { key: "mode", label: modeLabel },
    game.press && { key: "press", label: game.press.name },
    game.category && { key: "category", label: game.category.name },
  ].filter(Boolean) as { key: string; label: string }[];

  return (
    <>
      <div className="overflow-hidden rounded-[14px] border border-border bg-card">
        <div className="flex w-full flex-col gap-5 p-4 lg:flex-row lg:gap-7 lg:p-6">
          {/* Cover */}
          {game.cover ? (
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={game.cover.url}
              alt={game.name}
              className="h-[200px] w-full shrink-0 rounded-xl object-cover lg:h-[280px] lg:w-[280px]"
            />
          ) : (
            <div
              className={`flex h-[200px] w-full shrink-0 items-center justify-center rounded-xl ${getGradient(game.id)} lg:h-[280px] lg:w-[280px]`}
            >
              <span className="text-3xl font-bold text-white/80">
                {modeLabel}
              </span>
            </div>
          )}

          {/* Info */}
          <div className="flex flex-1 flex-col justify-between gap-4 self-stretch">
            <div className="flex flex-col gap-3">
              <h1 className="text-xl font-extrabold text-foreground lg:text-2xl">
                {game.name}
              </h1>
              {game.description && (
                <p className="text-sm leading-relaxed text-muted-foreground">
                  {game.description}
                </p>
              )}
              <div className="flex flex-wrap items-center gap-2">
                {tags.map((tag) => (
                  <span
                    key={tag.key}
                    className={`rounded-md px-2.5 py-1 text-xs font-medium ${TAG_STYLES[tag.key]}`}
                  >
                    {tag.label}
                  </span>
                ))}
              </div>
              <div className="flex items-center gap-5 text-[13px] text-muted-foreground">
                <span>共 {game._count.levels} 个关卡</span>
                <span>创建于 {formatDate(game.createdAt)}</span>
                <span>{game._count.stats} 人已参与</span>
              </div>
            </div>

            {/* Action buttons */}
            <div className="flex flex-wrap items-center gap-3">
              {canPublish && (
                <button
                  type="button"
                  onClick={() => setPublishOpen(true)}
                  className="flex items-center gap-2 rounded-xl bg-gradient-to-b from-teal-500 to-teal-700 px-6 py-2.5"
                >
                  <Send className="h-4 w-4 text-white" />
                  <span className="text-sm font-semibold text-white">
                    发布
                  </span>
                </button>
              )}
              {isPublished && (
                <Link
                  href={`/hall/games/${game.id}`}
                  className="flex items-center gap-2 rounded-xl bg-gradient-to-b from-teal-500 to-teal-700 px-6 py-2.5"
                >
                  <Play className="h-4 w-4 text-white" />
                  <span className="text-sm font-semibold text-white">
                    去玩
                  </span>
                </Link>
              )}
              {isPublished && (
                <button
                  type="button"
                  onClick={() => setWithdrawOpen(true)}
                  className="flex items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-6 py-2.5"
                >
                  <Undo2 className="h-4 w-4 text-amber-600" />
                  <span className="text-sm font-semibold text-amber-600">
                    撤回
                  </span>
                </button>
              )}
              <button
                type="button"
                onClick={() => !isPublished && setEditOpen(true)}
                disabled={isPublished}
                title={isPublished ? "已发布的游戏不可编辑，请先撤回" : undefined}
                className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-2.5 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <Pencil className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-semibold text-muted-foreground">
                  编辑
                </span>
              </button>
              {canDelete && (
                <button
                  type="button"
                  onClick={() => setDeleteOpen(true)}
                  className="flex items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-5 py-2.5"
                >
                  <Trash2 className="h-4 w-4 text-red-500" />
                  <span className="text-sm font-semibold text-red-500">
                    删除
                  </span>
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Edit dialog */}
      <EditGameDialog
        gameId={game.id}
        categories={categories}
        presses={presses}
        defaultValues={{
          name: game.name,
          description: game.description,
          mode: game.mode,
          gameCategoryId: game.gameCategoryId,
          gamePressId: game.gamePressId,
          coverUrl: game.cover?.url ?? null,
          isPrivate: game.isPrivate,
        }}
        open={editOpen}
        onOpenChange={setEditOpen}
      />

      {/* Delete game confirmation */}
      <AlertDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (!open) deleteAction.clearError()
          setDeleteOpen(open)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确定删除这个游戏？</AlertDialogTitle>
            <AlertDialogDescription>
              删除后将无法恢复，包括所有关卡数据都会被删除。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {deleteAction.error && (
            <p className="text-sm text-red-500">{deleteAction.error}</p>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteAction.isPending}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault()
                deleteAction.execute(() => setDeleteOpen(false))
              }}
              disabled={deleteAction.isPending}
              className="bg-red-600 hover:bg-red-700"
            >
              {deleteAction.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              确定删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Publish confirmation */}
      <AlertDialog open={publishOpen} onOpenChange={setPublishOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确定发布这个游戏？</AlertDialogTitle>
            <AlertDialogDescription>
              发布后其他用户将可以看到并参与这个游戏。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={publishAction.isPending}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={publishAction.execute}
              disabled={publishAction.isPending}
              className="!bg-teal-600 hover:!bg-teal-700"
            >
              {publishAction.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              确定发布
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Withdraw confirmation */}
      <AlertDialog
        open={withdrawOpen}
        onOpenChange={(open) => {
          if (!open) withdrawAction.clearError()
          setWithdrawOpen(open)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确定撤回这个游戏？</AlertDialogTitle>
            <AlertDialogDescription>
              撤回后其他用户将无法看到这个游戏，你可以随时重新发布。
            </AlertDialogDescription>
          </AlertDialogHeader>
          {withdrawAction.error && (
            <p className="text-sm text-red-500">{withdrawAction.error}</p>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={withdrawAction.isPending}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault()
                withdrawAction.execute(() => setWithdrawOpen(false))
              }}
              disabled={withdrawAction.isPending}
              className="!bg-amber-600 hover:!bg-amber-700"
            >
              {withdrawAction.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              确定撤回
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
