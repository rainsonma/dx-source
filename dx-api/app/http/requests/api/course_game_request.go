package api

import (
	"github.com/goravel/framework/contracts/http"

	"dx-api/app/helpers"
)

// ---------- CreateGameRequest ----------

type CreateGameRequest struct {
	Name           string  `form:"name" json:"name"`
	Description    *string `form:"description" json:"description"`
	GameMode       string  `form:"gameMode" json:"gameMode"`
	GameCategoryID string  `form:"gameCategoryId" json:"gameCategoryId"`
	GamePressID    string  `form:"gamePressId" json:"gamePressId"`
	CoverID        *string `form:"coverId" json:"coverId"`
}

func (r *CreateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":           "required|min_len:2|max_len:100",
		"description":    "max_len:500",
		"gameMode":       "required|" + helpers.InEnum("mode"),
		"gameCategoryId": "required|uuid",
		"gamePressId":    "uuid",
		"coverId":        "uuid",
	}
}
func (r *CreateGameRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *CreateGameRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":           "请输入游戏名称",
		"name.min_len":            "游戏名称至少需要2个字符",
		"name.max_len":            "游戏名称不能超过100个字符",
		"description.max_len":     "游戏描述不能超过500个字符",
		"gameMode.required":       "请选择游戏模式",
		"gameMode.in":             "无效的游戏模式",
		"gameCategoryId.required": "请选择游戏分类",
		"gameCategoryId.uuid":     "无效的游戏分类",
		"gamePressId.uuid": "无效的出版社",
		"coverId.uuid":            "无效的封面图片",
	}
}

// ---------- UpdateGameRequest ----------

type UpdateGameRequest struct {
	Name           string  `form:"name" json:"name"`
	Description    *string `form:"description" json:"description"`
	GameMode       string  `form:"gameMode" json:"gameMode"`
	GameCategoryID string  `form:"gameCategoryId" json:"gameCategoryId"`
	GamePressID    string  `form:"gamePressId" json:"gamePressId"`
	CoverID        *string `form:"coverId" json:"coverId"`
}

func (r *UpdateGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":           "required|min_len:2|max_len:100",
		"description":    "max_len:500",
		"gameMode":       "required|" + helpers.InEnum("mode"),
		"gameCategoryId": "required|uuid",
		"gamePressId":    "uuid",
		"coverId":        "uuid",
	}
}
func (r *UpdateGameRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *UpdateGameRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":           "请输入游戏名称",
		"name.min_len":            "游戏名称至少需要2个字符",
		"name.max_len":            "游戏名称不能超过100个字符",
		"description.max_len":     "游戏描述不能超过500个字符",
		"gameMode.required":       "请选择游戏模式",
		"gameMode.in":             "无效的游戏模式",
		"gameCategoryId.required": "请选择游戏分类",
		"gameCategoryId.uuid":     "无效的游戏分类",
		"gamePressId.uuid": "无效的出版社",
		"coverId.uuid":            "无效的封面图片",
	}
}

// ---------- CreateLevelRequest ----------

type CreateLevelRequest struct {
	Name        string  `form:"name" json:"name"`
	Description *string `form:"description" json:"description"`
}

func (r *CreateLevelRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateLevelRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "required|min_len:1|max_len:100",
		"description": "max_len:500",
	}
}
func (r *CreateLevelRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *CreateLevelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":       "请输入关卡名称",
		"name.max_len":        "关卡名称不能超过100个字符",
		"description.max_len": "关卡描述不能超过500个字符",
	}
}

// ---------- SaveMetadataBatchRequest ----------

type SaveMetadataBatchRequest struct {
	GameLevelID string              `json:"gameLevelId"`
	SourceFrom  string              `json:"sourceFrom"`
	Entries     []MetadataEntryJSON `json:"entries"`
}

