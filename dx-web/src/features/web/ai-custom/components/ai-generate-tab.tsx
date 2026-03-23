"use client";

import { Gauge, TextCursorInput, Eye, Hash, CircleAlert } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DIFFICULTY_OPTIONS } from "@/consts/difficulty";

type AiGenerateTabProps = {
  difficulty: string;
  onDifficultyChange: (value: string) => void;
  keywords: string;
  onKeywordsChange: (value: string) => void;
  preview: string;
  error: string;
};

const MAX_KEYWORDS = 5;
const MAX_WORD_LENGTH = 30;

export function getKeywordsWarning(keywords: string): string {
  const trimmed = keywords.trim();
  if (!trimmed) return "";
  const words = trimmed.split(/\s+/).filter(Boolean);
  if (words.length === 1 && /[,，、;；/|]/.test(trimmed)) {
    return "请用空格分隔关键词";
  }
  if (words.length > MAX_KEYWORDS) {
    return `最多输入 ${MAX_KEYWORDS} 个关键词，当前 ${words.length} 个`;
  }
  const long = words.find((w) => w.length > MAX_WORD_LENGTH);
  if (long) {
    return `单个关键词不能超过 ${MAX_WORD_LENGTH} 个字符`;
  }
  return "";
}

export function AiGenerateTab({
  difficulty,
  onDifficultyChange,
  keywords,
  onKeywordsChange,
  preview,
  error,
}: AiGenerateTabProps) {
  const keywordsWarning = getKeywordsWarning(keywords);

  return (
    <div className="flex flex-col gap-5 px-6 py-3">
      {/* Difficulty */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-1.5">
          <Gauge className="h-3.5 w-3.5 text-teal-600" />
          <span className="text-[13px] font-semibold text-foreground">难度</span>
        </div>
        <Select value={difficulty} onValueChange={onDifficultyChange}>
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
          <span className="text-[13px] font-semibold text-foreground">
            关键词
          </span>
          <span className="text-xs text-muted-foreground">
            最多 5 个单词，用空格分开
          </span>
        </div>
        <input
          value={keywords}
          onChange={(e) => onKeywordsChange(e.target.value)}
          placeholder="示例: food fish plate"
          className={`h-11 rounded-xl border bg-muted px-4 text-[13px] text-foreground outline-none focus:ring-1 ${keywordsWarning ? "border-red-400 focus:ring-red-400" : "border-border focus:ring-teal-500"}`}
        />
        {keywordsWarning && (
          <p className="flex items-center gap-1.5 text-xs text-red-500">
            <CircleAlert className="h-3.5 w-3.5 shrink-0" />
            {keywordsWarning}
          </p>
        )}
      </div>

      {/* Preview */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-1.5">
          <Eye className="h-3.5 w-3.5 text-teal-600" />
          <span className="text-[13px] font-semibold text-foreground">生成预览</span>
        </div>
        <div className="min-h-[180px] max-h-[280px] overflow-y-auto rounded-xl border border-border bg-muted p-4">
          {preview ? (
            <p className="whitespace-pre-line text-[13px] leading-[1.8] text-foreground">
              {preview}
            </p>
          ) : (
            <p className="text-xs text-muted-foreground">
              生成后将在此处显示预览内容...
            </p>
          )}
        </div>
        {error && (
          <p className="flex items-center gap-1.5 text-xs text-red-500">
            <CircleAlert className="h-3.5 w-3.5 shrink-0" />
            {error}
          </p>
        )}
        <div className="flex items-center gap-1.5">
          <Hash className="h-3 w-3 text-muted-foreground" />
          <span className="text-xs text-muted-foreground">
            句数：{preview ? preview.split("\n").filter((l) => l.trim()).length : 0} &nbsp;
            词数：{preview ? preview.split(/\s+/).filter(Boolean).length : 0}
          </span>
        </div>
      </div>
    </div>
  );
}
