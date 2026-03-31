"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { authApi } from "@/lib/api-client";
import { setAccessToken } from "@/lib/token";
import {
  sendSignInCodeSchema,
  emailSignInSchema,
  accountSignInSchema,
} from "@/features/web/auth/schemas/signin.schema";

type Tab = "email" | "account";

type ActionResult = {
  success?: boolean;
  error?: string;
  fieldErrors?: Record<string, string[]>;
};

const COUNTDOWN_SECONDS = 60;

export function useSignIn() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirectTo = searchParams.get("redirect") || "/hall";

  const [activeTab, setActiveTab] = useState<Tab>("account");

  // Email login state
  const [codeState, setCodeState] = useState<ActionResult>({});
  const [codePending, setCodePending] = useState(false);
  const [emailState, setEmailState] = useState<ActionResult>({});
  const [emailPending, setEmailPending] = useState(false);

  // Account login state
  const [accountState, setAccountState] = useState<ActionResult>({});
  const [accountPending, setAccountPending] = useState(false);

  // UI state
  const [email, setEmail] = useState("");
  const [countdown, setCountdown] = useState(0);
  const [showPassword, setShowPassword] = useState(false);

  // Countdown timer
  useEffect(() => {
    if (countdown <= 0) return;
    const timer = setInterval(() => {
      setCountdown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [countdown]);

  // Redirect on successful login (either method)
  useEffect(() => {
    if (emailState.success) {
      router.push(redirectTo);
    }
  }, [emailState, router, redirectTo]);

  useEffect(() => {
    if (accountState.success) {
      router.push(redirectTo);
    }
  }, [accountState, router, redirectTo]);

  /** Send verification code via Go API */
  const handleSendCode = useCallback(
    async (formData: FormData) => {
      const raw = { email: formData.get("email") };
      const parsed = sendSignInCodeSchema.safeParse(raw);

      if (!parsed.success) {
        setCodeState({ fieldErrors: parsed.error.flatten().fieldErrors });
        return;
      }

      setCodePending(true);
      setCodeState({});

      try {
        const res = await authApi.sendSignInCode(parsed.data.email);
        if (res.code !== 0) {
          setCodeState({ error: res.message || "发送验证码失败" });
        } else {
          setCodeState({ success: true });
          setCountdown(COUNTDOWN_SECONDS);
        }
      } catch {
        setCodeState({ error: "网络错误，请稍后重试" });
      } finally {
        setCodePending(false);
      }
    },
    []
  );

  /** Sign in with email + code via Go API */
  const handleEmailSignIn = useCallback(
    async (formData: FormData) => {
      const raw = {
        email: formData.get("email"),
        code: formData.get("code"),
      };
      const parsed = emailSignInSchema.safeParse(raw);

      if (!parsed.success) {
        setEmailState({ fieldErrors: parsed.error.flatten().fieldErrors });
        return;
      }

      setEmailPending(true);
      setEmailState({});

      try {
        const res = await authApi.signIn({
          email: parsed.data.email,
          code: parsed.data.code,
        });
        if (res.code !== 0) {
          setEmailState({ error: res.message || "登录失败" });
        } else {
          setAccessToken(res.data.access_token);
          setEmailState({ success: true });
        }
      } catch {
        setEmailState({ error: "网络错误，请稍后重试" });
      } finally {
        setEmailPending(false);
      }
    },
    []
  );

  /** Sign in with account + password via Go API */
  const handleAccountSignIn = useCallback(
    async (formData: FormData) => {
      const raw = {
        account: formData.get("account"),
        password: formData.get("password"),
      };
      const parsed = accountSignInSchema.safeParse(raw);

      if (!parsed.success) {
        setAccountState({ fieldErrors: parsed.error.flatten().fieldErrors });
        return;
      }

      setAccountPending(true);
      setAccountState({});

      try {
        const res = await authApi.signIn({
          account: parsed.data.account,
          password: parsed.data.password,
        });
        if (res.code !== 0) {
          setAccountState({ error: res.message || "登录失败" });
        } else {
          setAccessToken(res.data.access_token);
          setAccountState({ success: true });
        }
      } catch {
        setAccountState({ error: "网络错误，请稍后重试" });
      } finally {
        setAccountPending(false);
      }
    },
    []
  );

  const togglePassword = useCallback(() => {
    setShowPassword((prev) => !prev);
  }, []);

  return {
    activeTab,
    setActiveTab,

    // Email login
    codeState,
    handleSendCode,
    codePending,
    emailState,
    handleEmailSignIn,
    emailPending,
    email,
    setEmail,
    countdown,
    canSendCode: countdown === 0 && !codePending,

    // Account login
    accountState,
    handleAccountSignIn,
    accountPending,

    // Shared
    showPassword,
    togglePassword,
  };
}
