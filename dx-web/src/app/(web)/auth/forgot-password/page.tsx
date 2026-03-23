import { X, EyeOff } from "lucide-react";

export default function ForgotPasswordPage() {
  return (
    <div className="flex min-h-screen w-full items-center justify-center bg-slate-900/50">
      <div className="flex w-[460px] flex-col gap-6 rounded-2xl bg-white p-8 shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-2">
            <h2 className="text-2xl font-bold text-slate-900">忘记密码</h2>
            <p className="text-sm text-slate-400">输入邮箱验证后重置密码</p>
          </div>
          <button
            type="button"
            aria-label="关闭"
            className="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-100"
          >
            <X className="h-4 w-4 text-slate-500" />
          </button>
        </div>

        {/* Form fields */}
        <div className="flex flex-col gap-4">
          {/* Email */}
          <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
            <span className="w-[50px] shrink-0 text-sm font-medium text-slate-900">
              邮箱
            </span>
            <span className="flex-1 text-[15px] text-slate-400">请输入</span>
            <span className="text-sm font-semibold text-teal-600">
              获取验证码
            </span>
          </div>

          {/* Verification code */}
          <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
            <span className="w-[50px] shrink-0 text-sm font-medium text-slate-900">
              验证码
            </span>
            <span className="flex-1 text-[15px] text-slate-400">请输入</span>
          </div>

          {/* New password */}
          <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
            <span className="w-[50px] shrink-0 text-sm font-medium text-slate-900">
              新密码
            </span>
            <span className="flex-1 text-[15px] text-slate-400">请输入</span>
            <EyeOff className="h-[18px] w-[18px] text-slate-400" />
          </div>
        </div>

        {/* Submit button */}
        <button
          type="button"
          className="flex h-12 items-center justify-center rounded-[10px] bg-teal-600"
        >
          <span className="text-base font-semibold text-white">重置密码</span>
        </button>
      </div>
    </div>
  );
}
