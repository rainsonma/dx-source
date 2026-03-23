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
import { changePasswordAction } from "@/features/web/me/actions/me.action";
import type { ActionResult } from "@/features/web/me/actions/me.action";

const initialState: ActionResult = {};

interface ChangePasswordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/** Dialog for changing password (current + new + confirm) */
export function ChangePasswordDialog({ open, onOpenChange }: ChangePasswordDialogProps) {
  const [state, formAction, pending] = useActionState(changePasswordAction, initialState);

  useEffect(() => {
    if (state.success) {
      onOpenChange(false);
    }
  }, [state.success, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>修改密码</DialogTitle>
        </DialogHeader>

        <form action={formAction} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">当前密码</label>
            <Input name="currentPassword" type="password" placeholder="请输入当前密码" />
            {state.fieldErrors?.currentPassword && (
              <p className="text-xs text-red-500">{state.fieldErrors.currentPassword[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">新密码</label>
            <Input name="newPassword" type="password" placeholder="至少8位，含大小写字母和数字" />
            {state.fieldErrors?.newPassword && (
              <p className="text-xs text-red-500">{state.fieldErrors.newPassword[0]}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">确认新密码</label>
            <Input name="confirmPassword" type="password" placeholder="再次输入新密码" />
            {state.fieldErrors?.confirmPassword && (
              <p className="text-xs text-red-500">{state.fieldErrors.confirmPassword[0]}</p>
            )}
          </div>

          {state.error && <p className="text-sm text-red-500">{state.error}</p>}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={pending} className="bg-teal-600 hover:bg-teal-700">
              {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认修改
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
