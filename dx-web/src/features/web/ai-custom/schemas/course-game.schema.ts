import { z } from "zod";
import { GAME_MODES, type GameMode } from "@/consts/game-mode";
import { SOURCE_FROMS, type SourceFrom } from "@/consts/source-from";
import { SOURCE_TYPES, type SourceType } from "@/consts/source-type";
import { MAX_ENTRIES, MAX_CONTENT_LENGTH } from "@/features/web/ai-custom/helpers/format-metadata";

const gameModeValues = Object.values(GAME_MODES) as [GameMode, ...GameMode[]];

export const createCourseGameSchema = z.object({
  gameCategoryId: z
    .string()
    .min(1, "请选择分类")
    .uuid("无效的分类"),
  gamePressId: z
    .string()
    .min(1, "请选择出版社")
    .uuid("无效的出版社"),
  gameMode: z.enum(gameModeValues, { message: "请选择游戏模式" }),
  name: z
    .string()
    .min(1, "请输入标题")
    .max(100, "标题最长 100 个字符"),
  description: z
    .string()
    .max(500, "描述最长 500 个字符")
    .optional()
    .or(z.literal("")),
  coverUrl: z
    .string()
    .optional()
    .or(z.literal("")),
});

export type CreateCourseGameInput = z.infer<typeof createCourseGameSchema>;

export const createGameLevelSchema = z.object({
  name: z
    .string()
    .min(1, "请输入关卡标题")
    .max(100, "标题最长 100 个字符"),
  description: z
    .string()
    .max(500, "描述最长 500 个字符")
    .optional()
    .or(z.literal("")),
});

export type CreateGameLevelInput = z.infer<typeof createGameLevelSchema>;

const metadataEntrySchema = z.object({
  content: z.string().min(1).max(300, "单条内容最长 300 字符"),
  translation: z.string().max(300, "翻译最长 300 字符").optional(),
  sourceType: z.enum(
    Object.values(SOURCE_TYPES) as [SourceType, ...SourceType[]]
  ),
});

export const createMetadataSchema = z.object({
  gameLevelId: z.string().uuid("无效的关卡"),
  entries: z
    .array(metadataEntrySchema)
    .min(1, "请至少添加一条元数据")
    .max(200, "每次最多提交 200 条"),
  sourceFrom: z.enum(
    Object.values(SOURCE_FROMS) as [SourceFrom, ...SourceFrom[]]
  ),
});

export type CreateMetadataInput = z.infer<typeof createMetadataSchema>;

export const formatMetadataSchema = z.object({
  content: z
    .string()
    .min(1, "内容不能为空")
    .max(MAX_CONTENT_LENGTH * MAX_ENTRIES, "内容过长"),
  formatType: z.enum(["sentence", "vocab"], {
    message: "无效的格式类型",
  }),
});

export type FormatMetadataInput = z.infer<typeof formatMetadataSchema>;

export const generateMetadataSchema = z.object({
  difficulty: z.enum(["a1-a2", "b1-b2", "c1-c2"], {
    message: "请选择有效的难度等级",
  }),
  keywords: z
    .array(z.string().min(1).max(50))
    .min(1, "请至少输入一个关键词")
    .max(5, "关键词最多 5 个"),
});

export type GenerateMetadataInput = z.infer<typeof generateMetadataSchema>;

export const reorderMetaSchema = z.object({
  metaId: z.string().uuid("无效的元数据 ID"),
  newOrder: z.number().positive("排序值必须为正数").max(999999999, "排序值超出范围"),
});

export type ReorderMetaInput = z.infer<typeof reorderMetaSchema>;

export const breakMetadataSchema = z.object({
  gameLevelId: z.string().uuid("无效的关卡 ID"),
});

export type BreakMetadataInput = z.infer<typeof breakMetadataSchema>;

export const generateContentItemsSchema = z.object({
  gameLevelId: z.string().uuid("无效的关卡 ID"),
});

export type GenerateContentItemsInput = z.infer<typeof generateContentItemsSchema>;

export const reorderItemSchema = z.object({
  itemId: z.string().uuid("无效的练习单元 ID"),
  newOrder: z.number().positive("排序值必须为正数").max(999999999, "排序值超出范围"),
});

export type ReorderItemInput = z.infer<typeof reorderItemSchema>;

export const updateContentItemTextSchema = z.object({
  itemId: z.string().uuid("无效的练习单元 ID"),
  content: z.string().min(1, "内容不能为空").max(300, "内容最长 300 字符"),
  translation: z.string().max(300, "翻译最长 300 字符").nullable(),
});

export type UpdateContentItemTextInput = z.infer<typeof updateContentItemTextSchema>;

export const insertContentItemSchema = z.object({
  gameLevelId: z.string().uuid("无效的关卡 ID"),
  contentMetaId: z.string().uuid("无效的元数据 ID"),
  content: z.string().max(300, "内容最长 300 字符"),
  contentType: z.string().min(1, "类型不能为空").max(20, "类型最长 20 字符"),
  translation: z.string().max(300, "翻译最长 300 字符").nullable(),
  referenceItemId: z.string().uuid("无效的参考项 ID"),
  direction: z.enum(["above", "below"], { message: "方向无效" }),
});

export type InsertContentItemInput = z.infer<typeof insertContentItemSchema>;

export const deleteContentItemSchema = z.object({
  itemId: z.string().uuid("无效的练习单元 ID"),
});
