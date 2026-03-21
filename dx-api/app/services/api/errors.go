package api

import "errors"

var (
	ErrRateLimited       = errors.New("rate limited")
	ErrInvalidCode       = errors.New("invalid or expired verification code")
	ErrDuplicateEmail    = errors.New("email already registered")
	ErrDuplicateUsername = errors.New("username already taken")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
)
