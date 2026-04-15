"use client";

import { useState, useCallback, useEffect, useRef, useId } from "react";
import { swrMutate } from "@/lib/swr";
import { useRateLimit } from "@/hooks/use-rate-limit";
import {
  DndContext,
  DragOverlay,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  sortableKeyboardCoordinates,
  arrayMove,
} from "@dnd-kit/sortable";
import {
  Database,
  Layers,
  Loader2,
  MessageSquareText,
  Plus,
  SpellCheck,
  Sparkles,
  Trash2,
  Puzzle,
  Scissors,
  CircleCheck,
  CircleDashed,
  ShieldAlert,
} from "lucide-react";
import { toast } from "sonner";
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
import type { GameMode } from "@/consts/game-mode";
import { GAME_MODES } from "@/consts/game-mode";
import { AddMetadataDialog } from "@/features/web/ai-custom/components/add-metadata-dialog";
import { AddVocabDialog } from "@/features/web/ai-custom/components/add-vocab-dialog";
import { SortableMetaItem } from "@/features/web/ai-custom/components/sortable-meta-item";
import { SortableContentItem } from "@/features/web/ai-custom/components/sortable-content-item";
import {
  reorderMetaAction,
  deleteMetaAction,
  fetchContentItemsAction,
  reorderItemAction,
  updateContentItemTextAction,
  insertContentItemAction,
  deleteContentItemAction,
  deleteAllLevelContentAction,
} from "@/features/web/ai-custom/actions/course-game.action";
import {
  breakMetadata,
  generateContentItems,
} from "@/features/web/ai-custom/helpers/generate-items-api";
import {
  breakVocabMetadata,
  generateVocabContentItems,
} from "@/features/web/ai-custom/helpers/vocab-generate-items-api";
import type { LevelMeta, LevelContentItem } from "@/features/web/ai-custom/actions/course-game.action";

type ContentGroup = {
  meta: { id: string } | null;
  items: LevelContentItem[];
};
import { MAX_SENTENCES, MAX_VOCAB } from "@/features/web/ai-custom/helpers/format-metadata";
import { MAX_METAS_PER_LEVEL } from "@/features/web/ai-custom/helpers/vocab-format-metadata";
import { SOURCE_TYPES } from "@/consts/source-type";
import { ProcessingOverlay } from "@/features/web/ai-custom/components/processing-overlay";
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";

type LevelUnitsPanelProps = {
  gameId: string;
  levelId: string;
  gameMode: GameMode;
  initialMetas: LevelMeta[];
  readOnly?: boolean;
  sentenceItemCount: number;
  vocabItemCount: number;
};

function calculateNewOrder(
  items: { order: number }[],
  oldIndex: number,
  newIndex: number
): number {
  const reordered = arrayMove(items, oldIndex, newIndex);
  const prev = newIndex > 0 ? reordered[newIndex - 1].order : 0;
  const next =
    newIndex < reordered.length - 1
      ? reordered[newIndex + 1].order
      : prev + 2000;

  return (prev + next) / 2;
}

