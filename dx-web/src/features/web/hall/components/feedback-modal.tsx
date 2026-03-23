"use client";

import { useState, useTransition } from "react";
import { Flag, Send, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog, DialogContent, DialogTitle, DialogDescription,
} from "@/components/ui/dialog";
import { FEEDBACK_TYPES, FEEDBACK_TYPE_LABELS } from "@/consts/feedback-type";
import type { FeedbackType } from "@/consts/feedback-type";
import { submitFeedbackAction } from "@/features/web/hall/actions/feedback.action";

const MAX_LEN = 200;

const TYPE_OPTIONS = Object.values(FEEDBACK_TYPES);

type FeedbackModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** Modal form for submitting feedback */
export function FeedbackModal({ open, onOpenChange }: FeedbackModalProps) {
  const [type, setType] = useState<FeedbackType>(FEEDBACK_TYPES.FEATURE);
  const [description, setDescription] = useState("");
  const [isPending, startTransition] = useTransition();

  /** Reset form fields */
  function resetForm() {
    setType(FEEDBACK_TYPES.FEATURE);
    setDescription("");
  }

  /** Handle form submission */
  function handleSubmit() {
    if (!description.trim()) return;

    startTransition(async () => {
      const result = await submitFeedbackAction({
        type,
        description: description.trim(),
      });

      if ("error" in result) {
        toast.error(result.error);
        return;
      }

      if ("duplicate" in result) {
        toast.info("已有相同建议，正在处理中...");
        return;
      }

      toast.success("提交成功");
      resetForm();
      onOpenChange(false);
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton
        className="max-w-[460px] gap-0 rounded-[20px] border-none p-0"
      >
        <div className="flex flex-col gap-5 p-7">
          {/* Header */}
          <DialogTitle className="flex items-center gap-2.5 text-xl font-bold text-foreground">
            <Flag className="h-[18px] w-[18px] text-teal-600" />
            建议反馈
          </DialogTitle>
          <DialogDescription className="sr-only">
            提交您的建议或反馈
          </DialogDescription>

          <div className="h-px bg-muted" />

          <p className="text-sm leading-[1.5] text-muted-foreground">
            您的每一条建议都能帮助我们做得更好！
          </p>

          {/* Type tags */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-foreground">
              建议类型
            </label>
            <div className="flex flex-wrap gap-1.5">
              {TYPE_OPTIONS.map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setType(t)}
                  disabled={isPending}
                  className={
                    t === type
                      ? "rounded-lg border border-teal-600/30 bg-teal-600/10 px-3 py-1.5 text-[13px] font-semibold text-teal-600"
                      : "rounded-lg border border-border bg-muted px-3 py-1.5 text-[13px] font-medium text-muted-foreground hover:bg-accent"
                  }
                >
                  {FEEDBACK_TYPE_LABELS[t]}
                </button>
              ))}
            </div>
          </div>

          {/* Description */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-foreground">
              详细描述
            </label>
            <div className="relative">
              <textarea
                placeholder="请详细描述你的建议或遇到的问题，我们会认真阅读每一条反馈..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isPending}
                maxLength={MAX_LEN}
                rows={4}
                className="w-full resize-none rounded-[10px] border border-border bg-muted px-4 py-3 text-sm leading-[1.5] text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
              {description.length > 0 && (
                <span className="pointer-events-none absolute right-3 bottom-2 text-xs text-muted-foreground">
                  {description.length}/{MAX_LEN}
                </span>
              )}
            </div>
          </div>

          {/* Buttons */}
          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
              className="flex h-12 flex-1 items-center justify-center rounded-xl border border-border bg-muted text-[15px] font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={isPending || !description.trim()}
              className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl bg-teal-600 text-[15px] font-semibold text-white transition-colors hover:bg-teal-700 disabled:opacity-50"
            >
              {isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
              提交建议
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
