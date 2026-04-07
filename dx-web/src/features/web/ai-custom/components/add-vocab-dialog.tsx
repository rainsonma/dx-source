"use client";

import { useState, useTransition } from "react";
import { swrMutate } from "@/lib/swr";
import {
  Database,
  X,
  PenLine,
  Sparkles,
  Save,
  RotateCcw,
  Loader2,
  Wand2,
  Play,
  ArrowLeft,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { toast } from "sonner";
import type { GameMode } from "@/consts/game-mode";
import { SOURCE_FROMS } from "@/consts/source-from";
import { VocabManualTab } from "@/features/web/ai-custom/components/vocab-manual-tab";
import { VocabAiTab, getKeywordsWarning } from "@/features/web/ai-custom/components/vocab-ai-tab";
import { parseVocabText, maxPairsForMode, MAX_METAS_PER_LEVEL } from "@/features/web/ai-custom/helpers/vocab-format-metadata";
import { saveMetadataAction } from "@/features/web/ai-custom/actions/course-game.action";
import { formatVocab } from "@/features/web/ai-custom/helpers/vocab-format-api";
import { generateVocab } from "@/features/web/ai-custom/helpers/vocab-generate-api";
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";

type Tab = "manual" | "ai";

type AddVocabDialogProps = {
  gameId: string;
  levelId: string;
  gameMode: GameMode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  existingMetaCount: number;
};

export function AddVocabDialog({
  gameId,
  levelId,
  gameMode,
  open,
  onOpenChange,
  existingMetaCount,
}: AddVocabDialogProps) {
  const [activeTab, setActiveTab] = useState<Tab>("manual");
  const [manualText, setManualText] = useState("");
  const [difficulty, setDifficulty] = useState("a1-a2");
  const [keywords, setKeywords] = useState("");
  const [isPending, startTransition] = useTransition();
  const [isFormatted, setIsFormatted] = useState(false);
  const [isFormatting, setIsFormatting] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [aiPreview, setAiPreview] = useState("");
  const [isGenerating, setIsGenerating] = useState(false);
  const [aiErrorMessage, setAiErrorMessage] = useState("");
  const [isFromAi, setIsFromAi] = useState(false);
  const [beanDialogOpen, setBeanDialogOpen] = useState(false);
  const [beanRequired, setBeanRequired] = useState(0);
  const [beanAvailable, setBeanAvailable] = useState(0);

  const maxPairs = maxPairsForMode(gameMode);

  function handleBeanError(result: { code?: string; required?: number; available?: number }) {
    if (result.code === "INSUFFICIENT_BEANS") {
      setBeanRequired(result.required ?? 0);
      setBeanAvailable(result.available ?? 0);
      setBeanDialogOpen(true);
      return true;
    }
    return false;
  }

  function handleOpenChange(open: boolean) {
    if (!open) {
      setActiveTab("manual");
      setManualText("");
      setIsFormatted(false);
      setIsFormatting(false);
      setErrorMessage("");
      setDifficulty("a1-a2");
      setKeywords("");
      setAiPreview("");
      setIsGenerating(false);
      setAiErrorMessage("");
      setIsFromAi(false);
      setBeanDialogOpen(false);
      setBeanRequired(0);
      setBeanAvailable(0);
    }
    onOpenChange(open);
  }

  function handleSave() {
    const result = parseVocabText(manualText, maxPairs);

    if (!result.ok) {
      setErrorMessage(result.error);
      toast.error(result.error);
      return;
    }

    if (existingMetaCount + result.pairs.length > MAX_METAS_PER_LEVEL) {
      const msg = `超出关卡上限：当前已有 ${existingMetaCount}/${MAX_METAS_PER_LEVEL} 条，本次新增 ${result.pairs.length} 条`;
      setErrorMessage(msg);
      toast.error(msg);
      return;
    }

    const entries = result.pairs.map((pair) => ({
      sourceData: pair.sourceData,
      translation: pair.translation,
      sourceType: "vocab" as const,
    }));

    startTransition(async () => {
      const res = await saveMetadataAction(gameId, {
        gameLevelId: levelId,
        entries,
        sourceFrom: isFromAi ? SOURCE_FROMS.AI : SOURCE_FROMS.MANUAL,
      });

      if (res.error) {
        setErrorMessage(res.error);
        toast.error(res.error);
        return;
      }

      setErrorMessage("");
      toast.success(`已保存 ${res.count ?? entries.length} 条元数据`);
      handleOpenChange(false);
      swrMutate("/api/course-games");
    });
  }

  async function handleFormat() {
    const text = manualText.trim();
    if (!text) return;

    setIsFormatting(true);
    const result = await formatVocab(text);
    setIsFormatting(false);

    if (!result.ok) {
      if (handleBeanError(result)) return;
      setErrorMessage(result.message);
      toast.warning(result.message);
      return;
    }

    setErrorMessage("");
    setManualText(result.formatted);
    setIsFormatted(true);
    toast.success("格式化完成");
  }

  async function handleGenerate() {
    const words = keywords.trim().split(/\s+/).filter(Boolean).slice(0, 5);
    if (words.length === 0) return;

    setIsGenerating(true);
    setAiErrorMessage("");
    const result = await generateVocab(difficulty, words, gameMode);
    setIsGenerating(false);

    if (!result.ok) {
      if (handleBeanError(result)) return;
      setAiErrorMessage(result.message);
      toast.warning(result.message);
      return;
    }

    setAiErrorMessage("");
    setAiPreview(result.generated);
    toast.success("词汇生成完成");
  }

  function handleAiReset() {
    setAiPreview("");
    setDifficulty("a1-a2");
    setKeywords("");
    setAiErrorMessage("");
  }

  function handleUseGenerated() {
    setManualText(aiPreview);
    setIsFormatted(true);
    setIsFromAi(true);
    setErrorMessage("");
    setActiveTab("manual");
    toast.success("已导入到手动添加，可直接保存或格式化后保存");
  }

  const canSave = manualText.trim().length > 0 && !isPending && !isFormatting && (isFormatted || isFromAi);
  const canFormat = manualText.trim().length > 0 && !isPending && !isFormatting;
  const canGenerate = keywords.trim().length > 0 && !isGenerating && !getKeywordsWarning(keywords);
  const canReset = aiPreview.length > 0 || keywords.trim().length > 0;
  const canUse = aiPreview.trim().length > 0 && !isGenerating;

  return (
    <>
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        aria-describedby={undefined}
        showCloseButton={false}
        className="sm:max-w-3xl overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
      >
        <VisuallyHidden>
          <DialogTitle>添加词汇</DialogTitle>
        </VisuallyHidden>

        <div className="flex flex-col">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4">
            <div className="flex items-center gap-2.5">
              <Database className="h-5 w-5 text-teal-600" />
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

          {/* Tabs */}
          <div className="relative px-6">
            <div className="absolute bottom-0 left-6 right-6 h-0.5 bg-border" />
            <div className="relative flex" role="tablist">
              <button
                type="button"
                role="tab"
                aria-selected={activeTab === "manual"}
                onClick={() => setActiveTab("manual")}
                className={`flex items-center gap-1.5 border-b-2 px-5 py-2.5 transition-colors ${
                  activeTab === "manual"
                    ? "border-teal-600"
                    : "border-transparent"
                }`}
              >
                <PenLine
                  className={`h-3.5 w-3.5 ${activeTab === "manual" ? "text-teal-600" : "text-muted-foreground"}`}
                />
                <span
                  className={`text-sm ${
                    activeTab === "manual"
                      ? "font-semibold text-teal-600"
                      : "font-medium text-muted-foreground"
                  }`}
                >
                  手动添加
                </span>
              </button>
              <button
                type="button"
                role="tab"
                aria-selected={activeTab === "ai"}
                onClick={() => setActiveTab("ai")}
                className={`flex items-center gap-1.5 border-b-2 px-5 py-2.5 transition-colors ${
                  activeTab === "ai"
                    ? "border-teal-600"
                    : "border-transparent"
                }`}
              >
                <Sparkles
                  className={`h-3.5 w-3.5 ${activeTab === "ai" ? "text-teal-600" : "text-muted-foreground"}`}
                />
                <span
                  className={`text-sm ${
                    activeTab === "ai"
                      ? "font-semibold text-teal-600"
                      : "font-medium text-muted-foreground"
                  }`}
                >
                  AI 生成
                </span>
              </button>
            </div>
          </div>

          {/* Tab content */}
          {activeTab === "manual" ? (
            <VocabManualTab
              value={manualText}
              onChange={(v) => { setManualText(v); setErrorMessage(""); setIsFromAi(false); }}
              error={errorMessage}
              maxPairs={maxPairs}
            />
          ) : (
            <VocabAiTab
              difficulty={difficulty}
              onDifficultyChange={setDifficulty}
              keywords={keywords}
              onKeywordsChange={setKeywords}
              preview={aiPreview}
              error={aiErrorMessage}
            />
          )}

          {/* Footer */}
          <div className="flex gap-3 px-6 pb-6 pt-3">
            {activeTab === "manual" ? (
              <div className="flex ml-auto overflow-hidden rounded-xl border border-border">
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
                  <span className="text-xs font-semibold text-violet-600">
                    词汇检查并格式化
                  </span>
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
            ) : (
              <div className="flex ml-auto overflow-hidden rounded-xl border border-border">
                <button
                  type="button"
                  disabled={!canReset}
                  onClick={handleAiReset}
                  className="flex h-11 items-center justify-center gap-1.5 border-r border-border bg-muted px-4 disabled:opacity-50"
                >
                  <RotateCcw className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="text-xs font-semibold text-muted-foreground">
                    重置
                  </span>
                </button>
                <button
                  type="button"
                  disabled={!canGenerate}
                  onClick={handleGenerate}
                  className="flex h-11 items-center justify-center gap-1.5 border-r border-border bg-teal-50 px-4 disabled:opacity-50"
                >
                  {isGenerating ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin text-teal-600" />
                  ) : (
                    <Play className="h-3.5 w-3.5 text-teal-600" />
                  )}
                  <span className="text-xs font-semibold text-teal-600">
                    {isGenerating ? "生成中..." : aiPreview ? "重新生成" : "AI 生成"}
                  </span>
                </button>
                <button
                  type="button"
                  disabled={!canUse}
                  onClick={handleUseGenerated}
                  className="flex h-11 items-center justify-center gap-1.5 bg-teal-600 px-5 disabled:opacity-50"
                >
                  <ArrowLeft className="h-4 w-4 text-white" />
                  <span className="text-sm font-semibold text-white">使用</span>
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
