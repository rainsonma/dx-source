package api

// UpdateProfileRequest validates profile update data.
type UpdateProfileRequest struct {
	Nickname     string `form:"nickname" json:"nickname"`
	City         string `form:"city" json:"city"`
	Introduction string `form:"introduction" json:"introduction"`
}

// UpdateAvatarRequest validates avatar update data.
type UpdateAvatarRequest struct {
	ImageID string `form:"image_id" json:"image_id"`
}

// SendEmailCodeRequest validates email code sending data.
type SendEmailCodeRequest struct {
	Email string `form:"email" json:"email"`
}

// ChangeEmailRequest validates email change data.
type ChangeEmailRequest struct {
	Email string `form:"email" json:"email"`
	Code  string `form:"code" json:"code"`
}

// ChangePasswordRequest validates password change data.
type ChangePasswordRequest struct {
	CurrentPassword string `form:"current_password" json:"current_password"`
	NewPassword     string `form:"new_password" json:"new_password"`
}
