import { z } from "zod";

export const sendSignInCodeSchema = z.object({
  email: z
    .string()
    .min(1, "请输入邮箱")
    .email("邮箱格式不正确"),
});

export const emailSignInSchema = z.object({
  email: z
    .string()
    .min(1, "请输入邮箱")
    .email("邮箱格式不正确"),
  code: z
    .string()
    .length(6, "验证码为6位数字")
    .regex(/^\d{6}$/, "验证码为6位数字"),
});

export const accountSignInSchema = z.object({
  account: z
    .string()
    .min(1, "请输入账号"),
  password: z
    .string()
    .min(1, "请输入密码"),
});

export type SendSignInCodeInput = z.infer<typeof sendSignInCodeSchema>;
export type EmailSignInInput = z.infer<typeof emailSignInSchema>;
export type AccountSignInInput = z.infer<typeof accountSignInSchema>;
