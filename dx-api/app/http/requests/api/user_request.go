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
func (r *UpdateProfileRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.max_len":     "nickname must be at most 20 characters",
		"city.max_len":         "city must be at most 50 characters",
		"introduction.max_len": "introduction must be at most 200 characters",
	}
}

// UpdateAvatarRequest validates avatar update data.
type UpdateAvatarRequest struct {
	ImageID string `form:"image_id" json:"image_id"`
}

func (r *UpdateAvatarRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateAvatarRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"image_id": "required",
	}
}

// SendEmailCodeRequest validates email code sending data.
type SendEmailCodeRequest struct {
	Email string `form:"email" json:"email"`
}

func (r *SendEmailCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SendEmailCodeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "required",
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
		"email": "required",
		"code":  "required|len:6",
	}
}
func (r *ChangeEmailRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.required": "a 6-digit verification code is required",
		"code.len":      "a 6-digit verification code is required",
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
		"new_password":     "required|min_len:8",
	}
}
func (r *ChangePasswordRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"new_password.min_len": "new password must be at least 8 characters",
	}
}
