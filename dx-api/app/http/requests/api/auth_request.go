package api

import "github.com/goravel/framework/contracts/http"

// SignUpRequest validates signup data.
type SignUpRequest struct {
	Email    string `form:"email" json:"email"`
	Code     string `form:"code" json:"code"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

func (r *SignUpRequest) Authorize(ctx http.Context) error { return nil }
func (r *SignUpRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email":    "required|email",
		"code":     "required|len:6",
		"username": "required|alpha_dash|min_len:3|max_len:20",
		"password": "required|min_len:8|strong_password",
	}
}
func (r *SignUpRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email":    "trim",
		"username": "trim",
	}
}
func (r *SignUpRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required":           "请输入邮箱地址",
		"email.email":              "邮箱地址格式不正确",
		"code.required":            "请输入6位验证码",
		"code.len":                 "请输入6位验证码",
		"username.required":        "请输入用户名",
		"username.alpha_dash":      "用户名只能包含字母、数字、下划线和横线",
		"username.min_len":         "用户名至少需要3个字符",
		"username.max_len":         "用户名不能超过20个字符",
		"password.required":        "请输入密码",
		"password.min_len":         "密码至少需要8个字符",
		"password.strong_password": "密码必须包含大写字母、小写字母、数字和特殊字符",
	}
}

// SignInRequest for signin — supports email+code OR account+password.
// Controller still validates the pairing logic (exactly one auth method).
type SignInRequest struct {
	Email    string `form:"email" json:"email"`
	Code     string `form:"code" json:"code"`
	Account  string `form:"account" json:"account"`
	Password string `form:"password" json:"password"`
}

func (r *SignInRequest) Authorize(ctx http.Context) error { return nil }
func (r *SignInRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email":    "required_without:account|email",
		"code":     "required_without:password|len:6",
		"account":  "required_without:email|min_len:3",
		"password": "required_without:code|min_len:8",
	}
}
func (r *SignInRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email":   "trim",
		"account": "trim",
	}
}
func (r *SignInRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.email":      "邮箱地址格式不正确",
		"code.len":         "请输入6位验证码",
		"account.min_len":  "账号至少需要3个字符",
		"password.min_len": "密码至少需要8个字符",
	}
}
