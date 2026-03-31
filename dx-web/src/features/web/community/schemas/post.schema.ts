import { z } from "zod";

export const createPostSchema = z.object({
  content: z
    .string()
    .min(1, "请输入内容")
    .max(2000, "内容不能超过2000个字符"),
  image_id: z.string().optional(),
  tags: z
    .array(z.string().max(20, "标签不能超过20个字符"))
    .max(5, "标签不能超过5个")
    .optional(),
});

export type CreatePostInput = z.infer<typeof createPostSchema>;

export const createCommentSchema = z.object({
  content: z
    .string()
    .min(1, "请输入评论内容")
    .max(500, "评论内容不能超过500个字符"),
  parent_id: z.string().optional(),
});

export type CreateCommentInput = z.infer<typeof createCommentSchema>;
