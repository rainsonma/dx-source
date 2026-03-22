package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/services/com"
)

// ProfileData holds the user profile with computed fields.
type ProfileData struct {
	ID               string `json:"id"`
	Grade            string `json:"grade"`
	Username         string `json:"username"`
	Nickname         *string `json:"nickname"`
	Email            *string `json:"email"`
	Phone            *string `json:"phone"`
	AvatarID         *string `json:"avatar_id"`
	AvatarURL        *string `json:"avatar_url"`
	City             *string `json:"city"`
	Introduction     *string `json:"introduction"`
	IsActive         bool   `json:"is_active"`
	Beans            int    `json:"beans"`
	GrantedBeans     int    `json:"granted_beans"`
	Exp              int    `json:"exp"`
	Level            int    `json:"level"`
	InviteCode       string `json:"invite_code"`
	CurrentPlayStreak int   `json:"current_play_streak"`
	MaxPlayStreak    int    `json:"max_play_streak"`
	LastPlayedAt     any    `json:"last_played_at"`
	VipDueAt         any    `json:"vip_due_at"`
	LastReadNoticeAt any    `json:"last_read_notice_at"`
	CreatedAt        any    `json:"created_at"`
	UpdatedAt        any    `json:"updated_at"`
}

// GetProfile retrieves a user profile by ID with computed level and avatar URL.
func GetProfile(userID string) (*ProfileData, error) {
	var user models.User
	if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return nil, ErrUserNotFound
	}

	level, err := consts.GetLevel(user.Exp)
	if err != nil {
		return nil, fmt.Errorf("failed to compute level: %w", err)
	}

	var avatarURL *string
	if user.AvatarID != nil && *user.AvatarID != "" {
		url := helpers.ImageServeURL(*user.AvatarID)
		avatarURL = &url
	}

	profile := &ProfileData{
		ID:                user.ID,
		Grade:             user.Grade,
		Username:          user.Username,
		Nickname:          user.Nickname,
		Email:             user.Email,
		Phone:             user.Phone,
		AvatarID:          user.AvatarID,
		AvatarURL:         avatarURL,
		City:              user.City,
		Introduction:      user.Introduction,
		IsActive:          user.IsActive,
		Beans:             user.Beans,
		GrantedBeans:      user.GrantedBeans,
		Exp:               user.Exp,
		Level:             level,
		InviteCode:        user.InviteCode,
		CurrentPlayStreak: user.CurrentPlayStreak,
		MaxPlayStreak:     user.MaxPlayStreak,
		LastPlayedAt:      user.LastPlayedAt,
		VipDueAt:          user.VipDueAt,
		LastReadNoticeAt:  user.LastReadNoticeAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}

	return profile, nil
}

// UpdateProfile updates the user's nickname, city, and introduction.
func UpdateProfile(userID, nickname, city, introduction string) error {
	var user models.User
	if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return ErrUserNotFound
	}

	// Check nickname uniqueness if changed
	if nickname != "" {
		currentNickname := ""
		if user.Nickname != nil {
			currentNickname = *user.Nickname
		}
		if nickname != currentNickname {
			var existing models.User
			err := facades.Orm().Query().Where("nickname", nickname).Where("id != ?", userID).First(&existing)
			if err == nil && existing.ID != "" {
				return ErrNicknameTaken
			}
		}
		user.Nickname = &nickname
	}

	if city != "" {
		user.City = &city
	}

	if introduction != "" {
		user.Introduction = &introduction
	}

	if err := facades.Orm().Query().Save(&user); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// UpdateAvatar sets the user's avatar from an image ID.
func UpdateAvatar(userID, imageID string) error {
	var image models.Image
	if err := facades.Orm().Query().Where("id", imageID).First(&image); err != nil {
		return fmt.Errorf("failed to find image: %w", err)
	}
	if image.ID == "" {
		return ErrImageNotFound
	}

	if image.UserID == nil || *image.UserID != userID {
		return ErrImageNotOwned
	}

	var user models.User
	if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return ErrUserNotFound
	}

	user.AvatarID = &imageID
	if err := facades.Orm().Query().Save(&user); err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	return nil
}

// SendChangeEmailCode sends a verification code for changing email.
func SendChangeEmailCode(userID, email string) error {
	key := fmt.Sprintf("change_email_code:%s", userID)

	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:change_email_code:%s", userID), 1, 60)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	// Check email not already taken by another user
	var existing models.User
	err = facades.Orm().Query().Where("email", email).Where("id != ?", userID).First(&existing)
	if err == nil && existing.ID != "" {
		return ErrDuplicateEmail
	}

	code := helpers.GenerateCode(6)
	if err := helpers.RedisSet(key, code, 300*time.Second); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	if err := com.SendVerificationEmail(email, code); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// ChangeEmail changes the user's email after verifying the code.
func ChangeEmail(userID, email, code string) error {
	key := fmt.Sprintf("change_email_code:%s", userID)
	storedCode, err := helpers.RedisGet(key)
	if err != nil || storedCode != code {
		return ErrInvalidCode
	}

	// Recheck email uniqueness
	var existing models.User
	err = facades.Orm().Query().Where("email", email).Where("id != ?", userID).First(&existing)
	if err == nil && existing.ID != "" {
		return ErrDuplicateEmail
	}

	var user models.User
	if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return ErrUserNotFound
	}

	user.Email = &email
	if err := facades.Orm().Query().Save(&user); err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	_ = helpers.RedisDel(key)

	return nil
}

// ChangePassword changes the user's password after verifying the current password.
func ChangePassword(userID, currentPassword, newPassword string) error {
	var user models.User
	if err := facades.Orm().Query().Where("id", userID).Select("id", "password").First(&user); err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return ErrUserNotFound
	}

	if !helpers.CheckPassword(currentPassword, user.Password) {
		return ErrInvalidPassword
	}

	hashedPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = hashedPassword
	if err := facades.Orm().Query().Save(&user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
