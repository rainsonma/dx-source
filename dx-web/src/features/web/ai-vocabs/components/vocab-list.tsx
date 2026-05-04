"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import { BookOpen, Trash2, PenLine, Volume2, Loader2 } from "lucide-react";
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
import type { ContentVocabData, PosKey } from "@/lib/api-client";
import {
  listMyVocabsAction,
  deleteVocabAction,
} from "@/features/web/ai-custom/actions/content-vocab.action";
import { VocabEditDialog } from "@/features/web/ai-vocabs/components/vocab-edit-dialog";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

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

type VocabListProps = {
  search: string;
  refreshKey: number;
};

const PAGE_SIZE = 30;

export function VocabList({ search, refreshKey }: VocabListProps) {
  const [items, setItems] = useState<ContentVocabData[]>([]);
  const [cursor, setCursor] = useState<string>("");
  const [hasMore, setHasMore] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [editVocab, setEditVocab] = useState<ContentVocabData | null>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const [playingUrl, setPlayingUrl] = useState<string | null>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  async function loadPage(nextCursor: string, reset: boolean) {
    if (reset) setIsLoading(true); else setIsLoadingMore(true);
    try {
      const res = await listMyVocabsAction({
        cursor: nextCursor || undefined,
        search: search || undefined,
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
    loadPage("", true);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search, refreshKey]);

  const loadMore = useCallback(async () => {
    if (isLoadingMore || !hasMore) return;
    await loadPage(cursor, false);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoadingMore, hasMore, cursor]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => { if (entries[0].isIntersecting && hasMore && !isLoadingMore) loadMore(); },
      { rootMargin: "200px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasMore, isLoadingMore, loadMore]);

  function handlePlayAudio(url: string) {
    if (playingUrl === url) {
      audioRef.current?.pause();
      setPlayingUrl(null);
      return;
    }
    if (audioRef.current) audioRef.current.pause();
    const audio = new Audio(`${API_URL}${url}`);
    audioRef.current = audio;
    setPlayingUrl(url);
    audio.onended = () => setPlayingUrl(null);
    audio.onerror = () => setPlayingUrl(null);
    audio.play().catch(() => setPlayingUrl(null));
  }

  const handleRequestDelete = useCallback((id: string) => {
    setPendingDeleteId(id);
    setDeleteConfirmOpen(true);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    if (!pendingDeleteId) return;
    setDeleteConfirmOpen(false);
    const prev = items;
    setItems((v) => v.filter((i) => i.id !== pendingDeleteId));
    const result = await deleteVocabAction(pendingDeleteId);
    if (result.code !== 0) {
      setItems(prev);
      toast.error(result.message);
    } else {
      toast.success("已删除");
    }
    setPendingDeleteId(null);
  }, [pendingDeleteId, items]);

  function handleEditSaved(updated: ContentVocabData) {
    setItems((prev) => prev.map((i) => (i.id === updated.id ? updated : i)));
    setEditVocab(null);
  }

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 py-16 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <span className="text-sm">加载中...</span>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 py-16 text-muted-foreground">
        <BookOpen className="h-8 w-8" />
        <span className="text-sm">还没有词汇 — 使用 AI 生成 或 手动添加</span>
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-col gap-2">
        {items.map((vocab) => {
          const defs = parseDefinition(vocab.definition);
          return (
            <div key={vocab.id}
              className="flex flex-col gap-2 rounded-xl border border-border bg-background p-3">
              <div className="flex items-start justify-between gap-2">
                <span className="text-[15px] font-bold text-foreground">{vocab.content}</span>
                <div className="flex shrink-0 items-center gap-1.5">
                  <button type="button" onClick={() => setEditVocab(vocab)}
                    className="flex h-7 items-center gap-1 rounded-lg bg-muted px-2 text-xs font-semibold text-muted-foreground hover:bg-blue-50 hover:text-blue-700">
                    <PenLine className="h-3 w-3" />编辑
                  </button>
                  <button type="button" onClick={() => handleRequestDelete(vocab.id)}
                    className="flex h-7 w-7 items-center justify-center rounded-lg bg-red-50 text-red-500 hover:bg-red-100">
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>

              {defs.length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {defs.map((entry, i) =>
                    Object.entries(entry).map(([pos, gloss]) => (
                      <span key={`${i}-${pos}`}
                        className="flex items-center gap-1 rounded-full bg-teal-50 px-2.5 py-0.5 text-xs">
                        <span className="font-semibold text-teal-700">{POS_LABELS[pos as PosKey] ?? pos}</span>
                        <span className="text-muted-foreground">{gloss}</span>
                      </span>
                    ))
                  )}
                </div>
              )}

              {(vocab.ukPhonetic || vocab.usPhonetic) && (
                <div className="flex flex-wrap items-center gap-2">
                  {vocab.ukPhonetic && (
                    <span className="flex items-center gap-1 text-xs text-muted-foreground">
                      <span className="font-medium text-foreground">UK</span>
                      <span>/{vocab.ukPhonetic}/</span>
                      {vocab.ukAudioUrl && (
                        <button type="button" onClick={() => handlePlayAudio(vocab.ukAudioUrl!)}
                          className={`flex h-5 w-5 items-center justify-center rounded-full ${playingUrl === vocab.ukAudioUrl ? "bg-teal-600 text-white" : "bg-muted text-muted-foreground hover:bg-teal-100 hover:text-teal-700"}`}>
                          <Volume2 className="h-2.5 w-2.5" />
                        </button>
                      )}
                    </span>
                  )}
                  {vocab.usPhonetic && (
                    <span className="flex items-center gap-1 text-xs text-muted-foreground">
                      <span className="font-medium text-foreground">US</span>
                      <span>/{vocab.usPhonetic}/</span>
                      {vocab.usAudioUrl && (
                        <button type="button" onClick={() => handlePlayAudio(vocab.usAudioUrl!)}
                          className={`flex h-5 w-5 items-center justify-center rounded-full ${playingUrl === vocab.usAudioUrl ? "bg-teal-600 text-white" : "bg-muted text-muted-foreground hover:bg-teal-100 hover:text-teal-700"}`}>
                          <Volume2 className="h-2.5 w-2.5" />
                        </button>
                      )}
                    </span>
                  )}
                </div>
              )}

              {vocab.explanation && (
                <p className="text-xs text-muted-foreground">{vocab.explanation}</p>
              )}
            </div>
          );
        })}
      </div>

      <div ref={sentinelRef} className="py-2 text-center">
        {isLoadingMore && <Loader2 className="mx-auto h-5 w-5 animate-spin text-muted-foreground" />}
      </div>

      {editVocab && (
        <VocabEditDialog vocab={editVocab} onClose={() => setEditVocab(null)} onSaved={handleEditSaved} />
      )}

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除词汇</AlertDialogTitle>
            <AlertDialogDescription>将从您的词汇库删除该词汇。此操作不可撤销。</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setPendingDeleteId(null)}>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-red-600 hover:bg-red-700" onClick={handleConfirmDelete}>
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
