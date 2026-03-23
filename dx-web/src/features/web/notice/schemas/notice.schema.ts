import { z } from "zod";

/** Validation schema for creating a new notice */
export const createNoticeSchema = z.object({
  title: z
    .string()
    .min(1, "标题不能为空")
    .max(200, "标题不能超过 200 个字符"),
  content: z
    .string()
    .max(2000, "内容不能超过 2000 个字符")
    .optional()
    .transform((v) => v || undefined),
  icon: z
    .string()
    .max(50, "图标名称不能超过 50 个字符")
    .optional()
    .transform((v) => v || undefined),
});

export type CreateNoticeInput = z.infer<typeof createNoticeSchema>;

/** Validation schema for updating an existing notice */
export const updateNoticeSchema = createNoticeSchema.extend({
  id: z.string().uuid("无效的通知 ID"),
});

export type UpdateNoticeInput = z.infer<typeof updateNoticeSchema>;
