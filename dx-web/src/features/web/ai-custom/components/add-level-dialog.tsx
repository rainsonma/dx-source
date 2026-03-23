"use client";

import { Layers, Plus, Loader2, X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useCreateGameLevel } from "@/features/web/ai-custom/hooks/use-create-game-level";

type AddLevelDialogProps = {
  gameId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function AddLevelDialog({
  gameId,
  open,
  onOpenChange,
}: AddLevelDialogProps) {
  const { state, formAction, isPending } = useCreateGameLevel(gameId, () =>
    onOpenChange(false)
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        aria-describedby={undefined}
        showCloseButton={false}
        className="max-w-[520px] overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
      >
        <VisuallyHidden>
          <DialogTitle>添加课程游戏关卡</DialogTitle>
        </VisuallyHidden>
        <form action={formAction} className="flex flex-col">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-5">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-[10px] bg-teal-50">
                <Layers className="h-5 w-5 text-teal-600" />
              </div>
              <h2 className="text-lg font-bold text-foreground">
                添加课程游戏关卡
              </h2>
            </div>
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              aria-label="关闭"
              className="flex h-8 w-8 items-center justify-center rounded-lg bg-muted"
            >
              <X className="h-4 w-4 text-muted-foreground" />
            </button>
          </div>

          <div className="h-px w-full bg-border" />

          {/* Form fields */}
          <div className="flex flex-col gap-5 px-7 py-5">
            {state.error && (
              <p className="text-sm text-red-500">{state.error}</p>
            )}

            {/* Title */}
            <div className="flex h-12 items-center gap-2 rounded-[10px] border border-border px-4">
              <span className="w-[65px] shrink-0 text-sm font-medium text-foreground">
                关卡标题
              </span>
              <Input
                name="name"
                placeholder="请输入"
                className="h-full flex-1 border-0 bg-transparent p-0 text-[15px] shadow-none focus-visible:ring-0"
              />
            </div>
            {state.fieldErrors?.name && (
              <p className="-mt-4 text-xs text-red-500">
                {state.fieldErrors.name[0]}
              </p>
            )}

            {/* Description */}
            <div className="flex items-start gap-2 rounded-[10px] border border-border p-4">
              <span className="mt-0.5 w-[65px] shrink-0 text-sm font-medium leading-relaxed text-foreground">
                关卡描述
              </span>
              <Textarea
                name="description"
                placeholder="请输入关卡描述..."
                rows={3}
                className="min-h-[60px] flex-1 resize-none border-0 bg-transparent p-0 text-[15px] leading-relaxed shadow-none [field-sizing:fixed] focus-visible:ring-0"
              />
            </div>
          </div>

          {/* Footer buttons */}
          <div className="flex gap-3 px-7 pb-6">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
              className="flex h-11 flex-1 items-center justify-center rounded-xl border border-border bg-muted disabled:opacity-50"
            >
              <span className="text-sm font-semibold text-muted-foreground">取消</span>
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="flex h-11 flex-1 items-center justify-center gap-2 rounded-xl bg-gradient-to-b from-teal-600 to-teal-700 disabled:opacity-50"
            >
              {isPending ? (
                <Loader2 className="h-4 w-4 animate-spin text-white" />
              ) : (
                <Plus className="h-4 w-4 text-white" />
              )}
              <span className="text-sm font-semibold text-white">添加</span>
            </button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
