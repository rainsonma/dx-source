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
  Plus,
  CircleAlert,
  Gauge,
  TextCursorInput,
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
import { toast } from "sonner";
import type { CreateVocabResult } from "@/lib/api-client";
import {
  generateVocabWordsAction,
  createVocabsFromWordsAction,
} from "@/features/web/ai-custom/actions/content-vocab.action";
import { formatVocab } from "@/features/web/ai-custom/helpers/vocab-format-api";
import { InsufficientBeansDialog } from "@/components/in/insufficient-beans-dialog";
import { DIFFICULTY_OPTIONS } from "@/consts/difficulty";

type Tab = "manual" | "ai";

type AddVocabDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdded: () => void;
};

const MAX_KEYWORDS = 10;
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

  // AI tab state
  const [difficulty, setDifficulty] = useState("a1-a2");
  const [keywords, setKeywords] = useState("");
  const [isGenerating, setIsGenerating] = useState(false);
  const [aiError, setAiError] = useState("");
  const [aiWords, setAiWords] = useState<string[]>([]);
  const [checkedWords, setCheckedWords] = useState<Set<number>>(new Set());
  const [inlineInput, setInlineInput] = useState("");
  const [showInlineAdd, setShowInlineAdd] = useState(false);

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
      setDifficulty("a1-a2");
      setKeywords("");
      setIsGenerating(false);
      setAiError("");
      setAiWords([]);
      setCheckedWords(new Set());
      setInlineInput("");
      setShowInlineAdd(false);
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
    const result = await formatVocab(raw);
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
    setAiWords(generated);
    setCheckedWords(new Set(generated.map((_, i) => i)));
    setAiError("");
    toast.success(`已生成 ${generated.length} 个词汇`);
  }

  function handleAiReset() {
    setAiWords([]);
    setCheckedWords(new Set());
    setKeywords("");
    setDifficulty("a1-a2");
    setAiError("");
    setInlineInput("");
    setShowInlineAdd(false);
  }

  function toggleWord(index: number) {
    setCheckedWords((prev) => {
      const next = new Set(prev);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  }

  function removeWord(index: number) {
    setAiWords((prev) => {
      const next = prev.filter((_, i) => i !== index);
      return next;
    });
    setCheckedWords((prev) => {
      // Re-index: remove the removed index and shift down indices above it
      const next = new Set<number>();
      for (const i of prev) {
        if (i < index) next.add(i);
        else if (i > index) next.add(i - 1);
      }
      return next;
    });
  }

  function handleAddInline() {
    const word = inlineInput.trim();
    if (!word) return;
    setAiWords((prev) => [...prev, word]);
    setCheckedWords((prev) => new Set([...prev, aiWords.length]));
    setInlineInput("");
    setShowInlineAdd(false);
  }

  // Shared save handler
  function handleSave() {
    let words: string[];

    if (activeTab === "manual") {
      const lines = manualText
        .split("\n")
        .map((l) => l.trim())
        .filter((l) => l.length > 0);
      // Dedup
      const seen = new Set<string>();
      words = lines.filter((l) => {
        const lc = l.toLowerCase();
        if (seen.has(lc)) return false;
        seen.add(lc);
        return true;
      });

      if (words.length === 0) {
        const msg = "请输入至少一个词汇";
        setManualError(msg);
        toast.error(msg);
        return;
      }
    } else {
      words = aiWords.filter((_, i) => checkedWords.has(i));
      if (words.length === 0) {
        toast.error("请至少勾选一个词汇");
        return;
      }
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
  const canReset = aiWords.length > 0 || keywords.trim().length > 0;
  const canSave = activeTab === "manual"
    ? manualText.trim().length > 0 && !isSaving && !isFormatting
    : checkedWords.size > 0 && !isSaving;
  const canFormat = manualText.trim().length > 0 && !isFormatting && !isSaving;

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
                    每行输入一个英文词汇，保存后 AI 将自动补全音标、释义和例句。
                  </p>
                  <textarea
                    value={manualText}
                    onChange={(e) => { setManualText(e.target.value); setManualError(""); }}
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
                      <span className="text-xs text-muted-foreground">最多 10 个单词，用空格分开</span>
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

                  {/* Preview word list */}
                  {aiWords.length > 0 && (
                    <div className="flex flex-col gap-2">
                      <div className="flex items-center justify-between">
                        <span className="text-[13px] font-semibold text-foreground">
                          生成结果
                          <span className="ml-1 text-xs font-normal text-muted-foreground">
                            （{checkedWords.size}/{aiWords.length} 已选）
                          </span>
                        </span>
                        <button
                          type="button"
                          onClick={() => setShowInlineAdd((v) => !v)}
                          className="flex items-center gap-1 rounded-lg bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground hover:bg-accent"
                        >
                          <Plus className="h-3 w-3" />
                          添加词汇
                        </button>
                      </div>

                      {showInlineAdd && (
                        <div className="flex gap-2">
                          <input
                            value={inlineInput}
                            onChange={(e) => setInlineInput(e.target.value)}
                            onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); handleAddInline(); } }}
                            placeholder="输入词汇后回车添加"
                            className="h-9 flex-1 rounded-lg border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-1 focus:ring-teal-500"
                            autoFocus
                          />
                          <button
                            type="button"
                            onClick={handleAddInline}
                            disabled={!inlineInput.trim()}
                            className="flex h-9 items-center rounded-lg bg-teal-600 px-3 text-sm font-semibold text-white disabled:opacity-50"
                          >
                            添加
                          </button>
                        </div>
                      )}

                      <div className="flex max-h-60 flex-col gap-1.5 overflow-y-auto">
                        {aiWords.map((word, i) => (
                          <div
                            key={i}
                            className="flex items-center gap-3 rounded-lg border border-border bg-background px-3 py-2"
                          >
                            <input
                              type="checkbox"
                              checked={checkedWords.has(i)}
                              onChange={() => toggleWord(i)}
                              className="h-4 w-4 accent-teal-600"
                            />
                            <span className="flex-1 text-sm font-medium text-foreground">{word}</span>
                            <button
                              type="button"
                              onClick={() => removeWord(i)}
                              aria-label="删除"
                              className="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-red-500"
                            >
                              <X className="h-3.5 w-3.5" />
                            </button>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {aiError && (
                    <p className="flex items-center gap-1.5 text-xs text-red-500">
                      <CircleAlert className="h-3.5 w-3.5 shrink-0" />
                      {aiError}
                    </p>
                  )}
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
                      {isGenerating ? "生成中..." : aiWords.length > 0 ? "重新生成" : "AI 生成"}
                    </span>
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
                    <span className="text-sm font-semibold text-white">
                      保存{checkedWords.size > 0 ? ` ${checkedWords.size} 条` : ""}
                    </span>
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
