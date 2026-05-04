"use client";

import { useState, useCallback, useRef } from "react";
import {
  BookOpen,
  Plus,
  Trash2,
  Volume2,
  ShieldAlert,
  TriangleAlert,
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
import type { PosKey, LevelVocabData, AddedGameVocab } from "@/lib/api-client";
import {
  deleteGameVocabAction,
  listGameVocabsAction,
} from "@/features/web/ai-custom/actions/game-vocab.action";
import { SelectVocabsDialog } from "@/features/web/ai-custom/components/select-vocabs-dialog";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

const POS_LABELS: Partial<Record<PosKey, string>> = {
  n: "n",
  v: "v",
  adj: "adj",
  adv: "adv",
  prep: "prep",
  conj: "conj",
  pron: "pron",
  art: "art",
  num: "num",
  int: "int",
  aux: "aux",
  det: "det",
};

type DefinitionMap = Partial<Record<PosKey, string>>;

function parseDefinition(raw: string | null | undefined): DefinitionMap[] {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) return parsed as DefinitionMap[];
    return [];
  } catch {
    return [];
  }
}

type LevelVocabsPanelProps = {
  gameId: string;
  levelId: string;
  gameMode: GameMode;
  initialVocabs: LevelVocabData[];
  readOnly?: boolean;
};

