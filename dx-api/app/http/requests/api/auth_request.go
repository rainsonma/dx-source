package api

import "github.com/goravel/framework/contracts/http"

// SendCodeRequest validates email for sending verification codes.
type SendCodeRequest struct {
	Email string `form:"email" json:"email"`
}

func (r *SendCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *SendCodeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email": "required",
	}
}

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
		"email": "required",
		"code":  "required|len:6",
	}
}
func (r *SignUpRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.required": "a 6-digit verification code is required",
		"code.len":      "a 6-digit verification code is required",
	}
}

// SignInRequest validates signin data.
// Supports two flows: email+code OR account+password.
type SignInRequest struct {
	// Email + code flow
	Email string `form:"email" json:"email"`
	Code  string `form:"code" json:"code"`

	// Account + password flow
	Account  string `form:"account" json:"account"`
	Password string `form:"password" json:"password"`
}

func (r *SignInRequest) Authorize(ctx http.Context) error { return nil }
func (r *SignInRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"email":    "required_without:account",
		"code":     "required_with:email|len:6",
		"account":  "required_without:email",
		"password": "required_with:account",
	}
}
func (r *SignInRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"email.required_without":   "email or account is required",
		"code.required_with":       "verification code is required",
		"code.len":                 "a 6-digit verification code is required",
		"account.required_without": "email or account is required",
		"password.required_with":   "password is required",
	}
}
