"use client";

import { useState, useTransition } from "react";
import {
  BookOpen,
  Sparkles,
  X,
  PenLine,
  Save,
  Loader2,
  Wand2,
  Play,
  RotateCcw,
  CircleAlert,
  Gauge,
  TextCursorInput,
  Eye,
  ArrowLeft,
  Type,
  Copy,
  Trash2,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";
import { toast } from "sonner";
import type { CreateVocabResult } from "@/lib/api-client";
import {
  generateVocabWordsAction,
  createVocabsFromWordsAction,
} from "@/features/web/ai-custom/actions/content-vocab.action";
import { formatVocabWords } from "@/features/web/ai-vocabs/helpers/vocab-words-format-api";
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";
import { DIFFICULTY_OPTIONS } from "@/consts/difficulty";

type Tab = "manual" | "ai";

type AddVocabDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdded: () => void;
};

const MAX_KEYWORDS = 5;
const MAX_KEYWORD_LENGTH = 30;

function getKeywordsWarning(keywords: string): string {
  const trimmed = keywords.trim();
  if (!trimmed) return "";
  const words = trimmed.split(/\s+/).filter(Boolean);
  if (words.length === 1 && /[,，、;；/|]/.test(trimmed)) {
    return "请用空格分隔关键词";
  }
  if (words.length > MAX_KEYWORDS) {
    return `最多输入 ${MAX_KEYWORDS} 个关键词，当前 ${words.length} 个`;
  }
  const long = words.find((w) => w.length > MAX_KEYWORD_LENGTH);
  if (long) {
    return `单个关键词不能超过 ${MAX_KEYWORD_LENGTH} 个字符`;
  }
  return "";
}

