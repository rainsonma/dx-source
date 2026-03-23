import { z } from "zod";
import { FEEDBACK_TYPES } from "@/consts/feedback-type";

const feedbackTypeValues = Object.values(FEEDBACK_TYPES) as [string, ...string[]];

/** Schema for submitting feedback */
export const feedbackSchema = z.object({
  type: z.enum(feedbackTypeValues, {
    error: "请选择建议类型",
  }),
  description: z
    .string()
    .trim()
    .min(1, "请输入详细描述")
    .max(200, "最多200个字符"),
});

export type FeedbackInput = z.infer<typeof feedbackSchema>;
