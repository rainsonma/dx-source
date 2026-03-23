package rules

import (
	"context"
	"unicode"

	"github.com/goravel/framework/contracts/validation"
)

type StrongPassword struct{}

func (r *StrongPassword) Signature() string {
	return "strong_password"
}

func (r *StrongPassword) Passes(_ context.Context, _ validation.Data, val any, _ ...any) bool {
	s, ok := val.(string)
	if !ok || s == "" {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func (r *StrongPassword) Message(_ context.Context) string {
	return ":attribute 必须包含大写字母、小写字母、数字和特殊字符"
}
