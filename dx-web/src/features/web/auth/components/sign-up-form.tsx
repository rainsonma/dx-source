"use client";

import Link from "next/link";
import { Eye, EyeOff, MessageCircle, Loader2, CircleCheck, CircleAlert } from "lucide-react";

import { WechatQrCard } from "@/features/web/auth/components/wechat-qr-card";
import { useSignup } from "@/features/web/auth/hooks/use-signup";

export function SignUpForm() {
  const {
    codeState,
    handleSendCode,
    codePending,
    countdown,
    canSendCode,
    signUpState,
    handleSignUp,
    signUpPending,
    email,
    setEmail,
    showPassword,
    togglePassword,
    agreed,
    toggleAgreed,
  } = useSignup();

  const codeError =
    codeState.error || codeState.fieldErrors?.email?.[0];
  const signUpError =
    signUpState.error ||
    signUpState.fieldErrors?.code?.[0] ||
    signUpState.fieldErrors?.username?.[0] ||
    signUpState.fieldErrors?.password?.[0] ||
    signUpState.fieldErrors?.agreed?.[0];

  return (
    <div className="flex w-[700px] flex-col items-center gap-6">
      {/* Header */}
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-[32px] font-extrabold text-slate-900">
          创建斗学专属账号
        </h1>
        <p className="text-sm text-slate-400">
          进入斗学英语游戏世界，开启英语学习冒险之旅
        </p>
      </div>

      {/* Card */}
      <div className="flex w-full gap-8 rounded-2xl border border-slate-200 bg-white p-8 shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
        {/* Left: WeChat QR */}
        <div className="flex items-center justify-center">
          <WechatQrCard label="微信扫码快速注册" />
        </div>

        {/* Divider */}
        <div className="w-px bg-slate-200" />

        {/* Right: Form Fields */}
        <div className="flex flex-1 flex-col gap-5">
          {/* Email + Send Code (separate form) */}
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleSendCode(new FormData(e.currentTarget));
            }}
          >
            <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
              <span className="w-[50px] text-sm font-medium text-slate-900">
                邮箱
              </span>
              <input
                name="email"
                type="email"
                placeholder="请输入"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
              />
              <button
                type="submit"
                disabled={!canSendCode}
                className="text-sm font-semibold text-teal-600 hover:text-teal-700 disabled:text-slate-400"
              >
                {codePending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : countdown > 0 ? (
                  `${countdown}s`
                ) : (
                  "获取验证码"
                )}
              </button>
            </div>
            {codeError && (
              <p className="mt-1 flex items-center gap-1.5 text-xs text-red-500">
                <CircleAlert className="h-3.5 w-3.5 flex-shrink-0" />
                {codeError}
              </p>
            )}
          </form>

          {/* Signup Form */}
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleSignUp(new FormData(e.currentTarget));
            }}
            className="flex flex-col gap-5"
          >
            <input type="hidden" name="email" value={email} />

            {/* Code */}
            <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
              <span className="w-[50px] text-sm font-medium text-slate-900">
                验证码
              </span>
              <input
                name="code"
                type="text"
                inputMode="numeric"
                maxLength={6}
                placeholder="请输入"
                className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
              />
            </div>

            {/* Code sent hint */}
            {codeState.success && (
              <p className="flex items-center gap-1.5 text-xs font-medium text-slate-700">
                <CircleCheck className="h-4 w-4 text-teal-600" />
                邮箱验证码已发送，可能会有延迟，请耐心等待
              </p>
            )}

            {/* Name */}
            <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
              <span className="w-[50px] text-sm font-medium text-slate-900">
                账号
              </span>
              <input
                name="username"
                type="text"
                placeholder="字母、数字、下划线或连字符，最长30位"
                className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
              />
            </div>

            {/* Password */}
            <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
              <span className="w-[50px] text-sm font-medium text-slate-900">
                密码
              </span>
              <input
                name="password"
                type={showPassword ? "text" : "password"}
                placeholder="至少8位，含大小写字母和数字"
                className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
              />
              <button type="button" onClick={togglePassword}>
                {showPassword ? (
                  <Eye className="h-[18px] w-[18px] cursor-pointer text-slate-400" />
                ) : (
                  <EyeOff className="h-[18px] w-[18px] cursor-pointer text-slate-400" />
                )}
              </button>
            </div>

            {/* Agreement */}
            <label className="flex cursor-pointer items-center gap-2">
              <input
                name="agreed"
                type="checkbox"
                checked={agreed}
                onChange={toggleAgreed}
                className="sr-only"
              />
              <div
                className={`h-4 w-4 flex-shrink-0 rounded border ${
                  agreed
                    ? "border-teal-600 bg-teal-600"
                    : "border-slate-300"
                }`}
              >
                {agreed && (
                  <svg viewBox="0 0 16 16" className="h-4 w-4 text-white">
                    <path
                      fill="currentColor"
                      d="M6.5 11.5L3 8l1-1 2.5 2.5L11 5l1 1z"
                    />
                  </svg>
                )}
              </div>
              <span className="text-xs text-slate-700">
                我已阅读并同意{" "}
                <span className="text-teal-600">用户协议</span>、
                <span className="text-teal-600">隐私政策</span>、
                <span className="text-teal-600">监护人同意书</span>、
                <span className="text-teal-600">产品服务协议</span>
              </span>
            </label>

            {/* Error */}
            {signUpError && (
              <p className="flex items-center gap-1.5 text-xs text-red-500">
                <CircleAlert className="h-3.5 w-3.5 flex-shrink-0" />
                {signUpError}
              </p>
            )}

            {/* Submit */}
            <button
              type="submit"
              disabled={signUpPending || !agreed}
              className="flex h-12 w-full items-center justify-center rounded-[10px] bg-teal-600 text-base font-semibold text-white hover:bg-teal-700 disabled:opacity-50"
            >
              {signUpPending ? (
                <Loader2 className="h-5 w-5 animate-spin" />
              ) : (
                "注册"
              )}
            </button>
          </form>

          {/* Separator */}
          <div className="h-px bg-slate-200" />

          {/* Footer */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1 text-sm">
              <span className="text-slate-400">已有账号？</span>
              <Link
                href="/auth/signin"
                className="font-medium text-teal-600 hover:text-teal-700"
              >
                立即登录
              </Link>
            </div>
            <button className="flex h-8 w-8 items-center justify-center rounded-full bg-teal-600 hover:bg-teal-700">
              <MessageCircle className="h-4 w-4 text-white" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
