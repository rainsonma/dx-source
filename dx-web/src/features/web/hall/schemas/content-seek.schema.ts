import { z } from "zod";

/** Schema for submitting a course request */
export const contentSeekSchema = z.object({
  courseName: z
    .string()
    .trim()
    .min(1, "请输入课程名称")
    .max(30, "最多30个字符"),
  description: z
    .string()
    .trim()
    .min(1, "请输入课程说明")
    .max(30, "最多30个字符"),
  diskUrl: z
    .string()
    .trim()
    .min(1, "请输入网盘链接")
    .max(30, "最多30个字符"),
});

export type ContentSeekInput = z.infer<typeof contentSeekSchema>;
