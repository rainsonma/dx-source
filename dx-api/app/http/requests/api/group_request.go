package api

import (
	"github.com/goravel/framework/contracts/http"
)

// ---------- CreateGroupRequest ----------

type CreateGroupRequest struct {
	Name        string  `form:"name" json:"name"`
	Description *string `form:"description" json:"description"`
}

func (r *CreateGroupRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateGroupRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "required|min_len:2|max_len:50",
		"description": "max_len:200",
	}
}
func (r *CreateGroupRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *CreateGroupRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required": "请输入学习群名称",
		"name.min_len":  "学习群名称至少需要2个字符",
		"name.max_len":  "学习群名称不能超过50个字符",
		"description.max_len": "学习群描述不能超过200个字符",
	}
}

// ---------- UpdateGroupRequest ----------

type UpdateGroupRequest struct {
	Name            string  `form:"name" json:"name"`
	Description     *string `form:"description" json:"description"`
	AnswerTimeLimit *int    `form:"answer_time_limit" json:"answer_time_limit"`
}

func (r *UpdateGroupRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateGroupRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":              "required|min_len:2|max_len:50",
		"description":       "max_len:200",
		"answer_time_limit": "min:5|max:60",
	}
}
func (r *UpdateGroupRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "trim",
		"description": "trim",
	}
}
func (r *UpdateGroupRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required": "请输入学习群名称",
		"name.min_len":  "学习群名称至少需要2个字符",
		"name.max_len":  "学习群名称不能超过50个字符",
		"description.max_len": "学习群描述不能超过200个字符",
	}
}

// ---------- HandleApplicationRequest ----------

type HandleApplicationRequest struct {
	Action string `form:"action" json:"action"`
}

func (r *HandleApplicationRequest) Authorize(ctx http.Context) error { return nil }
func (r *HandleApplicationRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"action": "required|in:accept,reject",
	}
}
func (r *HandleApplicationRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{}
}
func (r *HandleApplicationRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"action.required": "请指定操作类型",
		"action.in":       "操作类型只能为accept或reject",
	}
}
