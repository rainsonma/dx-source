package api

import "errors"

var (
	ErrRateLimited       = errors.New("rate limited")
	ErrInvalidCode       = errors.New("invalid or expired verification code")
	ErrDuplicateEmail    = errors.New("email already registered")
	ErrDuplicateUsername = errors.New("username already taken")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrNicknameTaken     = errors.New("nickname already taken")
	ErrImageNotFound     = errors.New("image not found")
	ErrImageNotOwned     = errors.New("image not owned by user")
	ErrGameNotFound      = errors.New("game not found")
	ErrSessionNotFound   = errors.New("session not found")
	ErrLevelNotFound     = errors.New("level not found")
	ErrSessionLevelNotFound = errors.New("session level not found")
	ErrNoGameLevels      = errors.New("game has no levels")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidPlayTime   = errors.New("invalid play time")
	ErrInsufficientBeans = errors.New("insufficient beans")
	ErrRedeemNotFound    = errors.New("redeem code not found")
	ErrRedeemAlreadyUsed = errors.New("redeem code already used")
	ErrContentSeekExists = errors.New("content seek already exists")
	ErrFileTooLarge      = errors.New("file size exceeds 2MB limit")
	ErrInvalidFileType   = errors.New("only JPEG and PNG files are allowed")
	ErrInvalidImageRole  = errors.New("invalid image role")
)
