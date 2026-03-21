package adm

import "errors"

var (
	ErrAdminNotFound   = errors.New("admin user not found")
	ErrAdminInactive   = errors.New("admin account is inactive")
	ErrInvalidPassword = errors.New("invalid password")
)
