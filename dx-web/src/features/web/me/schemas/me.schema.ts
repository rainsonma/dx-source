import { z } from "zod";

/** Profile edit form schema (nickname, city, introduction) */
export const updateProfileSchema = z.object({
  nickname: z.string().max(30, "昵称最长30个字符").optional().or(z.literal("")),
  city: z.string().max(50, "城市最长50个字符").optional().or(z.literal("")),
  introduction: z.string().max(200, "简介最长200个字符").optional().or(z.literal("")),
});

/** Email change - send verification code */
export const sendEmailCodeSchema = z.object({
  email: z.string().min(1, "请输入邮箱").email("邮箱格式不正确"),
});

/** Email change - verify code and update */
export const updateEmailSchema = z.object({
  email: z.string().min(1, "请输入邮箱").email("邮箱格式不正确"),
  code: z.string().length(6, "验证码为6位数字").regex(/^\d{6}$/, "验证码为6位数字"),
});

/** Password change form schema */
export const changePasswordSchema = z
  .object({
    currentPassword: z.string().min(1, "请输入当前密码"),
    newPassword: z
      .string()
      .min(8, "密码至少8个字符")
      .regex(/[a-z]/, "密码需包含小写字母")
      .regex(/[A-Z]/, "密码需包含大写字母")
      .regex(/\d/, "密码需包含数字"),
    confirmPassword: z.string().min(1, "请确认新密码"),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "两次输入的密码不一致",
    path: ["confirmPassword"],
  });

export type UpdateProfileInput = z.infer<typeof updateProfileSchema>;
export type SendEmailCodeInput = z.infer<typeof sendEmailCodeSchema>;
export type UpdateEmailInput = z.infer<typeof updateEmailSchema>;
export type ChangePasswordInput = z.infer<typeof changePasswordSchema>;
