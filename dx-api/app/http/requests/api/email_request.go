package api

import "github.com/goravel/framework/contracts/http"

// SendCodeRequest validates email for sending verification codes.
type SendCodeRequest struct {
	Email string `form:"email" json:"email"`
}

func (r *SendCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SendCodeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "required|email",
	}
}
func (r *SendCodeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "trim",
	}
}
func (r *SendCodeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required": "请输入邮箱地址",
		"email.email":    "邮箱地址格式不正确",
	}
}