func (r *SaveMetadataBatchRequest) Authorize(ctx http.Context) error { return nil }
func (r *SaveMetadataBatchRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entries":              "required|min_len:1|max_len:200",
		"gameLevelId":          "required|uuid",
		"sourceFrom":           "required|" + helpers.InEnum("source_from"),
		"entries.*.sourceData": "required",
		"entries.*.sourceType": "required|" + helpers.InEnum("source_type"),
	}
}
func (r *SaveMetadataBatchRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{}
}
func (r *SaveMetadataBatchRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entries.required":              "请提供内容数据",
		"entries.min_len":               "请至少提供一条数据",
		"entries.max_len":               "单次最多提交200条",
		"gameLevelId.required":          "请指定关卡",
		"gameLevelId.uuid":              "无效的关卡ID",
		"sourceFrom.required":           "请指定来源",
		"sourceFrom.in":                 "无效的来源类型",
		"entries.*.sourceData.required": "每条数据的内容不能为空",
		"entries.*.sourceType.required": "每条数据的类型不能为空",
		"entries.*.sourceType.in":       "无效的内容类型",
	}
}

type MetadataEntryJSON struct {
	SourceData  string  `json:"sourceData"`
	Translation *string `json:"translation"`
	SourceType  string  `json:"sourceType"`
}

// ---------- ReorderMetadataRequest ----------

type ReorderMetadataRequest struct {
	GameLevelID string  `json:"gameLevelId"`
	MetaID      string  `json:"metaId"`
	NewOrder    float64 `json:"newOrder"`
}

func (r *ReorderMetadataRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderMetadataRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"metaId":      "required|uuid",
		"gameLevelId": "required|uuid",
		"newOrder":    "required|min:0",
	}
}
func (r *ReorderMetadataRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"metaId.required":      "请指定元数据",
		"metaId.uuid":          "无效的元数据ID",
		"gameLevelId.required": "请指定关卡",
		"gameLevelId.uuid":     "无效的关卡ID",
		"newOrder.required":    "请指定排序位置",
		"newOrder.min":         "排序位置不能为负数",
	}
}

// ---------- InsertContentItemRequest ----------

type InsertContentItemRequest struct {
	GameLevelID     string  `json:"gameLevelId"`
	ContentMetaID   string  `json:"contentMetaId"`
	Content         string  `json:"content"`
	ContentType     string  `json:"contentType"`
	Translation     *string `json:"translation"`
	ReferenceItemID string  `json:"referenceItemId"`
	Direction       string  `json:"direction"`
}

func (r *InsertContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *InsertContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"contentMetaId":   "required|uuid",
		"gameLevelId":     "required|uuid",
		"contentType":     helpers.InEnum("content_type"),
		"direction":       "in:before,after,above,below",
		"referenceItemId": "uuid",
	}
}
func (r *InsertContentItemRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}
func (r *InsertContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"contentMetaId.required": "请指定元数据",
		"contentMetaId.uuid":     "无效的元数据ID",
		"gameLevelId.required":   "请指定关卡",
		"gameLevelId.uuid":       "无效的关卡ID",
		"contentType.in":         "无效的内容类型",
		"direction.in":           "插入方向只能为前或后",
		"referenceItemId.uuid":   "无效的参考项ID",
	}
}

// ---------- UpdateContentItemTextRequest ----------

type UpdateContentItemTextRequest struct {
	Content     string  `json:"content"`
	Translation *string `json:"translation"`
}

func (r *UpdateContentItemTextRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateContentItemTextRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content":     "max_len:2000",
		"translation": "max_len:2000",
	}
}
func (r *UpdateContentItemTextRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content":     "trim",
		"translation": "trim",
	}
}
func (r *UpdateContentItemTextRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.max_len":     "内容不能超过2000个字符",
		"translation.max_len": "翻译不能超过2000个字符",
	}
}

// ---------- ReorderContentItemRequest ----------

type ReorderContentItemRequest struct {
	GameLevelID string  `json:"gameLevelId"`
	ItemID      string  `json:"itemId"`
	NewOrder    float64 `json:"newOrder"`
}

func (r *ReorderContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"gameLevelId": "required|uuid",
		"itemId":      "required|uuid",
		"newOrder":    "required|min:0",
	}
}
func (r *ReorderContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"gameLevelId.required": "请指定关卡",
		"gameLevelId.uuid":     "无效的关卡ID",
		"itemId.required":      "请指定内容项",
		"itemId.uuid":          "无效的内容项ID",
		"newOrder.required":    "请指定排序位置",
		"newOrder.min":         "排序位置不能为负数",
	}
}