export function LevelVocabsPanel({
  gameId,
  levelId,
  gameMode,
  initialVocabs,
  readOnly,
}: LevelVocabsPanelProps) {
  const [vocabs, setVocabs] = useState<LevelVocabData[]>(initialVocabs);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const [playingUrl, setPlayingUrl] = useState<string | null>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const handleRequestDelete = useCallback((gvId: string) => {
    setPendingDeleteId(gvId);
    setDeleteConfirmOpen(true);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    if (!pendingDeleteId) return;
    setDeleteConfirmOpen(false);

    const prev = vocabs;
    setVocabs((v) => v.filter((row) => row.gameVocabId !== pendingDeleteId));

    const result = await deleteGameVocabAction(gameId, pendingDeleteId);
    if (result.code !== 0) {
      setVocabs(prev);
      toast.error(result.message);
    }
    setPendingDeleteId(null);
  }, [pendingDeleteId, vocabs, gameId]);

  function handlePlayAudio(url: string) {
    if (playingUrl === url) {
      audioRef.current?.pause();
      setPlayingUrl(null);
      return;
    }
    if (audioRef.current) {
      audioRef.current.pause();
    }
    const audio = new Audio(`${API_URL}${url}`);
    audioRef.current = audio;
    setPlayingUrl(url);
    audio.onended = () => setPlayingUrl(null);
    audio.onerror = () => setPlayingUrl(null);
    audio.play().catch(() => setPlayingUrl(null));
  }

  async function handlePickerAdded(added: AddedGameVocab[]) {
    setPickerOpen(false);
    // Re-fetch the level's vocabs to pick up the full LevelVocabData
    // (AddedGameVocab only carries id+content; the picker dialog returns the
    // backend's add-batch response shape).
    const result = await listGameVocabsAction(gameId, levelId);
    if (result.code === 0 && result.data) {
      setVocabs(result.data);
    }
    toast.success(`已添加 ${added.length} 个词汇`);
  }

  const placedVocabIds = vocabs
    .map((row) => row.vocab?.id)
    .filter((id): id is string => Boolean(id));

  return (
    <>
      <div className="flex flex-1 flex-col gap-3 overflow-hidden">
        {readOnly && (
          <div className="flex shrink-0 items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm font-medium text-amber-700">
            <ShieldAlert className="h-4 w-4 shrink-0" />
            已发布的游戏内容不可编辑，如需修改请先撤回游戏
          </div>
        )}

        <div className="flex flex-1 flex-col overflow-hidden rounded-[14px] border border-border bg-card">
          {/* Header */}
          <div className="shrink-0 p-4 pb-0 lg:p-5 lg:pb-0">
            <div className="flex flex-col items-start justify-between gap-3 pb-2 pt-2 sm:flex-row sm:items-center">
              <div className="flex items-center gap-2">
                <BookOpen className="h-4 w-4 text-teal-600" />
                <span className="text-[15px] font-bold text-foreground">词汇列表</span>
                <span className="rounded-[10px] bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
                  {vocabs.length} 个
                </span>
              </div>

              <button
                type="button"
                onClick={() => setPickerOpen(true)}
                disabled={readOnly}
                title={readOnly ? "已发布的游戏不可编辑" : "从我的词库中选择"}
                className="flex items-center gap-1 rounded-lg bg-teal-600 px-3 py-1.5 disabled:opacity-50"
              >
                <Plus className="h-3.5 w-3.5 text-white" />
                <span className="text-xs font-semibold text-white">选择词汇</span>
              </button>
            </div>
          </div>

          {/* List */}
          <div className="flex-1 overflow-y-auto px-4 pb-4 pt-3 lg:px-5 lg:pb-5">
            {vocabs.length === 0 ? (
              <div className="flex flex-col items-center justify-center gap-2 py-16 text-muted-foreground">
                <BookOpen className="h-8 w-8" />
                <span className="text-sm">暂无词汇，从「AI 词汇库」选择已添加的词条</span>
              </div>
            ) : (
              <div className="flex flex-col gap-2">
                {vocabs.map((row) => {
                  const vocab = row.vocab;
                  const defs = parseDefinition(vocab?.definition);

                  return (
                    <div
                      key={row.gameVocabId}
                      className="flex flex-col gap-2 rounded-xl border border-border bg-background p-3"
                    >
                      {/* Top row: word + delete */}
                      <div className="flex items-start justify-between gap-2">
                        <span className="text-[15px] font-bold text-foreground">
                          {vocab?.content ?? <span className="text-muted-foreground italic">词条缺失</span>}
                        </span>
                        <button
                          type="button"
                          onClick={() => handleRequestDelete(row.gameVocabId)}
                          disabled={readOnly}
                          title="移除"
                          className="flex h-7 w-7 items-center justify-center rounded-lg bg-red-50 text-red-500 hover:bg-red-100 disabled:opacity-50"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>

                      {/* Definition pills */}
                      {defs.length > 0 && (
                        <div className="flex flex-wrap gap-1.5">
                          {defs.map((entry, i) =>
                            Object.entries(entry).map(([pos, gloss]) => (
                              <span
                                key={`${i}-${pos}`}
                                className="flex items-center gap-1 rounded-full bg-teal-50 px-2.5 py-0.5 text-xs"
                              >
                                <span className="font-semibold text-teal-700">
                                  {POS_LABELS[pos as PosKey] ?? pos}
                                </span>
                                <span className="text-muted-foreground">{gloss}</span>
                              </span>
                            ))
                          )}
                        </div>
                      )}

                      {/* Phonetics + audio */}
                      {vocab && (vocab.ukPhonetic || vocab.usPhonetic) && (
                        <div className="flex flex-wrap items-center gap-2">
                          {vocab.ukPhonetic && (
                            <span className="flex items-center gap-1 text-xs text-muted-foreground">
                              <span className="font-medium text-foreground">UK</span>
                              <span>{vocab.ukPhonetic}</span>
                              {vocab.ukAudioUrl && (
                                <button
                                  type="button"
                                  onClick={() => handlePlayAudio(vocab.ukAudioUrl!)}
                                  className={`flex h-5 w-5 items-center justify-center rounded-full ${playingUrl === vocab.ukAudioUrl ? "bg-teal-600 text-white" : "bg-muted text-muted-foreground hover:bg-teal-100 hover:text-teal-700"}`}
                                >
                                  <Volume2 className="h-2.5 w-2.5" />
                                </button>
                              )}
                            </span>
                          )}
                          {vocab.usPhonetic && (
                            <span className="flex items-center gap-1 text-xs text-muted-foreground">
                              <span className="font-medium text-foreground">US</span>
                              <span>{vocab.usPhonetic}</span>
                              {vocab.usAudioUrl && (
                                <button
                                  type="button"
                                  onClick={() => handlePlayAudio(vocab.usAudioUrl!)}
                                  className={`flex h-5 w-5 items-center justify-center rounded-full ${playingUrl === vocab.usAudioUrl ? "bg-teal-600 text-white" : "bg-muted text-muted-foreground hover:bg-teal-100 hover:text-teal-700"}`}
                                >
                                  <Volume2 className="h-2.5 w-2.5" />
                                </button>
                              )}
                            </span>
                          )}
                        </div>
                      )}

                      {/* Explanation */}
                      {vocab?.explanation && (
                        <p className="text-xs text-muted-foreground">{vocab.explanation}</p>
                      )}

                      {/* Warning if vocab is null (canonical row deleted from user pool) */}
                      {!vocab && (
                        <div className="flex items-center gap-1 text-xs text-amber-600">
                          <TriangleAlert className="h-3 w-3" />
                          该词条已从词库中删除，建议从本关卡移除
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </div>

      {pickerOpen && (
        <SelectVocabsDialog
          gameId={gameId}
          levelId={levelId}
          gameMode={gameMode}
          alreadyPlacedVocabIds={placedVocabIds}
          onClose={() => setPickerOpen(false)}
          onAdded={handlePickerAdded}
        />
      )}

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认移除词汇</AlertDialogTitle>
            <AlertDialogDescription>
              将从本关卡移除该词汇，不会从「AI 词汇库」中删除词条。此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setPendingDeleteId(null)}>取消</AlertDialogCancel>
            <AlertDialogAction className="bg-red-600 hover:bg-red-700" onClick={handleConfirmDelete}>
              确认移除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
