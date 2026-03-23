"use client";

import { useState, useTransition } from "react";
import { Lightbulb, Send, Loader2 } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog, DialogContent, DialogTitle, DialogDescription,
} from "@/components/ui/dialog";
import { submitContentSeekAction } from "@/features/web/hall/actions/content-seek.action";

const MAX_LEN = 30;

type ContentSeekModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** Modal form for submitting a course request */
export function ContentSeekModal({ open, onOpenChange }: ContentSeekModalProps) {
  const [courseName, setCourseName] = useState("");
  const [description, setDescription] = useState("");
  const [diskUrl, setDiskUrl] = useState("");
  const [isPending, startTransition] = useTransition();

  /** Reset form fields */
  function resetForm() {
    setCourseName("");
    setDescription("");
    setDiskUrl("");
  }

  /** Handle form submission */
  function handleSubmit() {
    if (!courseName.trim()) return;

    startTransition(async () => {
      const result = await submitContentSeekAction({
        courseName: courseName.trim(),
        description: description.trim(),
        diskUrl: diskUrl.trim(),
      });

      if ("error" in result) {
        toast.error(result.error);
        return;
      }

      if ("duplicate" in result) {
        toast.info("已有相同申请，正在处理中...");
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
            <Lightbulb className="h-[18px] w-[18px] text-teal-600" />
            求课程
          </DialogTitle>
          <DialogDescription className="sr-only">
            提交您想学习的课程请求
          </DialogDescription>

          <div className="h-px bg-muted" />

          <p className="text-sm leading-[1.5] text-muted-foreground">
            告诉我们想要学习的内容，我们会尽力安排上线！
          </p>

          {/* Course name */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-foreground">
              课程名称
            </label>
            <div className="relative">
              <input
                type="text"
                placeholder="例如：新概念英语第一册"
                value={courseName}
                onChange={(e) => setCourseName(e.target.value)}
                disabled={isPending}
                maxLength={MAX_LEN}
                className="h-11 w-full rounded-[10px] border border-border bg-muted px-4 pr-14 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
              {courseName.length > 0 && (
                <span className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-xs text-muted-foreground">
                  {courseName.length}/{MAX_LEN}
                </span>
              )}
            </div>
          </div>

          {/* Description */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-foreground">
              课程说明
            </label>
            <div className="relative">
              <input
                type="text"
                placeholder="例如：希望增加同步练习"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isPending}
                maxLength={MAX_LEN}
                className="h-11 w-full rounded-[10px] border border-border bg-muted px-4 pr-14 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
              {description.length > 0 && (
                <span className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-xs text-muted-foreground">
                  {description.length}/{MAX_LEN}
                </span>
              )}
            </div>
          </div>

          {/* Disk URL */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-foreground">
              网盘链接
            </label>
            <div className="relative">
              <input
                type="text"
                placeholder="例如：百度网盘/阿里云盘链接"
                value={diskUrl}
                onChange={(e) => setDiskUrl(e.target.value)}
                disabled={isPending}
                maxLength={MAX_LEN}
                className="h-11 w-full rounded-[10px] border border-border bg-muted px-4 pr-14 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-teal-500 focus:ring-1 focus:ring-teal-500 disabled:opacity-50"
              />
              {diskUrl.length > 0 && (
                <span className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-xs text-muted-foreground">
                  {diskUrl.length}/{MAX_LEN}
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
              disabled={isPending || !courseName.trim() || !description.trim() || !diskUrl.trim()}
              className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl bg-teal-600 text-[15px] font-semibold text-white transition-colors hover:bg-teal-700 disabled:opacity-50"
            >
              {isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
              提交请求
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
