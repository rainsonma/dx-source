"use client";

import { useEffect, useState } from "react";
import { Gamepad2, X, Plus, Loader2 } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ImageUploader } from "@/features/com/images/components/image-uploader";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { GAME_MODE_OPTIONS } from "@/consts/game-mode";
import { IMAGE_ROLES } from "@/consts/image-role";
import { useCreateCourseGame } from "@/features/web/ai-custom/hooks/use-create-course-game";
import { toast } from "sonner";

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean };
type SelectOption = { id: string; name: string };

type CreateCourseFormProps = {
  categories: CategoryOption[];
  presses: SelectOption[];
  onClose: () => void;
};

export function CreateCourseForm({
  categories,
  presses,
  onClose,
}: CreateCourseFormProps) {
  const [pressId, setPressId] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);
  const { state, formAction, isPending, coverUrl, setCoverUrl } =
    useCreateCourseGame(onClose);

  useEffect(() => {
    if (state.error) toast.error(state.error);
    if (state.fieldErrors) {
      const first = Object.values(state.fieldErrors).flat()[0];
      if (first) toast.error(first);
    }
  }, [state]);

  return (
    <form action={formAction} className="flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-5 md:px-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-[10px] bg-teal-100">
            <Gamepad2 className="h-5 w-5 text-teal-600" />
          </div>
          <h2 className="text-lg font-bold text-foreground">
            创建我的课程游戏
          </h2>
        </div>
        <button
          type="button"
          onClick={onClose}
          aria-label="关闭"
          className="flex h-8 w-8 items-center justify-center rounded-lg bg-muted"
        >
          <X className="h-4 w-4 text-muted-foreground" />
        </button>
      </div>

      <div className="h-px w-full bg-border" />

      {/* Form fields */}
      <div className="flex flex-col gap-5 px-4 py-5 md:px-7">
        {/* Category select */}
        <div className="flex h-12 items-center gap-2 rounded-[10px] border border-border px-4">
          <span className="w-[65px] shrink-0 text-sm font-medium text-foreground">
            分类
          </span>
          <Select name="gameCategoryId">
            <SelectTrigger className="h-full flex-1 border-0 bg-transparent p-0 shadow-none focus:ring-0">
              <SelectValue placeholder="请选择" />
            </SelectTrigger>
            <SelectContent>
              {categories.map((c) =>
                c.isLeaf ? (
                  <SelectItem
                    key={c.id}
                    value={c.id}
                    style={c.depth > 0 ? { paddingLeft: c.depth * 24 + 8 } : undefined}
                  >
                    {c.name}
                  </SelectItem>
                ) : (
                  <div
                    key={c.id}
                    className="px-2 py-1.5 text-xs font-semibold text-muted-foreground"
                    style={c.depth > 0 ? { paddingLeft: c.depth * 24 + 8 } : undefined}
                  >
                    {c.name}
                  </div>
                )
              )}
            </SelectContent>
          </Select>
        </div>

        {/* Publisher select */}
        <div className="flex h-12 items-center gap-2 rounded-[10px] border border-border px-4">
          <span className="w-[65px] shrink-0 text-sm font-medium text-foreground">
            出版社
          </span>
          <Select value={pressId} onValueChange={setPressId}>
            <SelectTrigger className="h-full flex-1 border-0 bg-transparent p-0 shadow-none focus:ring-0">
              <SelectValue placeholder="请选择（可选）" />
            </SelectTrigger>
            <SelectContent>
              {presses.map((p) => (
                <SelectItem key={p.id} value={p.id}>
                  {p.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {pressId && (
            <button type="button" onClick={() => setPressId("")} className="shrink-0 text-muted-foreground hover:text-foreground">
              <X className="h-3.5 w-3.5" />
            </button>
          )}
          <input type="hidden" name="gamePressId" value={pressId} />
        </div>

        {/* Game mode select */}
        <div className="flex h-12 items-center gap-2 rounded-[10px] border border-border px-4">
          <span className="w-[65px] shrink-0 text-sm font-medium text-foreground">
            游戏模式
          </span>
          <Select name="gameMode">
            <SelectTrigger className="h-full flex-1 border-0 bg-transparent p-0 shadow-none focus:ring-0">
              <SelectValue placeholder="请选择" />
            </SelectTrigger>
            <SelectContent>
              {GAME_MODE_OPTIONS.map((m) => (
                <SelectItem key={m.value} value={m.value}>
                  {m.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Title input */}
        <div className="flex h-12 items-center gap-2 rounded-[10px] border border-border px-4">
          <span className="w-[65px] shrink-0 text-sm font-medium text-foreground">
            标题
          </span>
          <Input
            name="name"
            placeholder="请输入"
            className="h-full flex-1 border-0 bg-transparent p-0 text-[15px] shadow-none focus-visible:ring-0"
          />
        </div>

        {/* Description textarea */}
        <div className="flex items-start gap-2 rounded-[10px] border border-border px-4 py-3">
          <span className="mt-1 w-[65px] shrink-0 text-sm font-medium text-foreground">
            描述
          </span>
          <Textarea
            name="description"
            placeholder="请输入"
            rows={2}
            className="min-h-[3.5rem] flex-1 resize-none border-0 bg-transparent p-0 text-[15px] leading-relaxed shadow-none [field-sizing:fixed] focus-visible:ring-0"
          />
        </div>

        {/* Cover image (optional) */}
        <div className="flex flex-col gap-1.5">
          <span className="text-sm font-semibold text-foreground">
            封面图
            <span className="ml-1 text-xs font-normal text-muted-foreground">
              (可选)
            </span>
          </span>
          <ImageUploader
            role={IMAGE_ROLES.GAME_COVER}
            onImageChange={setCoverUrl}
          />
          <input type="hidden" name="coverUrl" value={coverUrl ?? ""} />
        </div>

        {/* Private switch */}
        <div className="flex items-center justify-between rounded-[10px] border border-border px-4 py-3">
          <Label htmlFor="isPrivate" className="text-sm font-medium text-foreground">
            私有
            <span className="ml-1 text-xs font-normal text-muted-foreground">
              (仅自己可见)
            </span>
          </Label>
          <Switch
            id="isPrivate"
            checked={isPrivate}
            onCheckedChange={setIsPrivate}
          />
        </div>
        <input type="hidden" name="isPrivate" value={isPrivate ? "true" : "false"} />
      </div>

      {/* Footer buttons */}
      <div className="flex gap-3 px-4 pb-6 md:px-7">
        <button
          type="button"
          onClick={onClose}
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
          <span className="text-sm font-semibold text-white">创建</span>
        </button>
      </div>
    </form>
  );
}
