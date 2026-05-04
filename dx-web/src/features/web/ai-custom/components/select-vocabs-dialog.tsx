"use client";

import { useState, useEffect, useCallback } from "react";
import { Library, X, Search, Loader2, Check } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { toast } from "sonner";
import type { ContentVocabData, AddedGameVocab, PosKey } from "@/lib/api-client";
import type { GameMode } from "@/consts/game-mode";
import { listMyVocabsAction } from "@/features/web/ai-custom/actions/content-vocab.action";
import { addGameVocabsAction } from "@/features/web/ai-custom/actions/game-vocab.action";
import { vocabBatchSize } from "@/features/web/ai-custom/helpers/vocab-format-metadata";

const POS_LABELS: Partial<Record<PosKey, string>> = {
  n: "n", v: "v", adj: "adj", adv: "adv", prep: "prep",
  conj: "conj", pron: "pron", art: "art", num: "num", int: "int", aux: "aux", det: "det",
};

type DefinitionMap = Partial<Record<PosKey, string>>;

function parseDefinition(raw: string | null | undefined): DefinitionMap[] {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? (parsed as DefinitionMap[]) : [];
  } catch {
    return [];
  }
}

type SelectVocabsDialogProps = {
  gameId: string;
  levelId: string;
  gameMode: GameMode;
  alreadyPlacedVocabIds: string[];
  onClose: () => void;
  onAdded: (added: AddedGameVocab[]) => void;
};

const PAGE_SIZE = 50;