export function LevelUnitsPanel({
  gameId,
  levelId,
  gameMode,
  initialMetas,
  readOnly,
  sentenceItemCount,
  vocabItemCount,
}: LevelUnitsPanelProps) {
  const isVocabMode = gameMode !== GAME_MODES.WORD_SENTENCE;
  const metaDndId = useId();
  const itemDndId = useId();
  const [metadataDialogOpen, setMetadataDialogOpen] = useState(false);
  const [metas, setMetas] = useState<LevelMeta[]>(initialMetas);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [isReordering, setIsReordering] = useState(false);

  const [isBreaking, setIsBreaking] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [contentItems, setContentItems] = useState<LevelContentItem[]>([]);
  const [isLoadingItems, setIsLoadingItems] = useState(false);
  const [activeItemId, setActiveItemId] = useState<string | null>(null);
  const [isReorderingItem, setIsReorderingItem] = useState(false);
  const [progress, setProgress] = useState<{ done: number; total: number } | null>(null);

  const [beanDialogOpen, setBeanDialogOpen] = useState(false);
  const [beanRequired, setBeanRequired] = useState(0);
  const [beanAvailable, setBeanAvailable] = useState(0);

  const [breakConfirmOpen, setBreakConfirmOpen] = useState(false);
  const [genConfirmOpen, setGenConfirmOpen] = useState(false);
  const [deleteMetaConfirmOpen, setDeleteMetaConfirmOpen] = useState(false);
  const [pendingDeleteMetaId, setPendingDeleteMetaId] = useState<string | null>(null);
  const [deleteItemConfirmOpen, setDeleteItemConfirmOpen] = useState(false);
  const [pendingDeleteItemId, setPendingDeleteItemId] = useState<string | null>(null);
  const [deleteAllConfirmOpen, setDeleteAllConfirmOpen] = useState(false);
  const [isDeletingAll, setIsDeletingAll] = useState(false);
  const [isInserting, setIsInserting] = useState(false);
  const insertingRef = useRef(false);
  const checkRateLimit = useRateLimit();

  const breakAbortRef = useRef<AbortController | null>(null);
  const genAbortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    setMetas(initialMetas);
  }, [initialMetas]);

  useEffect(() => {
    if (selectedId) {
      fetchContentItemsAction(gameId, levelId).then((result) => {
        // Filter items by the selected meta from the grouped response
        const group = (result.items as unknown as ContentGroup[]).find(
          (g) => g.meta?.id === selectedId
        );
        setContentItems(group?.items ?? []);
      });
    }
  }, [initialMetas, selectedId, gameId, levelId]);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  function handleDragStart(event: DragStartEvent) {
    setActiveId(event.active.id as string);
  }

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      setActiveId(null);
      const { active, over } = event;
      if (!over || active.id === over.id) return;
      if (isReordering) return;

      const oldIndex = metas.findIndex((m) => m.id === active.id);
      const newIndex = metas.findIndex((m) => m.id === over.id);
      if (oldIndex === -1 || newIndex === -1) return;

      const newOrder = calculateNewOrder(metas, oldIndex, newIndex);
      const reordered = arrayMove(metas, oldIndex, newIndex).map((m) =>
        m.id === active.id ? { ...m, order: newOrder } : m
      );

      const prevMetas = metas;
      setMetas(reordered);
      setIsReordering(true);

      try {
        const result = await reorderMetaAction(
          gameId,
          levelId,
          active.id as string,
          newOrder
        );

        if (result.error) {
          setMetas(prevMetas);
          toast.error(result.error);
        }
      } finally {
        setIsReordering(false);
      }
    },
    [metas, gameId, levelId, isReordering]
  );

  const handleSelectMeta = useCallback(async (metaId: string) => {
    setSelectedId(metaId);
    setIsLoadingItems(true);
    const result = await fetchContentItemsAction(gameId, levelId);
    // Filter items by the selected meta from the grouped response
    const group = (result.items as unknown as ContentGroup[]).find(
      (g) => g.meta?.id === metaId
    );
    setContentItems(group?.items ?? []);
    setIsLoadingItems(false);
  }, [gameId, levelId]);

  const handleBreak = useCallback(async () => {
    const pending = metas.filter((m) => !m.isBreakDone);
    if (pending.length === 0) return;

    const controller = new AbortController();
    breakAbortRef.current = controller;
    setIsBreaking(true);
    setProgress({ done: 0, total: pending.length });

    const result = isVocabMode
      ? await breakVocabMetadata(levelId, controller.signal, (event) => { setProgress({ done: event.done, total: event.total }); })
      : await breakMetadata(levelId, controller.signal, (event) => { setProgress({ done: event.done, total: event.total }); });

    breakAbortRef.current = null;
    setIsBreaking(false);
    setProgress(null);

    if (!result.ok) {
      if (result.message !== "已取消") {
        if (result.code === "INSUFFICIENT_BEANS") {
          setBeanRequired(result.required ?? 0);
          setBeanAvailable(result.available ?? 0);
          setBeanDialogOpen(true);
        } else {
          toast.error(result.message);
        }
      }
      return;
    }

    if (result.failed === 0) {
      toast.success(`已分解 ${result.processed} 条元数据`);
    } else if (result.processed > 0) {
      toast.warning(
        `部分分解失败 (成功: ${result.processed}, 失败: ${result.failed})`
      );
    } else {
      toast.error("分解失败，请稍后重试");
    }

    swrMutate("/api/course-games");
  }, [metas, levelId, isVocabMode]);

  const handleGenerate = useCallback(async () => {
    const pending = metas.filter((m) => m.isBreakDone && !m.isItemDone);
    if (pending.length === 0) return;

    const controller = new AbortController();
    genAbortRef.current = controller;
    setIsGenerating(true);
    setProgress({ done: 0, total: pending.length });

    const result = isVocabMode
      ? await generateVocabContentItems(levelId, controller.signal, (event) => { setProgress({ done: event.done, total: event.total }); })
      : await generateContentItems(levelId, controller.signal, (event) => { setProgress({ done: event.done, total: event.total }); });

    genAbortRef.current = null;
    setIsGenerating(false);
    setProgress(null);

    if (!result.ok) {
      if (result.message !== "已取消") {
        if (result.code === "INSUFFICIENT_BEANS") {
          setBeanRequired(result.required ?? 0);
          setBeanAvailable(result.available ?? 0);
          setBeanDialogOpen(true);
        } else {
          toast.error(result.message);
        }
      }
      return;
    }

    if (result.failed === 0) {
      toast.success(`已生成 ${result.processed} 条练习单元`);
    } else if (result.processed > 0) {
      toast.warning(
        `部分生成失败 (成功: ${result.processed}, 失败: ${result.failed})`
      );
    } else {
      toast.error("生成失败，请稍后重试");
    }

    swrMutate("/api/course-games");
  }, [metas, levelId, isVocabMode]);

  function handleItemDragStart(event: DragStartEvent) {
    setActiveItemId(event.active.id as string);
  }

  const handleItemDragEnd = useCallback(
    async (event: DragEndEvent) => {
      setActiveItemId(null);
      const { active, over } = event;
      if (!over || active.id === over.id) return;
      if (isReorderingItem) return;

      const oldIndex = contentItems.findIndex((i) => i.id === active.id);
      const newIndex = contentItems.findIndex((i) => i.id === over.id);
      if (oldIndex === -1 || newIndex === -1) return;

      const newOrder = calculateNewOrder(contentItems, oldIndex, newIndex);
      const reordered = arrayMove(contentItems, oldIndex, newIndex).map((i) =>
        i.id === active.id ? { ...i, order: newOrder } : i
      );

      const prevItems = contentItems;
      setContentItems(reordered);
      setIsReorderingItem(true);

      try {
        const result = await reorderItemAction(
          gameId,
          levelId,
          active.id as string,
          newOrder
        );
        if (result.error) {
          setContentItems(prevItems);
          toast.error(result.error);
        }
      } finally {
        setIsReorderingItem(false);
      }
    },
    [contentItems, gameId, isReorderingItem, levelId]
  );

  const handleUpdateItemText = useCallback(
    async (itemId: string, content: string, translation: string | null): Promise<boolean> => {
      const prevItems = contentItems;
      setContentItems((items) =>
        items.map((i) => (i.id === itemId ? { ...i, content, translation } : i))
      );

      const result = await updateContentItemTextAction(gameId, itemId, content, translation);
      if (result.error) {
        setContentItems(prevItems);
        toast.error(result.error);
        return false;
      }
      return true;
    },
    [contentItems, gameId]
  );

  const handleInsertItem = useCallback(
    async (itemId: string, direction: "above" | "below") => {
      if (!selectedId || insertingRef.current) return;
      if (!checkRateLimit()) return;
      insertingRef.current = true;
      setIsInserting(true);

      try {
        const result = await insertContentItemAction(gameId, {
          gameLevelId: levelId,
          contentMetaId: selectedId,
          content: "",
          contentType: "word",
          translation: null,
          referenceItemId: itemId,
          direction,
        });

        if (result.error) {
          toast.error(result.error);
          return;
        }

        if (result.item) {
          setContentItems((prev) => {
            const idx = prev.findIndex((i) => i.id === itemId);
            const insertIdx = direction === "above" ? idx : idx + 1;
            const next = [...prev];
            next.splice(insertIdx, 0, result.item!);
            return next;
          });
          setMetas((prev) =>
            prev.map((m) =>
              m.id === selectedId
                ? { ...m, isItemDone: false, itemCount: m.itemCount + 1 }
                : m
            )
          );
        }
      } finally {
        insertingRef.current = false;
        setIsInserting(false);
      }
    },
    [selectedId, gameId, levelId, checkRateLimit]
  );

  const handleCopyItem = useCallback(
    async (itemId: string, direction: "above" | "below") => {
      if (!selectedId || insertingRef.current) return;
      if (!checkRateLimit()) return;
      insertingRef.current = true;
      setIsInserting(true);
      const source = contentItems.find((i) => i.id === itemId);
      if (!source) {
        insertingRef.current = false;
        setIsInserting(false);
        return;
      }

      try {
        const result = await insertContentItemAction(gameId, {
          gameLevelId: levelId,
          contentMetaId: selectedId,
          content: source.content,
          contentType: source.contentType,
          translation: source.translation,
          referenceItemId: itemId,
          direction,
        });

        if (result.error) {
          toast.error(result.error);
          return;
        }

        if (result.item) {
          setContentItems((prev) => {
            const idx = prev.findIndex((i) => i.id === itemId);
            const insertIdx = direction === "above" ? idx : idx + 1;
            const next = [...prev];
            next.splice(insertIdx, 0, result.item!);
            return next;
          });
          setMetas((prev) =>
            prev.map((m) =>
              m.id === selectedId
                ? { ...m, isItemDone: false, itemCount: m.itemCount + 1 }
                : m
            )
          );
        }
      } finally {
        insertingRef.current = false;
        setIsInserting(false);
      }
    },
    [contentItems, selectedId, gameId, levelId, checkRateLimit]
  );

  const handleRequestDeleteMeta = useCallback((metaId: string) => {
    setPendingDeleteMetaId(metaId);
    setDeleteMetaConfirmOpen(true);
  }, []);

  const handleConfirmDeleteMeta = useCallback(async () => {
    if (!pendingDeleteMetaId) return;
    setDeleteMetaConfirmOpen(false);

    const prevMetas = metas;
    setMetas((prev) => prev.filter((m) => m.id !== pendingDeleteMetaId));
    if (selectedId === pendingDeleteMetaId) {
      setSelectedId(null);
      setContentItems([]);
    }

    const result = await deleteMetaAction(gameId, pendingDeleteMetaId);
    if (result.error) {
      setMetas(prevMetas);
      toast.error(result.error);
    }

    setPendingDeleteMetaId(null);
  }, [pendingDeleteMetaId, metas, selectedId, gameId]);

  const handleRequestDeleteItem = useCallback((itemId: string) => {
    setPendingDeleteItemId(itemId);
    setDeleteItemConfirmOpen(true);
  }, []);

  const handleConfirmDeleteItem = useCallback(async () => {
    if (!pendingDeleteItemId) return;
    setDeleteItemConfirmOpen(false);

    const prevItems = contentItems;
    const prevMetas = metas;
    const remaining = contentItems.filter((i) => i.id !== pendingDeleteItemId);

    setContentItems(remaining);
    setMetas((prev) =>
      prev.map((m) => {
        if (m.id !== selectedId) return m;
        const newCount = m.itemCount - 1;
        const allGenerated = remaining.length > 0 && remaining.every((i) => i.items !== null);
        const breakDone = newCount > 0 ? m.isBreakDone : false;
        return {
          ...m,
          itemCount: newCount,
          isBreakDone: breakDone,
          isItemDone: breakDone && newCount > 0 && allGenerated,
        };
      })
    );

    const result = await deleteContentItemAction(gameId, pendingDeleteItemId);
    if (result.error) {
      setContentItems(prevItems);
      setMetas(prevMetas);
      toast.error(result.error);
    }

    setPendingDeleteItemId(null);
  }, [pendingDeleteItemId, contentItems, metas, selectedId, gameId]);

  const handleConfirmDeleteAll = useCallback(async () => {
    setDeleteAllConfirmOpen(false);
    setIsDeletingAll(true);

    const result = await deleteAllLevelContentAction(gameId, levelId);

    setIsDeletingAll(false);

    if (result.error) {
      toast.error(result.error);
      return;
    }

    setMetas([]);
    setContentItems([]);
    setSelectedId(null);
    toast.success("已删除当前关卡全部元数据和练习单元");
  }, [gameId, levelId]);

  const activeMeta = activeId ? metas.find((m) => m.id === activeId) : null;
  const selectedMeta = selectedId ? metas.find((m) => m.id === selectedId) : null;
  const breakPendingCount = metas.filter((m) => !m.isBreakDone).length;
  const genPendingCount = metas.filter(
    (m) => m.isBreakDone && !m.isItemDone
  ).length;
  const isLevelComplete =
    metas.length > 0 && metas.every((m) => m.isBreakDone && m.isItemDone);

  const metaSentenceCount = metas.filter((m) => m.sourceType === SOURCE_TYPES.SENTENCE).length;
  const metaVocabCount = metas.filter((m) => m.sourceType === SOURCE_TYPES.VOCAB).length;
  const isAtCapacity = isVocabMode
    ? metas.length >= MAX_METAS_PER_LEVEL
    : metaSentenceCount / MAX_SENTENCES + metaVocabCount / MAX_VOCAB >= 1;
  const totalItemCount = metas.reduce((sum, m) => sum + m.itemCount, 0);

  return (
    <>
      <div className="flex flex-1 flex-col gap-3 overflow-hidden">

      {readOnly && (
        <div className="flex shrink-0 items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm font-medium text-amber-700">
          <ShieldAlert className="h-4 w-4 shrink-0" />
          已发布的游戏内容不可编辑，如需修改请先撤回游戏
        </div>
      )}

      {/* Two columns */}
      <div className="flex flex-1 flex-col gap-5 overflow-hidden lg:flex-row">
        {/* Left: Metadata */}
        <div className="relative flex flex-1 flex-col overflow-hidden pt-3">
          {/* Floating level status badge */}
          <div className={`absolute top-0 left-4 z-10 flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold shadow-sm ${isLevelComplete ? "bg-teal-600 text-white" : "bg-amber-500 text-white"}`}>
            <span className={isLevelComplete ? "text-teal-100" : "text-amber-100"}>关卡状态</span>
            {isLevelComplete ? (
              <>
                <CircleCheck className="h-3.5 w-3.5 text-emerald-300" />
                <span>已完成</span>
              </>
            ) : (
              <>
                <CircleDashed className="h-3.5 w-3.5 text-amber-200" />
                <span>未完成</span>
              </>
            )}
          </div>

          <div className="flex flex-1 flex-col overflow-hidden rounded-[14px] border border-border bg-card">
          <div className="shrink-0 p-4 pb-0 lg:p-5 lg:pb-0">
            <div className="flex flex-col items-start justify-between gap-3 pb-2 pt-2 sm:flex-row sm:items-center lg:flex-col lg:items-start xl:flex-row xl:items-center">
              <div className="flex items-center gap-2">
                <Database className="h-4 w-4 text-teal-600" />
                <span className="text-[15px] font-bold text-foreground">元数据</span>
              </div>
              <div className="flex items-center gap-2 lg:flex-col lg:items-start xl:flex-row xl:items-center">
                <div className="flex items-center overflow-hidden rounded-lg bg-gradient-to-r from-teal-500 via-teal-600 to-teal-700">
                  <button
                    type="button"
                    onClick={() => setMetadataDialogOpen(true)}
                    disabled={isAtCapacity || readOnly}
                    title={readOnly ? "已发布的游戏不可编辑，请先撤回" : (isAtCapacity ? (isVocabMode ? `已达上限（${metas.length}/${MAX_METAS_PER_LEVEL}）` : `已达容量上限（语句 ${metaSentenceCount}/${MAX_SENTENCES}，词汇 ${metaVocabCount}/${MAX_VOCAB}）`) : undefined)}
                    className="flex items-center gap-1 px-3 py-1.5 disabled:opacity-50"
                  >
                    <Plus className="h-3.5 w-3.5 text-white" />
                    <span className="text-xs font-semibold text-white">1: 添加</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setBreakConfirmOpen(true)}
                    disabled={isBreaking || isGenerating || breakPendingCount === 0 || readOnly}
                    title={readOnly ? "已发布的游戏不可编辑，请先撤回" : undefined}
                    className="flex items-center gap-1 border-x border-white/20 px-3 py-1.5 disabled:opacity-50"
                  >
                    <Scissors className="h-3.5 w-3.5 text-white" />
                    <span className="text-xs font-semibold text-white">2: 分解</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setGenConfirmOpen(true)}
                    disabled={isGenerating || isBreaking || genPendingCount === 0 || readOnly}
                    title={readOnly ? "已发布的游戏不可编辑，请先撤回" : undefined}
                    className="flex items-center gap-1 px-3 py-1.5 disabled:opacity-50"
                  >
                    <Sparkles className="h-3.5 w-3.5 text-white" />
                    <span className="text-xs font-semibold text-white">3: 生成</span>
                  </button>
                </div>
                <button
                  type="button"
                  onClick={() => setDeleteAllConfirmOpen(true)}
                  disabled={metas.length === 0 || isBreaking || isGenerating || isDeletingAll || readOnly}
                  title={readOnly ? "已发布的游戏不可编辑，请先撤回" : undefined}
                  className="flex items-center gap-1 rounded-lg bg-red-100 px-3 py-1.5 disabled:opacity-50"
                >
                  <Trash2 className="h-3.5 w-3.5 text-red-500" />
                  <span className="text-xs font-semibold text-red-500">删除</span>
                </button>
              </div>
            </div>
            {/* Stats bar */}
            <div className="mb-3 flex items-center gap-4 rounded-lg bg-teal-100 px-3 py-2 text-xs text-teal-600 lg:flex-col lg:items-start lg:gap-2 xl:flex-row xl:items-center xl:gap-4">
              <span className="flex items-center gap-1"><Layers className="h-3 w-3" />共计：<span className="font-semibold text-teal-800">{metas.length}</span></span>
              {!isVocabMode && (
                <>
                  <span className="flex items-center gap-1"><MessageSquareText className="h-3 w-3" />语句：<span className="font-semibold text-teal-800">{sentenceItemCount}</span></span>
                  <span className="flex items-center gap-1"><SpellCheck className="h-3 w-3" />词汇：<span className="font-semibold text-teal-800">{vocabItemCount}</span></span>
                </>
              )}
              <span className="ml-auto flex items-center gap-1 lg:ml-0 xl:ml-auto"><Puzzle className="h-3 w-3" />练习单元总数：<span className="font-semibold text-teal-800">{totalItemCount}</span></span>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto px-4 pb-4 lg:px-5 lg:pb-5">
          {metas.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-12 text-muted-foreground">
              <Database className="h-8 w-8" />
              <span className="text-sm">暂无元数据，点击上方按钮添加</span>
            </div>
          ) : (
            <DndContext
              id={metaDndId}
              sensors={readOnly ? [] : sensors}
              collisionDetection={closestCenter}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
            >
              <SortableContext
                items={metas.map((m) => m.id)}
                strategy={verticalListSortingStrategy}
              >
                <div className="flex flex-col gap-2">
                  {metas.map((meta) => (
                    <SortableMetaItem
                      key={meta.id}
                      meta={meta}
                      isSelected={selectedId === meta.id}
                      onClick={() => handleSelectMeta(meta.id)}
                      onDelete={readOnly ? undefined : handleRequestDeleteMeta}
                    />
                  ))}
                </div>
              </SortableContext>
              <DragOverlay>
                {activeMeta && (
                  <SortableMetaItem meta={activeMeta} />
                )}
              </DragOverlay>
            </DndContext>
          )}
          </div>
          </div>
        </div>

        {/* Right: Practice units */}
        <div className="flex flex-1 flex-col overflow-hidden rounded-[14px] border-3 border-teal-600 bg-teal-50 p-4 lg:p-5">
          <div className="flex shrink-0 items-center justify-between pb-4">
            <div className="flex items-center gap-2">
              <Puzzle className="h-4 w-4 text-teal-600" />
              <span className="text-[15px] font-bold text-foreground">练习单元</span>
              {selectedId && (
                <span className="rounded-[10px] bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
                  {contentItems.length} 个
                </span>
              )}
            </div>
          </div>

          {selectedMeta && (
            <div className="mb-3 shrink-0 rounded-lg bg-teal-100 px-3 py-5 ring-2 ring-teal-300">
              <p className="text-sm font-bold text-foreground">{selectedMeta.sourceData}</p>
              {selectedMeta.translation && (
                <p className="mt-0.5 text-xs text-muted-foreground">{selectedMeta.translation}</p>
              )}
            </div>
          )}

          <div className="flex-1 overflow-y-auto">
            {!selectedId ? (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-muted-foreground">
                <Puzzle className="h-8 w-8" />
                <span className="text-sm">选择左侧元数据查看练习单元</span>
              </div>
            ) : isLoadingItems ? (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-muted-foreground">
                <Loader2 className="h-8 w-8 animate-spin" />
                <span className="text-sm">加载中...</span>
              </div>
            ) : contentItems.length === 0 ? (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-muted-foreground">
                <Puzzle className="h-8 w-8" />
                <span className="text-sm">暂无练习单元</span>
              </div>
            ) : (
              <DndContext
                id={itemDndId}
                sensors={readOnly ? [] : sensors}
                collisionDetection={closestCenter}
                onDragStart={handleItemDragStart}
                onDragEnd={handleItemDragEnd}
              >
                <SortableContext
                  items={contentItems.map((i) => i.id)}
                  strategy={verticalListSortingStrategy}
                >
                  <div className="flex flex-col gap-2.5">
                    {contentItems.map((item, index) => (
                      <SortableContentItem
                        key={item.id}
                        item={item}
                        index={index + 1}
                        actionsDisabled={isInserting || readOnly}
                        onUpdateText={readOnly ? undefined : handleUpdateItemText}
                        onInsert={readOnly ? undefined : handleInsertItem}
                        onCopy={readOnly ? undefined : handleCopyItem}
                        onDelete={readOnly ? undefined : handleRequestDeleteItem}
                      />
                    ))}
                  </div>
                </SortableContext>
                <DragOverlay>
                  {activeItemId && contentItems.find((i) => i.id === activeItemId) && (
                    <SortableContentItem
                      item={contentItems.find((i) => i.id === activeItemId)!}
                      index={contentItems.findIndex((i) => i.id === activeItemId) + 1}
                      onUpdateText={handleUpdateItemText}
                    />
                  )}
                </DragOverlay>
              </DndContext>
            )}
          </div>
        </div>
      </div>
      </div>

      {isVocabMode ? (
        <AddVocabDialog
          gameId={gameId}
          levelId={levelId}
          gameMode={gameMode}
          open={metadataDialogOpen}
          onOpenChange={setMetadataDialogOpen}
          existingMetaCount={metas.length}
        />
      ) : (
        <AddMetadataDialog
          gameId={gameId}
          levelId={levelId}
          open={metadataDialogOpen}
          onOpenChange={setMetadataDialogOpen}
          existingSentenceCount={metaSentenceCount}
          existingVocabCount={metaVocabCount}
        />
      )}

      {progress && (
        <ProcessingOverlay
          done={progress.done}
          total={progress.total}
        />
      )}

      <AlertDialog open={breakConfirmOpen} onOpenChange={setBreakConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认分解</AlertDialogTitle>
            <AlertDialogDescription>
              将对当前关卡全部 {breakPendingCount} 条未分解的元数据进行分解，此操作由 AI 处理，可能需要一些时间。确定要继续吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-teal-600 hover:bg-teal-700" onClick={() => { setBreakConfirmOpen(false); handleBreak(); }}>
              确认分解
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={genConfirmOpen} onOpenChange={setGenConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认生成</AlertDialogTitle>
            <AlertDialogDescription>
              将对当前关卡全部 {genPendingCount} 条待生成的元数据生成练习单元，此操作由 AI 处理，可能需要一些时间。确定要继续吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-teal-600 hover:bg-teal-700" onClick={() => { setGenConfirmOpen(false); handleGenerate(); }}>
              确认生成
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={deleteMetaConfirmOpen} onOpenChange={setDeleteMetaConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除元数据</AlertDialogTitle>
            <AlertDialogDescription>
              将删除该元数据及其关联的练习单元。此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setPendingDeleteMetaId(null)}>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-red-600 hover:bg-red-700" onClick={handleConfirmDeleteMeta}>
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={deleteItemConfirmOpen} onOpenChange={setDeleteItemConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除该练习单元吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setPendingDeleteItemId(null)}>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-red-600 hover:bg-red-700" onClick={handleConfirmDeleteItem}>
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={deleteAllConfirmOpen} onOpenChange={setDeleteAllConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认清空关卡</AlertDialogTitle>
            <AlertDialogDescription>
              将删除当前关卡全部 {metas.length} 条元数据和 {totalItemCount} 个练习单元。此操作不可撤销，确定要继续吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-red-600 hover:bg-red-700" onClick={handleConfirmDeleteAll}>
              确认清空
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <InsufficientBeansDialog
        open={beanDialogOpen}
        onOpenChange={setBeanDialogOpen}
        required={beanRequired}
        available={beanAvailable}
      />
    </>
  );
}
