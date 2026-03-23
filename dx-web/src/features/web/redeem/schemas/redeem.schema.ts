import { z } from "zod";
import { USER_GRADES } from "@/consts/user-grade";

/** Schema for redeeming a code */
export const redeemCodeSchema = z.object({
  code: z
    .string()
    .trim()
    .toUpperCase()
    .length(19, "兑换码格式不正确")
    .regex(/^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/, "兑换码格式不正确"),
});

export type RedeemCodeInput = z.infer<typeof redeemCodeSchema>;

const redeemableGrades = [
  USER_GRADES.MONTH,
  USER_GRADES.SEASON,
  USER_GRADES.YEAR,
  USER_GRADES.LIFETIME,
] as const;

/** Schema for generating codes (admin only) */
export const generateCodesSchema = z.object({
  grade: z.enum(redeemableGrades, {
    error: "请选择生成类型",
  }),
  quantity: z.enum(["10", "50", "100", "500"], {
    error: "请选择生成数量",
  }),
});

export type GenerateCodesInput = z.infer<typeof generateCodesSchema>;
