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

// SignInRequest for signin — supports email+code OR account+password.
// OR logic validated manually in controller (not expressible in Goravel rules).
type SignInRequest struct {
	Email    string `form:"email" json:"email"`
	Code     string `form:"code" json:"code"`
	Account  string `form:"account" json:"account"`
	Password string `form:"password" json:"password"`
}
