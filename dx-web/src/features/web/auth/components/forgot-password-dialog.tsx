"use client";

import { X, EyeOff } from "lucide-react";

interface ForgotPasswordDialogProps {
  open: boolean;
  onClose: () => void;
}

export function ForgotPasswordDialog({
  open,
  onClose,
}: ForgotPasswordDialogProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50">
      <div className="flex w-[460px] flex-col gap-6 rounded-2xl bg-white p-8 shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div className="flex flex-col gap-2">
            <h2 className="text-2xl font-bold text-slate-900">忘记密码</h2>
            <p className="text-sm text-slate-400">输入邮箱验证后重置密码</p>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-lg"
          >
            <X className="h-[18px] w-[18px] text-slate-500" />
          </button>
        </div>

        {/* Fields */}
        <div className="flex flex-col gap-4">
          <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
            <span className="w-[50px] text-sm font-medium text-slate-900">
              邮箱
            </span>
            <input
              type="email"
              placeholder="请输入"
              className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
            />
            <button className="text-sm font-semibold text-teal-600 hover:text-teal-700">
              获取验证码
            </button>
          </div>
          <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
            <span className="w-[50px] text-sm font-medium text-slate-900">
              验证码
            </span>
            <input
              type="text"
              placeholder="请输入"
              className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
            />
          </div>
          <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
            <span className="w-[50px] text-sm font-medium text-slate-900">
              新密码
            </span>
            <input
              type="password"
              placeholder="请输入"
              className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
            />
            <EyeOff className="h-[18px] w-[18px] cursor-pointer text-slate-400" />
          </div>
        </div>

        {/* Submit */}
        <button className="flex h-12 w-full items-center justify-center rounded-[10px] bg-teal-600 text-base font-semibold text-white hover:bg-teal-700">
          重置密码
        </button>
      </div>
    </div>
  );
}
