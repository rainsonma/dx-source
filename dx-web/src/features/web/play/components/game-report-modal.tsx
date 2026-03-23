"use client";

import { useState } from "react";
import { Flag, X, Send, Loader2, ListChecks, MessageSquareText } from "lucide-react";
import { toast } from "sonner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { useGameStore } from "@/features/web/play/hooks/use-game-store";
import { submitReportAction } from "@/features/web/play/actions/report.action";

const reportReasons = [
  "英文错误",
  "翻译准确",
  "音频问题",
  "题目问题",
  "其它问题",
];

export function GameReportModal() {
  const closeOverlay = useGameStore((s) => s.closeOverlay);
  const gameId = useGameStore((s) => s.gameId);
  const levelId = useGameStore((s) => s.levelId);
  const contentItems = useGameStore((s) => s.contentItems);
  const currentIndex = useGameStore((s) => s.currentIndex);

  const [selectedReason, setSelectedReason] = useState("");
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);

  /** Get the current content item ID being played */
  const currentContentItemId = contentItems?.[currentIndex]?.id ?? null;

  /** Submit the report to the server */
  async function handleSubmit() {
    if (!selectedReason || !note.trim() || !gameId || !levelId || !currentContentItemId) return;

    setSubmitting(true);
    const { error } = await submitReportAction({
      gameId,
      gameLevelId: levelId,
      contentItemId: currentContentItemId,
      reason: selectedReason,
      note: note.trim(),
    });
    setSubmitting(false);

    if (error) {
      toast.error(error);
      return;
    }

    toast.success("反馈已提交");
    closeOverlay();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 px-4">
      <div className="flex w-full max-w-[420px] flex-col gap-6 rounded-[20px] bg-card p-6 md:p-8">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <Flag className="h-5 w-5 text-red-500" />
            <h2 className="text-xl font-bold text-foreground">问题反馈</h2>
          </div>
          <button
            type="button"
            aria-label="关闭"
            onClick={closeOverlay}
            disabled={submitting}
            className="flex h-8 w-8 items-center justify-center rounded-lg bg-muted"
          >
            <X className="h-4 w-4 text-muted-foreground" />
          </button>
        </div>

        <div className="h-px w-full bg-border" />

        <div className="flex flex-col gap-2">
          <Label htmlFor="report-reason">
            <ListChecks className="h-4 w-4 text-muted-foreground" />
            问题或建议选择
          </Label>
          <Select
            value={selectedReason}
            onValueChange={setSelectedReason}
            disabled={submitting}
          >
            <SelectTrigger id="report-reason" className="w-full">
              <SelectValue placeholder="请选择问题类型" />
            </SelectTrigger>
            <SelectContent>
              {reportReasons.map((reason) => (
                <SelectItem key={reason} value={reason}>
                  {reason}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="report-note">
            <MessageSquareText className="h-4 w-4 text-muted-foreground" />
            问题或建议描述
          </Label>
          <textarea
            id="report-note"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            disabled={submitting}
            placeholder="请描述您遇到的问题或建议"
            required
            rows={3}
            className="resize-none rounded-xl border-[1.5px] border-border bg-card px-4 py-3 text-[15px] text-foreground placeholder:text-muted-foreground focus:border-ring focus:outline-none"
          />
        </div>

        <button
          type="button"
          onClick={handleSubmit}
          disabled={!selectedReason || !note.trim() || submitting}
          className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-teal-600 disabled:opacity-50"
        >
          {submitting ? (
            <Loader2 className="h-[18px] w-[18px] animate-spin text-white" />
          ) : (
            <Send className="h-[18px] w-[18px] text-white" />
          )}
          <span className="text-[15px] font-semibold text-white">
            {submitting ? "提交中..." : "提交反馈"}
          </span>
        </button>
      </div>
    </div>
  );
}
