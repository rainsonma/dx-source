import { z } from "zod";

export const createGroupSchema = z.object({
  name: z.string().min(2, "群名称至少需要2个字符").max(50, "群名称不能超过50个字符"),
  description: z.string().max(200, "群描述不能超过200个字符").optional(),
});

export type CreateGroupInput = z.infer<typeof createGroupSchema>;

export const createSubgroupSchema = z.object({
  name: z.string().min(1, "请输入小组名称").max(50, "小组名称不能超过50个字符"),
});

export type CreateSubgroupInput = z.infer<typeof createSubgroupSchema>;