export function SelectVocabsDialog({
  gameId,
  levelId,
  gameMode,
  alreadyPlacedVocabIds,
  onClose,
  onAdded,
}: SelectVocabsDialogProps) {
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [items, setItems] = useState<ContentVocabData[]>([]);
  const [cursor, setCursor] = useState<string>("");
  const [hasMore, setHasMore] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [isSaving, setIsSaving] = useState(false);

  const batchSize = vocabBatchSize(gameMode);
  const placedSet = new Set(alreadyPlacedVocabIds);

  async function loadPage(nextCursor: string, reset: boolean, searchVal: string) {
    if (reset) setIsLoading(true); else setIsLoadingMore(true);
    try {
      const res = await listMyVocabsAction({
        cursor: nextCursor || undefined,
        search: searchVal || undefined,
        limit: PAGE_SIZE,
      });
      if (res.code !== 0) { toast.error(res.message); return; }
      const data = res.data;
      setItems((prev) => reset ? data.items : [...prev, ...data.items]);
      setCursor(data.nextCursor);
      setHasMore(data.hasMore);
    } finally {
      if (reset) setIsLoading(false); else setIsLoadingMore(false);
    }
  }

  useEffect(() => {
    loadPage("", true, debouncedSearch);
  }, [debouncedSearch]);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  function toggleSelect(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }

  const handleLoadMore = useCallback(async () => {
    if (isLoadingMore || !hasMore) return;
    await loadPage(cursor, false, debouncedSearch);
  }, [isLoadingMore, hasMore, cursor, debouncedSearch]);

  async function handleConfirm() {
    if (selected.size === 0) return;
    setIsSaving(true);
    try {
      const res = await addGameVocabsAction(gameId, levelId, Array.from(selected));
      if (res.code !== 0) { toast.error(res.message); return; }
      onAdded(res.data ?? []);
      onClose();
    } finally {
      setIsSaving(false);
    }
  }

  const isConfirmDisabled = isSaving || selected.size === 0 ||
    (batchSize > 0 && selected.size % batchSize !== 0);

  const batchHint = batchSize > 0
    ? (selected.size % batchSize !== 0
      ? `需要 ${batchSize} 的倍数（当前 ${selected.size} 个）`
      : `已选 ${selected.size} 个`)
    : `已选 ${selected.size} 个`;

  return (
    <Dialog open onOpenChange={(open) => { if (!open) onClose(); }}>
      <DialogContent
        aria-describedby={undefined}
        showCloseButton={false}
        className="sm:max-w-lg overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
      >
        <VisuallyHidden><DialogTitle>从词汇库选择词汇</DialogTitle></VisuallyHidden>

        <div className="flex flex-col max-h-[90vh]">
          {/* Header */}
          <div className="flex shrink-0 items-center justify-between px-6 py-4">
            <div className="flex items-center gap-2.5">
              <Library className="h-5 w-5 text-teal-600" />
              <h2 className="text-lg font-bold text-foreground">选择词汇</h2>
            </div>
            <button type="button" onClick={onClose} aria-label="关闭"
              className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted">
              <X className="h-3.5 w-3.5 text-muted-foreground" />
            </button>
          </div>

          {/* Search */}
          <div className="shrink-0 px-6 pb-3">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input type="text" value={search} onChange={(e) => setSearch(e.target.value)}
                placeholder="搜索词汇..."
                className="h-9 w-full rounded-lg border border-border bg-muted/50 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500" />
            </div>
          </div>

          {/* List */}
          <div className="flex-1 overflow-y-auto px-6">
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : items.length === 0 ? (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-muted-foreground">
                <Library className="h-8 w-8" />
                <span className="text-sm">词汇库为空，请先添加词汇</span>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5 pb-3">
                {items.map((vocab) => {
                  const defs = parseDefinition(vocab.definition);
                  const isPlaced = placedSet.has(vocab.id);
                  const isSelected = selected.has(vocab.id);
                  return (
                    <label key={vocab.id}
                      className={`flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-2.5 transition-colors ${
                        isSelected
                          ? "border-teal-400 bg-teal-50"
                          : "border-border bg-background hover:bg-muted/50"
                      }`}>
                      <input type="checkbox" checked={isSelected} onChange={() => toggleSelect(vocab.id)}
                        className="mt-0.5 h-4 w-4 accent-teal-600" />
                      <div className="flex flex-1 flex-col gap-1 min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="text-sm font-medium text-foreground">{vocab.content}</span>
                          {isPlaced && (
                            <span className="flex items-center gap-0.5 rounded-full bg-blue-100 px-2 py-0.5 text-[10px] font-semibold text-blue-700">
                              <Check className="h-2.5 w-2.5" />已使用
                            </span>
                          )}
                        </div>
                        {defs.length > 0 && (
                          <div className="flex flex-wrap gap-1">
                            {defs.slice(0, 3).map((entry, i) =>
                              Object.entries(entry).map(([pos, gloss]) => (
                                <span key={`${i}-${pos}`}
                                  className="flex items-center gap-0.5 text-[11px] text-muted-foreground">
                                  <span className="font-semibold text-teal-700">{POS_LABELS[pos as PosKey] ?? pos}</span>
                                  {gloss}
                                </span>
                              ))
                            )}
                          </div>
                        )}
                        {(vocab.ukPhonetic || vocab.usPhonetic) && (
                          <span className="text-[11px] text-muted-foreground">
                            {vocab.ukPhonetic && `UK /${vocab.ukPhonetic}/`}
                            {vocab.ukPhonetic && vocab.usPhonetic && "  "}
                            {vocab.usPhonetic && `US /${vocab.usPhonetic}/`}
                          </span>
                        )}
                      </div>
                    </label>
                  );
                })}
                {hasMore && (
                  <button type="button" onClick={handleLoadMore} disabled={isLoadingMore}
                    className="mt-1 flex h-9 w-full items-center justify-center rounded-lg border border-border bg-muted text-xs font-medium text-muted-foreground hover:bg-muted/80 disabled:opacity-50">
                    {isLoadingMore ? <Loader2 className="h-4 w-4 animate-spin" /> : "加载更多"}
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="flex shrink-0 items-center justify-between border-t border-border px-6 py-4">
            <span className={`text-xs font-medium ${batchSize > 0 && selected.size > 0 && selected.size % batchSize !== 0 ? "text-amber-600" : "text-muted-foreground"}`}>
              {batchHint}
              {batchSize > 0 && <span className="ml-1 text-muted-foreground">(每批 {batchSize} 的倍数)</span>}
            </span>
            <div className="flex gap-2">
              <button type="button" onClick={onClose}
                className="flex h-10 items-center gap-1.5 rounded-xl border border-border bg-muted px-4">
                <span className="text-xs font-semibold text-muted-foreground">取消</span>
              </button>
              <button type="button" onClick={handleConfirm} disabled={isConfirmDisabled}
                className="flex h-10 items-center gap-1.5 rounded-xl bg-teal-600 px-5 disabled:opacity-50">
                {isSaving ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : null}
                <span className="text-sm font-semibold text-white">确认</span>
              </button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
