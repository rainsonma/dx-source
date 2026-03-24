package api

import (
	"github.com/goravel/framework/contracts/http"
)

// ---------- CreateSubgroupRequest ----------

type CreateSubgroupRequest struct {
	Name string `form:"name" json:"name"`
}

func (r *CreateSubgroupRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateSubgroupRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "required|min_len:1|max_len:50",
	}
}
func (r *CreateSubgroupRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "trim",
	}
}
func (r *CreateSubgroupRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required": "请输入小组名称",
		"name.min_len":  "小组名称不能为空",
		"name.max_len":  "小组名称不能超过50个字符",
	}
}

// ---------- UpdateSubgroupRequest ----------

type UpdateSubgroupRequest struct {
	Name string `form:"name" json:"name"`
}

func (r *UpdateSubgroupRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateSubgroupRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "required|min_len:1|max_len:50",
	}
}
func (r *UpdateSubgroupRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "trim",
	}
}
func (r *UpdateSubgroupRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required": "请输入小组名称",
		"name.min_len":  "小组名称不能为空",
		"name.max_len":  "小组名称不能超过50个字符",
	}
}

// ---------- AssignSubgroupMembersRequest ----------

type AssignSubgroupMembersRequest struct {
	UserIDs []string `form:"user_ids" json:"user_ids"`
}

func (r *AssignSubgroupMembersRequest) Authorize(ctx http.Context) error { return nil }
func (r *AssignSubgroupMembersRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"user_ids":   "required|min_len:1|max_len:50",
		"user_ids.*": "required|uuid",
	}
}
func (r *AssignSubgroupMembersRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{}
}
func (r *AssignSubgroupMembersRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"user_ids.required":   "请提供用户列表",
		"user_ids.min_len":    "请至少提供一个用户",
		"user_ids.max_len":    "最多一次分配50个用户",
		"user_ids.*.required": "用户ID不能为空",
		"user_ids.*.uuid":     "无效的用户ID",
	}
}
