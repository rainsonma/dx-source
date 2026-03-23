"use client";

import Link from "next/link";
import { Eye, EyeOff, MessageCircle, Loader2, CircleCheck, CircleAlert } from "lucide-react";

import { WechatQrCard } from "@/features/web/auth/components/wechat-qr-card";
import { useSignIn } from "@/features/web/auth/hooks/use-signin";

export function SignInForm() {
  const {
    activeTab,
    setActiveTab,
    codeState,
    handleSendCode,
    codePending,
    emailState,
    handleEmailSignIn,
    emailPending,
    email,
    setEmail,
    countdown,
    canSendCode,
    accountState,
    handleAccountSignIn,
    accountPending,
    showPassword,
    togglePassword,
  } = useSignIn();

  const emailCodeError =
    codeState.error || codeState.fieldErrors?.email?.[0];
  const emailLoginError =
    emailState.error ||
    emailState.fieldErrors?.email?.[0] ||
    emailState.fieldErrors?.code?.[0];
  const accountError =
    accountState.error ||
    accountState.fieldErrors?.account?.[0] ||
    accountState.fieldErrors?.password?.[0];

  return (
    <div className="flex w-[700px] flex-col items-center gap-6">
      {/* Header */}
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-[32px] font-extrabold text-slate-900">
          欢迎回到斗学
        </h1>
        <p className="text-sm text-slate-400">
          登录斗学专属账号，继续欢乐学习之旅
        </p>
      </div>

      {/* Card */}
      <div className="flex w-full gap-8 rounded-2xl border border-slate-200 bg-white p-8 shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
        {/* Left: WeChat QR */}
        <div className="flex items-center justify-center">
          <WechatQrCard label="微信扫码快速登录" />
        </div>

        {/* Divider */}
        <div className="w-px bg-slate-200" />

        {/* Right: Form Fields */}
        <div className="flex flex-1 flex-col gap-5">
          {/* Tab Row */}
          <div className="flex items-center gap-6">
            <button
              type="button"
              onClick={() => setActiveTab("account")}
              className={
                activeTab === "account"
                  ? "font-semibold text-teal-600"
                  : "font-medium text-slate-400 hover:text-slate-500"
              }
            >
              账密登录
            </button>
            <button
              type="button"
              onClick={() => setActiveTab("email")}
              className={
                activeTab === "email"
                  ? "font-semibold text-teal-600"
                  : "font-medium text-slate-400 hover:text-slate-500"
              }
            >
              邮箱登录
            </button>
          </div>

          {/* Email Login Tab */}
          {activeTab === "email" && (
            <>
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
                {emailCodeError && (
                  <p className="mt-1 flex items-center gap-1.5 text-xs text-red-500">
                    <CircleAlert className="h-3.5 w-3.5 flex-shrink-0" />
                    {emailCodeError}
                  </p>
                )}
              </form>

              {/* Code sent hint */}
              {codeState.success && (
                <p className="flex items-center gap-1.5 text-xs font-medium text-slate-700">
                  <CircleCheck className="h-4 w-4 text-teal-600" />
                  验证码已发送，可能会有延迟，请耐心等待
                </p>
              )}

              {/* Login form */}
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  handleEmailSignIn(new FormData(e.currentTarget));
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
                    placeholder="请输入6位验证码"
                    className="flex-1 text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
                  />
                </div>

                {/* Error */}
                {emailLoginError && (
                  <p className="flex items-center gap-1.5 text-xs text-red-500">
                    <CircleAlert className="h-3.5 w-3.5 flex-shrink-0" />
                    {emailLoginError}
                  </p>
                )}

                {/* Submit */}
                <button
                  type="submit"
                  disabled={emailPending}
                  className="flex h-12 w-full items-center justify-center rounded-[10px] bg-teal-600 text-base font-semibold text-white hover:bg-teal-700 disabled:opacity-50"
                >
                  {emailPending ? (
                    <Loader2 className="h-5 w-5 animate-spin" />
                  ) : (
                    "登录"
                  )}
                </button>
              </form>
            </>
          )}

          {/* Account Login Tab */}
          {activeTab === "account" && (
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleAccountSignIn(new FormData(e.currentTarget));
              }}
              className="flex flex-col gap-5"
            >
              {/* Account */}
              <div className="flex h-12 items-center gap-2 rounded-[10px] border border-slate-200 px-4">
                <span className="w-[50px] text-sm font-medium text-slate-900">
                  账号
                </span>
                <input
                  name="account"
                  type="text"
                  placeholder="用户名/邮箱/手机号"
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
                  placeholder="请输入"
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

              {/* Error */}
              {accountError && (
                <p className="flex items-center gap-1.5 text-xs text-red-500">
                  <CircleAlert className="h-3.5 w-3.5 flex-shrink-0" />
                  {accountError}
                </p>
              )}

              {/* Submit */}
              <button
                type="submit"
                disabled={accountPending}
                className="flex h-12 w-full items-center justify-center rounded-[10px] bg-teal-600 text-base font-semibold text-white hover:bg-teal-700 disabled:opacity-50"
              >
                {accountPending ? (
                  <Loader2 className="h-5 w-5 animate-spin" />
                ) : (
                  "登录"
                )}
              </button>

              <div className="flex justify-end">
                <span className="cursor-pointer text-[13px] font-medium text-teal-600 hover:text-teal-700">
                  忘记密码？
                </span>
              </div>
            </form>
          )}

          {/* Separator */}
          <div className="h-px bg-slate-200" />

          {/* Footer */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1 text-sm">
              <span className="text-slate-400">没有账号？</span>
              <Link
                href="/auth/signup"
                className="font-medium text-teal-600 hover:text-teal-700"
              >
                立即注册
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