export function AddVocabDialog({ open, onOpenChange, onAdded }: AddVocabDialogProps) {
  const [activeTab, setActiveTab] = useState<Tab>("manual");

  // Manual tab state
  const [manualText, setManualText] = useState("");
  const [manualError, setManualError] = useState("");
  const [isFormatting, setIsFormatting] = useState(false);
  // Save gate: must be true to enable 保存. Set by either successful format
  // or by 使用 (AI import). Any manual edit resets it back to false.
  const [isFormatted, setIsFormatted] = useState(false);
  const [isFromAi, setIsFromAi] = useState(false);

  // AI tab state
  const [difficulty, setDifficulty] = useState("a1-a2");
  const [keywords, setKeywords] = useState("");
  const [isGenerating, setIsGenerating] = useState(false);
  const [aiError, setAiError] = useState("");
  const [aiWordsText, setAiWordsText] = useState("");

  // Shared save state
  const [isSaving, startSaveTransition] = useTransition();

  // Bean dialog
  const [beanDialogOpen, setBeanDialogOpen] = useState(false);
  const [beanRequired, setBeanRequired] = useState(0);
  const [beanAvailable, setBeanAvailable] = useState(0);

  function handleOpenChange(newOpen: boolean) {
    if (!newOpen) {
      setActiveTab("manual");
      setManualText("");
      setManualError("");
      setIsFormatting(false);
      setIsFormatted(false);
      setIsFromAi(false);
      setDifficulty("a1-a2");
      setKeywords("");
      setIsGenerating(false);
      setAiError("");
      setAiWordsText("");
      setBeanDialogOpen(false);
      setBeanRequired(0);
      setBeanAvailable(0);
    }
    onOpenChange(newOpen);
  }

  function handleBeanError(result: { code?: number; required?: number; available?: number }) {
    if (result.code === 40007) {
      setBeanRequired(result.required ?? 0);
      setBeanAvailable(result.available ?? 0);
      setBeanDialogOpen(true);
      return true;
    }
    return false;
  }

  // Manual tab handlers
  async function handleFormat() {
    const raw = manualText.trim();
    if (!raw) return;
    setIsFormatting(true);
    const result = await formatVocabWords(raw);
    setIsFormatting(false);

    if (!result.ok) {
      if (result.code === "INSUFFICIENT_BEANS") {
        setBeanRequired(result.required ?? 0);
        setBeanAvailable(result.available ?? 0);
        setBeanDialogOpen(true);
        return;
      }
      setManualError(result.message);
      toast.warning(result.message);
      return;
    }

    setManualText(result.formatted);
    setManualError("");
    setIsFormatted(true);
    setIsFromAi(false);
    toast.success("格式化完成");
  }

  // AI tab handlers
  async function handleGenerate() {
    const words = keywords.trim().split(/\s+/).filter(Boolean);
    if (words.length === 0) return;

    setIsGenerating(true);
    setAiError("");
    const result = await generateVocabWordsAction(words, difficulty);
    setIsGenerating(false);

    if (result.code !== 0) {
      if (handleBeanError(result as unknown as { code?: number })) return;
      const msg = result.message ?? "生成失败，请重试";
      setAiError(msg);
      toast.warning(msg);
      return;
    }

    const generated = result.data.words ?? [];
    setAiWordsText(generated.join("\n"));
    setAiError("");
    toast.success(`已生成 ${generated.length} 个词汇`);
  }

  function handleAiReset() {
    setAiWordsText("");
    setKeywords("");
    setDifficulty("a1-a2");
    setAiError("");
  }

  function parseWordsFromText(text: string): string[] {
    const lines = text
      .split("\n")
      .map((l) => l.trim())
      .filter((l) => l.length > 0);
    const seen = new Set<string>();
    return lines.filter((l) => {
      const lc = l.toLowerCase();
      if (seen.has(lc)) return false;
      seen.add(lc);
      return true;
    });
  }

  // Save handler — only the manual tab can save directly. The AI tab uses
  // "使用" to import into the manual tab first, where the user reviews and saves.
  function handleSave() {
    const words = parseWordsFromText(manualText);
    if (words.length === 0) {
      const msg = "请输入至少一个词汇";
      setManualError(msg);
      toast.error(msg);
      return;
    }

    startSaveTransition(async () => {
      const result = await createVocabsFromWordsAction(words);
      if (result.code !== 0) {
        if (handleBeanError(result as unknown as { code?: number })) return;
        toast.error(result.message ?? "保存失败");
        return;
      }

      const results: CreateVocabResult[] = result.data ?? [];
      const reused = results.filter((r) => r.wasReused).length;
      const added = results.length - reused;
      const parts: string[] = [];
      if (added > 0) parts.push(`已添加 ${added} 条`);
      if (reused > 0) parts.push(`${reused} 条复用了已有词条`);
      toast.success(parts.join("，") || "操作完成");
      onAdded();
      handleOpenChange(false);
    });
  }

  const keywordsWarning = getKeywordsWarning(keywords);
  const canGenerate = keywords.trim().length > 0 && !isGenerating && !keywordsWarning;
  const canReset = aiWordsText.length > 0 || keywords.trim().length > 0;
  const aiWordCount = parseWordsFromText(aiWordsText).length;
  const canUse = aiWordCount > 0 && !isGenerating;
  const canSave = manualText.trim().length > 0 && !isSaving && !isFormatting && (isFormatted || isFromAi);
  const canFormat = manualText.trim().length > 0 && !isFormatting && !isSaving;

  function handleUseGenerated() {
    if (aiWordCount === 0) return;
    // Mirrors metadata dialog: 使用 OVERWRITES the manual textarea with the
    // AI-generated content (which is already validated/clean). Sets isFromAi
    // to bypass the format gate, since AI output doesn't need format step.
    setManualText(aiWordsText);
    setManualError("");
    setIsFormatted(false);
    setIsFromAi(true);
    setActiveTab("manual");
    toast.success("已导入到手动添加，可直接保存或编辑后保存");
  }

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent
          aria-describedby={undefined}
          showCloseButton={false}
          className="sm:max-w-2xl overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
        >
          <VisuallyHidden>
            <DialogTitle>添加词汇</DialogTitle>
          </VisuallyHidden>

          <div className="flex flex-col max-h-[90vh]">
            {/* Header */}
            <div className="flex shrink-0 items-center justify-between px-6 py-4">
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

            {/* Tabs */}
            <div className="relative shrink-0 px-6">
              <div className="absolute bottom-0 left-6 right-6 h-0.5 bg-border" />
              <div className="relative flex" role="tablist">
                <button
                  type="button"
                  role="tab"
                  aria-selected={activeTab === "manual"}
                  onClick={() => setActiveTab("manual")}
                  className={`flex items-center gap-1.5 border-b-2 px-5 py-2.5 transition-colors ${
                    activeTab === "manual" ? "border-teal-600" : "border-transparent"
                  }`}
                >
                  <PenLine className={`h-3.5 w-3.5 ${activeTab === "manual" ? "text-teal-600" : "text-muted-foreground"}`} />
                  <span className={`text-sm ${activeTab === "manual" ? "font-semibold text-teal-600" : "font-medium text-muted-foreground"}`}>
                    手动添加
                  </span>
                </button>
                <button
                  type="button"
                  role="tab"
                  aria-selected={activeTab === "ai"}
                  onClick={() => setActiveTab("ai")}
                  className={`flex items-center gap-1.5 border-b-2 px-5 py-2.5 transition-colors ${
                    activeTab === "ai" ? "border-teal-600" : "border-transparent"
                  }`}
                >
                  <Sparkles className={`h-3.5 w-3.5 ${activeTab === "ai" ? "text-teal-600" : "text-muted-foreground"}`} />
                  <span className={`text-sm ${activeTab === "ai" ? "font-semibold text-teal-600" : "font-medium text-muted-foreground"}`}>
                    AI 生成
                  </span>
                </button>
              </div>
            </div>

            {/* Tab content */}
            <div className="flex-1 overflow-y-auto">
              {activeTab === "manual" ? (
                <div className="flex flex-col gap-3 px-6 py-4">
                  <p className="text-xs text-muted-foreground">
                    每行输入一个英文词汇，保存后 AI 将自动补全音标和释义。
                  </p>

                  {/* Toolbar: format example (hover) + copy/clear */}
                  <div className="flex items-center gap-2">
                    <HoverCard openDelay={200}>
                      <HoverCardTrigger asChild>
                        <button
                          type="button"
                          className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent"
                        >
                          <Type className="h-3.5 w-3.5" />
                          词汇输入格式示例
                        </button>
                      </HoverCardTrigger>
                      <HoverCardContent align="start" className="w-80">
                        <div className="flex flex-col gap-2">
                          <p className="text-xs font-semibold text-foreground">词汇输入格式示例</p>
                          <p className="text-xs text-muted-foreground">
                            每行一个英文单词或短语；不要包含中文、句子或标点符号。
                          </p>
                          <div className="rounded-lg bg-muted p-3 text-xs leading-[1.8] text-muted-foreground">
                            <p>dolphin</p>
                            <p>sunflower</p>
                            <p>polar bear</p>
                            <p>well-known</p>
                            <p>iPhone</p>
                          </div>
                        </div>
                      </HoverCardContent>
                    </HoverCard>
                    <div className="ml-auto flex items-center gap-2">
                      <button
                        type="button"
                        disabled={!manualText}
                        onClick={async () => {
                          await navigator.clipboard.writeText(manualText);
                          toast.success("已复制到剪贴板");
                        }}
                        className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
                      >
                        <Copy className="h-3.5 w-3.5" />
                        复制
                      </button>
                      <button
                        type="button"
                        disabled={!manualText}
                        onClick={() => {
                          setManualText("");
                          setManualError("");
                          setIsFormatted(false);
                          setIsFromAi(false);
                        }}
                        className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                        清空
                      </button>
                    </div>
                  </div>

                  <textarea
                    value={manualText}
                    onChange={(e) => {
                      setManualText(e.target.value);
                      setManualError("");
                      // Any manual edit invalidates the format/AI gate — the user
                      // must re-format (or re-use AI) before save unlocks again.
                      setIsFormatted(false);
                      setIsFromAi(false);
                    }}
                    placeholder={"fast\nquick\nrun fast"}
                    rows={10}
                    className={`w-full resize-none rounded-xl border bg-muted/50 px-3 py-2.5 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 ${
                      manualError ? "border-red-400 focus:ring-red-400" : "border-border focus:ring-teal-500"
                    }`}
                  />
                  {manualError && (
                    <p className="flex items-center gap-1.5 text-xs text-red-500">
                      <CircleAlert className="h-3.5 w-3.5 shrink-0" />
                      {manualError}
                    </p>
                  )}
                </div>
              ) : (
                <div className="flex flex-col gap-5 px-6 py-4">
                  {/* Difficulty */}
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-1.5">
                      <Gauge className="h-3.5 w-3.5 text-teal-600" />
                      <span className="text-[13px] font-semibold text-foreground">难度</span>
                    </div>
                    <Select value={difficulty} onValueChange={setDifficulty}>
                      <SelectTrigger className="h-11 rounded-xl border-border bg-muted px-4 text-[13px] shadow-none focus:ring-1 focus:ring-teal-500">
                        <SelectValue placeholder="选择难度" />
                      </SelectTrigger>
                      <SelectContent>
                        {DIFFICULTY_OPTIONS.map((opt) => (
                          <SelectItem key={opt.value} value={opt.value}>
                            {opt.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {/* Keywords */}
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-2">
                      <TextCursorInput className="h-3.5 w-3.5 text-teal-600" />
                      <span className="text-[13px] font-semibold text-foreground">关键词</span>
                      <span className="text-xs text-muted-foreground">最多 5 个单词，用空格分开</span>
                    </div>
                    <input
                      value={keywords}
                      onChange={(e) => setKeywords(e.target.value)}
                      placeholder="示例: food fish plate"
                      className={`h-11 rounded-xl border bg-muted px-4 text-[13px] text-foreground outline-none focus:ring-1 ${
                        keywordsWarning ? "border-red-400 focus:ring-red-400" : "border-border focus:ring-teal-500"
                      }`}
                    />
                    {keywordsWarning && (
                      <p className="flex items-center gap-1.5 text-xs text-red-500">
                        <CircleAlert className="h-3.5 w-3.5 shrink-0" />
                        {keywordsWarning}
                      </p>
                    )}
                  </div>

                  {/* Read-only preview (mirrors metadata AI tab pattern). To
                      adjust the list, click 使用 to import into the manual tab. */}
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1.5">
                        <Eye className="h-3.5 w-3.5 text-teal-600" />
                        <span className="text-[13px] font-semibold text-foreground">生成预览</span>
                      </div>
                      {aiWordsText.length > 0 && (
                        <span className="text-xs text-muted-foreground">
                          {aiWordCount} 条
                        </span>
                      )}
                    </div>
                    <div className="min-h-[180px] max-h-[280px] overflow-y-auto rounded-xl border border-border bg-muted p-4">
                      {aiWordsText ? (
                        <p className="whitespace-pre-line font-mono text-[13px] leading-[1.8] text-foreground">
                          {aiWordsText}
                        </p>
                      ) : (
                        <p className="text-xs text-muted-foreground">
                          生成后将在此处显示预览内容...
                        </p>
                      )}
                    </div>
                    {aiError && (
                      <p className="flex items-center gap-1.5 text-xs text-red-500">
                        <CircleAlert className="h-3.5 w-3.5 shrink-0" />
                        {aiError}
                      </p>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Footer */}
            <div className="flex shrink-0 gap-3 px-6 pb-6 pt-3">
              {activeTab === "manual" ? (
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
                    <span className="text-xs font-semibold text-violet-600">检查并格式化</span>
                  </button>
                  <button
                    type="button"
                    disabled={!canSave}
                    onClick={handleSave}
                    className="flex h-11 items-center justify-center gap-1.5 bg-teal-600 px-5 disabled:opacity-50"
                  >
                    {isSaving ? (
                      <Loader2 className="h-4 w-4 animate-spin text-white" />
                    ) : (
                      <Save className="h-4 w-4 text-white" />
                    )}
                    <span className="text-sm font-semibold text-white">保存</span>
                  </button>
                </div>
              ) : (
                <div className="ml-auto flex overflow-hidden rounded-xl border border-border">
                  <button
                    type="button"
                    disabled={!canReset}
                    onClick={handleAiReset}
                    className="flex h-11 items-center justify-center gap-1.5 border-r border-border bg-muted px-4 disabled:opacity-50"
                  >
                    <RotateCcw className="h-3.5 w-3.5 text-muted-foreground" />
                    <span className="text-xs font-semibold text-muted-foreground">重置</span>
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
                      {isGenerating ? "生成中..." : aiWordsText.length > 0 ? "重新生成" : "AI 生成"}
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
