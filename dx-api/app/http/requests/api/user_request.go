package api

import "github.com/goravel/framework/contracts/http"

// UpdateProfileRequest validates profile update data.
type UpdateProfileRequest struct {
	Nickname     string `form:"nickname" json:"nickname"`
	City         string `form:"city" json:"city"`
	Introduction string `form:"introduction" json:"introduction"`
}

func (r *UpdateProfileRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateProfileRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname":     "max_len:20",
		"city":         "max_len:50",
		"introduction": "max_len:200",
	}
}
func (r *UpdateProfileRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname":     "trim",
		"city":         "trim",
		"introduction": "trim",
	}
}
func (r *UpdateProfileRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.max_len":     "昵称不能超过20个字符",
		"city.max_len":         "城市不能超过50个字符",
		"introduction.max_len": "简介不能超过200个字符",
	}
}

// UpdateAvatarRequest validates avatar update data.
type UpdateAvatarRequest struct {
	AvatarURL string `form:"avatar_url" json:"avatar_url"`
}

func (r *UpdateAvatarRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateAvatarRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"avatar_url": "required",
	}
}
func (r *UpdateAvatarRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"avatar_url.required": "请选择头像",
	}
}

// ChangeEmailRequest validates email change data.
type ChangeEmailRequest struct {
	Email string `form:"email" json:"email"`
	Code  string `form:"code" json:"code"`
}

func (r *ChangeEmailRequest) Authorize(ctx http.Context) error { return nil }
func (r *ChangeEmailRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "required|email",
		"code":  "required|len:6",
	}
}
func (r *ChangeEmailRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "trim",
	}
}
func (r *ChangeEmailRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required": "请输入邮箱地址",
		"email.email":    "邮箱地址格式不正确",
		"code.required":  "请输入6位验证码",
		"code.len":       "请输入6位验证码",
	}
}

// ChangePasswordRequest validates password change data.
type ChangePasswordRequest struct {
	CurrentPassword string `form:"current_password" json:"current_password"`
	NewPassword     string `form:"new_password" json:"new_password"`
}

func (r *ChangePasswordRequest) Authorize(ctx http.Context) error { return nil }
func (r *ChangePasswordRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"current_password": "required",
		"new_password":     "required|min_len:8|strong_password",
	}
}
func (r *ChangePasswordRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"current_password.required":    "请输入当前密码",
		"new_password.required":        "请输入新密码",
		"new_password.min_len":         "新密码至少需要8个字符",
		"new_password.strong_password": "新密码必须包含大写字母、小写字母、数字和特殊字符",
	}
}
