"use client";

import { useState, useTransition } from "react";
import { Sparkles, X, Loader2, Save } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { toast } from "sonner";
import type { VocabInput, CreateVocabResult } from "@/lib/api-client";
import {
  generateVocabsFromKeywordsAction,
  createVocabsBatchAction,
} from "@/features/web/ai-custom/actions/content-vocab.action";

type AddVocabFromAiDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdded: () => void;
};

export function AddVocabFromAiDialog({ open, onOpenChange, onAdded }: AddVocabFromAiDialogProps) {
  const [keywords, setKeywords] = useState<string[]>(["", "", "", "", ""]);
  const [isGenerating, setIsGenerating] = useState(false);
  const [previewItems, setPreviewItems] = useState<VocabInput[] | null>(null);
  const [checkedIndices, setCheckedIndices] = useState<Set<number>>(new Set());
  const [isSaving, startSaveTransition] = useTransition();

  function handleOpenChange(newOpen: boolean) {
    if (!newOpen) {
      setKeywords(["", "", "", "", ""]);
      setIsGenerating(false);
      setPreviewItems(null);
      setCheckedIndices(new Set());
    }
    onOpenChange(newOpen);
  }

  function updateKeyword(index: number, value: string) {
    setKeywords((prev) => prev.map((k, i) => (i === index ? value : k)));
  }

  function toggleCheck(index: number) {
    setCheckedIndices((prev) => {
      const next = new Set(prev);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  }

  async function handleGenerate() {
    const trimmed = keywords.map((k) => k.trim()).filter((k) => k.length > 0);
    if (trimmed.length === 0) { toast.error("请至少输入一个关键词"); return; }

    setIsGenerating(true);
    setPreviewItems(null);
    try {
      const res = await generateVocabsFromKeywordsAction(trimmed);
      if (res.code !== 0) { toast.error(res.message); return; }

      let parsed: VocabInput[];
      try {
        parsed = JSON.parse(res.data) as VocabInput[];
        if (!Array.isArray(parsed)) throw new Error("not array");
      } catch {
        toast.error("AI 返回格式异常，请重试");
        return;
      }

      setPreviewItems(parsed);
      setCheckedIndices(new Set(parsed.map((_, i) => i)));
    } finally {
      setIsGenerating(false);
    }
  }

  function handleSave() {
    if (!previewItems || checkedIndices.size === 0) return;
    const inputs = previewItems.filter((_, i) => checkedIndices.has(i));

    startSaveTransition(async () => {
      const res = await createVocabsBatchAction(inputs);
      if (res.code !== 0) { toast.error(res.message); return; }

      const results: CreateVocabResult[] = res.data ?? [];
      const reused = results.filter((r) => r.wasReused).length;
      toast.success(`已添加 ${results.length} 条${reused > 0 ? `（${reused} 条复用了已有词条）` : ""}`);
      onAdded();
      handleOpenChange(false);
    });
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        aria-describedby={undefined}
        showCloseButton={false}
        className="sm:max-w-lg overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
      >
        <VisuallyHidden><DialogTitle>AI 生成词汇</DialogTitle></VisuallyHidden>

        <div className="flex flex-col max-h-[90vh]">
          <div className="flex shrink-0 items-center justify-between px-6 py-4">
            <div className="flex items-center gap-2.5">
              <Sparkles className="h-5 w-5 text-violet-600" />
              <h2 className="text-lg font-bold text-foreground">AI 生成词汇</h2>
            </div>
            <button type="button" onClick={() => handleOpenChange(false)} aria-label="关闭"
              className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted">
              <X className="h-3.5 w-3.5 text-muted-foreground" />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto px-6 pb-6 flex flex-col gap-5">
            {!previewItems ? (
              <section>
                <p className="mb-2 text-sm font-semibold text-foreground">
                  关键词
                  <span className="ml-1 text-xs font-normal text-muted-foreground">（输入英文词汇主题，AI 将据此生成相关词汇）</span>
                </p>
                <div className="flex flex-col gap-2">
                  {keywords.map((kw, i) => (
                    <input key={i} type="text" value={kw} onChange={(e) => updateKeyword(i, e.target.value)}
                      placeholder={`关键词 ${i + 1}`}
                      className="h-9 w-full rounded-lg border border-border bg-muted/50 px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500" />
                  ))}
                </div>
              </section>
            ) : (
              <section>
                <p className="mb-2 text-sm font-semibold text-foreground">
                  生成结果
                  <span className="ml-1 text-xs font-normal text-muted-foreground">（共 {previewItems.length} 条，勾选后保存）</span>
                </p>
                <div className="flex flex-col gap-1.5 max-h-80 overflow-y-auto">
                  {previewItems.map((item, i) => (
                    <label key={i}
                      className="flex cursor-pointer items-center gap-3 rounded-lg border border-border bg-background px-3 py-2 hover:bg-muted/50">
                      <input type="checkbox" checked={checkedIndices.has(i)} onChange={() => toggleCheck(i)}
                        className="h-4 w-4 accent-violet-600" />
                      <div className="flex flex-1 flex-col gap-0.5">
                        <span className="text-sm font-medium text-foreground">{item.content}</span>
                        {Array.isArray(item.definition) && item.definition.length > 0 && (
                          <span className="text-xs text-muted-foreground">
                            {item.definition.map((d) => Object.entries(d).map(([pos, gloss]) => `${pos} ${gloss}`).join("; ")).join(" | ")}
                          </span>
                        )}
                      </div>
                    </label>
                  ))}
                </div>
              </section>
            )}
          </div>

          <div className="flex shrink-0 gap-3 px-6 pb-6 pt-2">
            {!previewItems ? (
              <button type="button" onClick={handleGenerate} disabled={isGenerating}
                className="ml-auto flex h-11 items-center gap-1.5 rounded-xl bg-violet-600 px-5 disabled:opacity-50">
                {isGenerating ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : <Sparkles className="h-4 w-4 text-white" />}
                <span className="text-sm font-semibold text-white">AI 生成</span>
              </button>
            ) : (
              <div className="ml-auto flex gap-2">
                <button type="button" onClick={() => setPreviewItems(null)}
                  className="flex h-11 items-center gap-1.5 rounded-xl border border-border bg-muted px-4">
                  <span className="text-xs font-semibold text-muted-foreground">重新生成</span>
                </button>
                <button type="button" onClick={handleSave}
                  disabled={isSaving || checkedIndices.size === 0}
                  className="flex h-11 items-center gap-1.5 rounded-xl bg-teal-600 px-5 disabled:opacity-50">
                  {isSaving ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : <Save className="h-4 w-4 text-white" />}
                  <span className="text-sm font-semibold text-white">保存 {checkedIndices.size > 0 ? checkedIndices.size : ""} 条</span>
                </button>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
