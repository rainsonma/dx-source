import { z } from "zod";

export const sendCodeSchema = z.object({
  email: z
    .string()
    .min(1, "请输入邮箱")
    .email("邮箱格式不正确"),
});

export const signUpSchema = z.object({
  email: z
    .string()
    .min(1, "请输入邮箱")
    .email("邮箱格式不正确"),
  code: z
    .string()
    .length(6, "验证码为6位数字")
    .regex(/^\d{6}$/, "验证码为6位数字"),
  username: z
    .string()
    .max(30, "账号最长30个字符")
    .regex(/^[a-zA-Z0-9_-]*$/, "账号只能包含字母、数字、下划线和连字符")
    .optional()
    .or(z.literal("")),
  password: z
    .string()
    .min(8, "密码至少8个字符")
    .regex(/[a-z]/, "密码需包含小写字母")
    .regex(/[A-Z]/, "密码需包含大写字母")
    .regex(/\d/, "密码需包含数字")
    .optional()
    .or(z.literal("")),
  agreed: z.literal(true, { message: "请阅读并同意相关协议" }),
});

export type SendCodeInput = z.infer<typeof sendCodeSchema>;
export type SignUpInput = z.infer<typeof signUpSchema>;
