"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";

import { authApi } from "@/lib/api-client";
import { setAccessToken } from "@/lib/token";
import { sendCodeSchema, signUpSchema } from "@/features/web/auth/schemas/signup.schema";

type ActionResult = {
  success?: boolean;
  error?: string;
  fieldErrors?: Record<string, string[]>;
};

const COUNTDOWN_SECONDS = 60;

export function useSignup() {
  const router = useRouter();

  const [codeState, setCodeState] = useState<ActionResult>({});
  const [codePending, setCodePending] = useState(false);

  const [signUpState, setSignUpState] = useState<ActionResult>({});
  const [signUpPending, setSignUpPending] = useState(false);

  const [email, setEmail] = useState("");
  const [countdown, setCountdown] = useState(0);
  const [showPassword, setShowPassword] = useState(false);
  const [agreed, setAgreed] = useState(false);

  // Countdown timer
  useEffect(() => {
    if (countdown <= 0) return;
    const timer = setInterval(() => {
      setCountdown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [countdown]);

  // Redirect on successful signup
  useEffect(() => {
    if (signUpState.success) {
      router.push("/hall");
    }
  }, [signUpState, router]);

  /** Send verification code via Go API */
  const handleSendCode = useCallback(
    async (formData: FormData) => {
      const raw = { email: formData.get("email") };
      const parsed = sendCodeSchema.safeParse(raw);

      if (!parsed.success) {
        setCodeState({ fieldErrors: parsed.error.flatten().fieldErrors });
        return;
      }

      setCodePending(true);
      setCodeState({});

      try {
        const res = await authApi.sendSignUpCode(parsed.data.email);
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

  /** Sign up via Go API */
  const handleSignUp = useCallback(
    async (formData: FormData) => {
      const raw = {
        email: formData.get("email"),
        code: formData.get("code"),
        username: formData.get("username") || "",
        password: formData.get("password") || "",
        agreed: formData.get("agreed") === "on" ? true : undefined,
      };
      const parsed = signUpSchema.safeParse(raw);

      if (!parsed.success) {
        setSignUpState({ fieldErrors: parsed.error.flatten().fieldErrors });
        return;
      }

      setSignUpPending(true);
      setSignUpState({});

      try {
        const res = await authApi.signUp({
          email: parsed.data.email,
          code: parsed.data.code,
          username: parsed.data.username || undefined,
          password: parsed.data.password || undefined,
        });
        if (res.code !== 0) {
          setSignUpState({ error: res.message || "注册失败" });
        } else {
          setAccessToken(res.data.access_token);
          setSignUpState({ success: true });
        }
      } catch {
        setSignUpState({ error: "网络错误，请稍后重试" });
      } finally {
        setSignUpPending(false);
      }
    },
    []
  );

  const togglePassword = useCallback(() => {
    setShowPassword((prev) => !prev);
  }, []);

  const toggleAgreed = useCallback(() => {
    setAgreed((prev) => !prev);
  }, []);

  return {
    codeState,
    handleSendCode,
    codePending,
    countdown,
    canSendCode: countdown === 0 && !codePending,

    signUpState,
    handleSignUp,
    signUpPending,

    email,
    setEmail,
    showPassword,
    togglePassword,
    agreed,
    toggleAgreed,
  };
}
