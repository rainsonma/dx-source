package api

// SendCodeRequest validates email for sending verification codes.
type SendCodeRequest struct {
	Email string `form:"email" json:"email"`
}

// SignUpRequest validates signup data.
type SignUpRequest struct {
	Email    string `form:"email" json:"email"`
	Code     string `form:"code" json:"code"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
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
