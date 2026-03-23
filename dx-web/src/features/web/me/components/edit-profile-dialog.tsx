"use client";

import { useActionState, useEffect } from "react";
import { Loader2 } from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { updateProfileAction } from "@/features/web/me/actions/me.action";
import type { ActionResult } from "@/features/web/me/actions/me.action";

const initialState: ActionResult = {};

interface EditProfileDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  nickname: string | null;
  city: string | null;
  introduction: string | null;
  onProfileChanged?: () => void;
}

/** Dialog for editing profile fields (nickname, city, introduction) */
export function EditProfileDialog({
  open,
  onOpenChange,
  nickname,
  city,
  introduction,
  onProfileChanged,
}: EditProfileDialogProps) {
  const [state, formAction, pending] = useActionState(updateProfileAction, initialState);

  useEffect(() => {
    if (state.success) {
      onOpenChange(false);
      onProfileChanged?.();
    }
  }, [state.success, onOpenChange, onProfileChanged]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>编辑个人资料</DialogTitle>
        </DialogHeader>

        <form action={formAction} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">昵称</label>
            <Input name="nickname" defaultValue={nickname ?? ""} placeholder="设置昵称" maxLength={30} />
            {state.fieldErrors?.nickname && (
              <p className="text-xs text-red-500">{state.fieldErrors.nickname[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">城市</label>
            <Input name="city" defaultValue={city ?? ""} placeholder="所在城市" maxLength={50} />
            {state.fieldErrors?.city && (
              <p className="text-xs text-red-500">{state.fieldErrors.city[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">简介</label>
            <textarea
              name="introduction"
              defaultValue={introduction ?? ""}
              placeholder="介绍一下自己吧"
              maxLength={200}
              rows={3}
              className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
            {state.fieldErrors?.introduction && (
              <p className="text-xs text-red-500">{state.fieldErrors.introduction[0]}</p>
            )}
          </div>

          {state.error && <p className="text-sm text-red-500">{state.error}</p>}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={pending} className="bg-teal-600 hover:bg-teal-700">
              {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
