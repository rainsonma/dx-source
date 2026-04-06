package api

import "github.com/goravel/framework/contracts/http"

type VerifyOnlineRequest struct {
	Username string `form:"username" json:"username"`
}

func (r *VerifyOnlineRequest) Authorize(ctx http.Context) error { return nil }

func (r *VerifyOnlineRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min_len:2|max_len:30",
	}
}

func (r *VerifyOnlineRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "trim",
	}
}

func (r *VerifyOnlineRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": "请输入用户名",
		"username.min_len":  "用户名至少2个字符",
		"username.max_len":  "用户名最长30个字符",
	}
}
