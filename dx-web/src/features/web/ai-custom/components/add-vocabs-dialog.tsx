"use client";

import { useState, useTransition } from "react";
import {
  BookOpen,
  X,
  Wand2,
  Save,
  Loader2,
  CircleCheck,
  CircleDashed,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { toast } from "sonner";
import type { GameMode } from "@/consts/game-mode";
import type { LevelVocabData, AddedGameVocab } from "@/lib/api-client";
import { addGameVocabsAction } from "@/features/web/ai-custom/actions/game-vocab.action";
import { formatVocab } from "@/features/web/ai-custom/helpers/vocab-format-api";
import { vocabBatchSize } from "@/features/web/ai-custom/helpers/vocab-format-metadata";
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";

const VOCAB_REGEX = /^[A-Za-z0-9' -]+$/;

function parseEntries(text: string): { ok: true; entries: string[] } | { ok: false; error: string } {
  const lines = text
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);

  if (lines.length === 0) return { ok: false, error: "请输入至少一个词汇" };

  for (const line of lines) {
    if (!VOCAB_REGEX.test(line)) {
      return { ok: false, error: `「${line}」含有不允许的字符，仅支持英文字母、数字、撇号、连字符和空格` };
    }
  }

  return { ok: true, entries: lines };
}

type AddVocabsDialogProps = {
  gameId: string;
  levelId: string;
  gameMode: GameMode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: (added: LevelVocabData[]) => void;
};

export function AddVocabsDialog({
  gameId,
  levelId,
  gameMode,
  open,
  onOpenChange,
  onSuccess,
}: AddVocabsDialogProps) {
  const [text, setText] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [isFormatting, setIsFormatting] = useState(false);
  const [addedResults, setAddedResults] = useState<AddedGameVocab[]>([]);
  const [isPending, startTransition] = useTransition();
  const [beanDialogOpen, setBeanDialogOpen] = useState(false);
  const [beanRequired, setBeanRequired] = useState(0);
  const [beanAvailable, setBeanAvailable] = useState(0);

  const batchSize = vocabBatchSize(gameMode);

  function handleOpenChange(newOpen: boolean) {
    if (!newOpen) {
      setText("");
      setErrorMessage("");
      setIsFormatting(false);
      setAddedResults([]);
    }
    onOpenChange(newOpen);
  }

  async function handleFormat() {
    const raw = text.trim();
    if (!raw) return;

    setIsFormatting(true);
    const result = await formatVocab(raw);
    setIsFormatting(false);

    if (!result.ok) {
      if (result.code === "INSUFFICIENT_BEANS") {
        setBeanRequired(result.required ?? 0);
        setBeanAvailable(result.available ?? 0);
        setBeanDialogOpen(true);
        return;
      }
      setErrorMessage(result.message);
      toast.warning(result.message);
      return;
    }

    setText(result.formatted);
    setErrorMessage("");
    toast.success("格式化完成");
  }

  function handleSave() {
    const parsed = parseEntries(text);
    if (!parsed.ok) {
      setErrorMessage(parsed.error);
      toast.error(parsed.error);
      return;
    }

    const { entries } = parsed;

    if (batchSize > 0 && entries.length % batchSize !== 0) {
      const msg = `词汇数量必须是 ${batchSize} 的倍数（当前 ${entries.length} 个）`;
      setErrorMessage(msg);
      toast.error(msg);
      return;
    }

    setErrorMessage("");

    startTransition(async () => {
      const res = await addGameVocabsAction(gameId, levelId, entries);

      if (res.code !== 0) {
        setErrorMessage(res.message);
        toast.error(res.message);
        return;
      }

      const added = res.data ?? [];
      setAddedResults(added);

      // Build LevelVocabData placeholders so parent can append rows
      // (full vocab data will come from next list fetch; we pass minimal here)
      const levelVocabs = added.map((item, i) => ({
        gameVocabId: item.gameVocabId,
        order: Date.now() + i,
        vocab: {
          id: item.contentVocabId,
          content: item.content,
          isVerified: false,
        },
      }));

      onSuccess(levelVocabs as import("@/lib/api-client").LevelVocabData[]);
    });
  }

  const canFormat = text.trim().length > 0 && !isFormatting && !isPending;
  const canSave = text.trim().length > 0 && !isPending && !isFormatting;

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent
          aria-describedby={undefined}
          showCloseButton={false}
          className="sm:max-w-lg overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
        >
          <VisuallyHidden>
            <DialogTitle>添加词汇到关卡</DialogTitle>
          </VisuallyHidden>

          <div className="flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4">
              <div className="flex items-center gap-2.5">
                <BookOpen className="h-5 w-5 text-teal-600" />
                <h2 className="text-lg font-bold text-foreground">添加词汇</h2>
              </div>
              <button
                type="button"
                onClick={() => handleOpenChange(false)}
                aria-label="关闭"
                className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted"
              >
                <X className="h-3.5 w-3.5 text-muted-foreground" />
              </button>
            </div>

            {/* Body */}
            {addedResults.length > 0 ? (
              <div className="px-6 pb-2">
                <p className="mb-3 text-sm font-medium text-foreground">已添加 {addedResults.length} 个词汇：</p>
                <div className="flex flex-col gap-1.5 max-h-64 overflow-y-auto">
                  {addedResults.map((item) => (
                    <div
                      key={item.gameVocabId}
                      className="flex items-center justify-between rounded-lg bg-muted px-3 py-2"
                    >
                      <span className="text-sm font-medium text-foreground">{item.content}</span>
                      {item.wasReused ? (
                        <span className="flex items-center gap-1 rounded-full bg-blue-100 px-2.5 py-0.5 text-[11px] font-semibold text-blue-700">
                          <CircleCheck className="h-3 w-3" />
                          用了已有词条
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 rounded-full bg-teal-100 px-2.5 py-0.5 text-[11px] font-semibold text-teal-700">
                          <CircleDashed className="h-3 w-3" />
                          新建词条
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="px-6 pb-2">
                <label className="mb-2 block text-sm font-medium text-foreground">
                  词汇列表
                  <span className="ml-1 text-xs font-normal text-muted-foreground">（每行一个英文词汇）</span>
                </label>
                <textarea
                  value={text}
                  onChange={(e) => { setText(e.target.value); setErrorMessage(""); }}
                  placeholder={"fast\nquick\nrun"}
                  rows={8}
                  className="w-full resize-none rounded-xl border border-border bg-muted/50 px-3 py-2.5 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                />
                {errorMessage && (
                  <p className="mt-1.5 text-xs text-red-500">{errorMessage}</p>
                )}
                {batchSize > 0 && (
                  <p className="mt-1.5 text-xs text-muted-foreground">
                    当前模式每批需要 {batchSize} 的倍数个词汇
                  </p>
                )}
              </div>
            )}

            {/* Footer */}
            <div className="flex gap-3 px-6 pb-6 pt-3">
              {addedResults.length > 0 ? (
                <button
                  type="button"
                  onClick={() => handleOpenChange(false)}
                  className="ml-auto flex h-11 items-center justify-center gap-1.5 rounded-xl bg-teal-600 px-5"
                >
                  <span className="text-sm font-semibold text-white">完成</span>
                </button>
              ) : (
                <div className="ml-auto flex overflow-hidden rounded-xl border border-border">
                  <button
                    type="button"
                    disabled={!canFormat}
                    onClick={handleFormat}
                    className="flex h-11 items-center justify-center gap-1.5 border-r border-border bg-violet-50 px-4 disabled:opacity-50"
                  >
                    {isFormatting ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin text-violet-600" />
                    ) : (
                      <Wand2 className="h-3.5 w-3.5 text-violet-600" />
                    )}
                    <span className="text-xs font-semibold text-violet-600">AI 格式化</span>
                  </button>
                  <button
                    type="button"
                    disabled={!canSave}
                    onClick={handleSave}
                    className="flex h-11 items-center justify-center gap-1.5 bg-teal-600 px-5 disabled:opacity-50"
                  >
                    {isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin text-white" />
                    ) : (
                      <Save className="h-4 w-4 text-white" />
                    )}
                    <span className="text-sm font-semibold text-white">保存</span>
                  </button>
                </div>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <InsufficientBeansDialog
        open={beanDialogOpen}
        onOpenChange={setBeanDialogOpen}
        required={beanRequired}
        available={beanAvailable}
      />
    </>
  );
}
