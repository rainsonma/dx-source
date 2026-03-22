package adm

import "errors"

var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrSessionReplaced     = errors.New("session replaced by another device")
	ErrAdminNotFound       = errors.New("admin user not found")
	ErrAdminInactive   = errors.New("admin account is inactive")
	ErrInvalidPassword = errors.New("invalid password")
	ErrNoticeNotFound  = errors.New("notice not found")
	ErrRedeemNotFound  = errors.New("redeem not found")
)
