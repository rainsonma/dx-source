"use client";

import { useState, useTransition } from "react";
import { BookOpen, X, Wand2, Save, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { toast } from "sonner";
import type { VocabInput, CreateVocabResult } from "@/lib/api-client";
import { createVocabsBatchAction } from "@/features/web/ai-custom/actions/content-vocab.action";
import { formatVocab } from "@/features/web/ai-custom/helpers/vocab-format-api";
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";

type AddVocabManualDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdded: () => void;
};

export function AddVocabManualDialog({ open, onOpenChange, onAdded }: AddVocabManualDialogProps) {
  const [text, setText] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [isFormatting, setIsFormatting] = useState(false);
  const [isSaving, startSaveTransition] = useTransition();
  const [beanDialogOpen, setBeanDialogOpen] = useState(false);
  const [beanRequired, setBeanRequired] = useState(0);
  const [beanAvailable, setBeanAvailable] = useState(0);

  function handleOpenChange(newOpen: boolean) {
    if (!newOpen) {
      setText("");
      setErrorMessage("");
      setIsFormatting(false);
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
    const lines = text.split("\n").map((l) => l.trim()).filter((l) => l.length > 0);
    if (lines.length === 0) {
      setErrorMessage("请输入至少一个词汇");
      toast.error("请输入至少一个词汇");
      return;
    }

    const inputs: VocabInput[] = lines.map((line) => ({
      content: line,
      definition: [],
    }));

    setErrorMessage("");
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

  const canFormat = text.trim().length > 0 && !isFormatting && !isSaving;
  const canSave = text.trim().length > 0 && !isSaving && !isFormatting;

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent
          aria-describedby={undefined}
          showCloseButton={false}
          className="sm:max-w-lg overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
        >
          <VisuallyHidden><DialogTitle>手动添加词汇</DialogTitle></VisuallyHidden>

          <div className="flex flex-col">
            <div className="flex items-center justify-between px-6 py-4">
              <div className="flex items-center gap-2.5">
                <BookOpen className="h-5 w-5 text-teal-600" />
                <h2 className="text-lg font-bold text-foreground">手动添加词汇</h2>
              </div>
              <button type="button" onClick={() => handleOpenChange(false)} aria-label="关闭"
                className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted">
                <X className="h-3.5 w-3.5 text-muted-foreground" />
              </button>
            </div>

            <div className="px-6 pb-2">
              <label className="mb-2 block text-sm font-medium text-foreground">
                词汇列表
                <span className="ml-1 text-xs font-normal text-muted-foreground">（每行一个英文词汇，可先 AI 格式化）</span>
              </label>
              <textarea value={text} onChange={(e) => { setText(e.target.value); setErrorMessage(""); }}
                placeholder={"fast\nquick\nrun"} rows={8}
                className="w-full resize-none rounded-xl border border-border bg-muted/50 px-3 py-2.5 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500" />
              {errorMessage && <p className="mt-1.5 text-xs text-red-500">{errorMessage}</p>}
            </div>

            <div className="flex gap-3 px-6 pb-6 pt-3">
              <div className="ml-auto flex overflow-hidden rounded-xl border border-border">
                <button type="button" disabled={!canFormat} onClick={handleFormat}
                  className="flex h-11 items-center justify-center gap-1.5 border-r border-border bg-violet-50 px-4 disabled:opacity-50">
                  {isFormatting ? <Loader2 className="h-3.5 w-3.5 animate-spin text-violet-600" /> : <Wand2 className="h-3.5 w-3.5 text-violet-600" />}
                  <span className="text-xs font-semibold text-violet-600">AI 格式化</span>
                </button>
                <button type="button" disabled={!canSave} onClick={handleSave}
                  className="flex h-11 items-center justify-center gap-1.5 bg-teal-600 px-5 disabled:opacity-50">
                  {isSaving ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : <Save className="h-4 w-4 text-white" />}
                  <span className="text-sm font-semibold text-white">保存</span>
                </button>
              </div>
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
