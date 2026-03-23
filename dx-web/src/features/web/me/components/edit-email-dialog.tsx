"use client";

import { useState, useActionState, useEffect } from "react";
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
import {
  sendEmailCodeAction,
  updateEmailAction,
} from "@/features/web/me/actions/me.action";
import type { ActionResult } from "@/features/web/me/actions/me.action";

const initialState: ActionResult = {};

interface EditEmailDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentEmail: string | null;
  onProfileChanged?: () => void;
}

/** Dialog for changing email with verification code */
export function EditEmailDialog({ open, onOpenChange, currentEmail, onProfileChanged }: EditEmailDialogProps) {
  const [email, setEmail] = useState(currentEmail ?? "");
  const [codeSent, setCodeSent] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const [codeState, codeFormAction, codePending] = useActionState(sendEmailCodeAction, initialState);
  const [emailState, emailFormAction, emailPending] = useActionState(updateEmailAction, initialState);

  // Adjust state during render when code action succeeds (React-recommended pattern)
  const [prevCodeState, setPrevCodeState] = useState(codeState);
  if (codeState !== prevCodeState) {
    setPrevCodeState(codeState);
    if (codeState.success) {
      setCodeSent(true);
      setCountdown(60);
    }
  }

  // Adjust state during render when email action succeeds
  const [prevEmailState, setPrevEmailState] = useState(emailState);
  if (emailState !== prevEmailState) {
    setPrevEmailState(emailState);
    if (emailState.success) {
      onOpenChange(false);
      setCodeSent(false);
      setCountdown(0);
      onProfileChanged?.();
    }
  }

  useEffect(() => {
    if (countdown <= 0) return;
    const timer = setInterval(() => setCountdown((c) => c - 1), 1000);
    return () => clearInterval(timer);
  }, [countdown]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>修改邮箱</DialogTitle>
        </DialogHeader>

        {!codeSent ? (
          <form action={codeFormAction} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-foreground">新邮箱</label>
              <Input
                name="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="请输入新邮箱"
              />
              {codeState.fieldErrors?.email && (
                <p className="text-xs text-red-500">{codeState.fieldErrors.email[0]}</p>
              )}
            </div>

            {codeState.error && <p className="text-sm text-red-500">{codeState.error}</p>}

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button type="submit" disabled={codePending} className="bg-teal-600 hover:bg-teal-700">
                {codePending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                获取验证码
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <form action={emailFormAction} className="flex flex-col gap-4">
            <input type="hidden" name="email" value={email} />

            <p className="text-sm text-muted-foreground">
              验证码已发送至 <span className="font-medium text-foreground">{email}</span>
            </p>

            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-foreground">验证码</label>
              <Input name="code" placeholder="请输入6位验证码" maxLength={6} />
              {emailState.fieldErrors?.code && (
                <p className="text-xs text-red-500">{emailState.fieldErrors.code[0]}</p>
              )}
            </div>

            {emailState.error && <p className="text-sm text-red-500">{emailState.error}</p>}

            <div className="flex items-center justify-between">
              <button
                type="button"
                disabled={countdown > 0}
                onClick={() => setCodeSent(false)}
                className="text-sm text-teal-600 hover:text-teal-700 disabled:text-muted-foreground"
              >
                {countdown > 0 ? `${countdown}s 后重新获取` : "重新获取"}
              </button>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button type="submit" disabled={emailPending} className="bg-teal-600 hover:bg-teal-700">
                {emailPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                确认修改
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
