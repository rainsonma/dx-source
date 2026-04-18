package api

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// ProfileData holds the user profile with computed fields.
type ProfileData struct {
	ID                string  `json:"id"`
	Grade             string  `json:"grade"`
	Username          string  `json:"username"`
	Nickname          *string `json:"nickname"`
	Email             *string `json:"email"`
	Phone             *string `json:"phone"`
	AvatarURL         *string `json:"avatar_url"`
	City              *string `json:"city"`
	Introduction      *string `json:"introduction"`
	IsActive          bool    `json:"is_active"`
	Beans             int     `json:"beans"`
	GrantedBeans      int     `json:"granted_beans"`
	Exp               int     `json:"exp"`
	Level             int     `json:"level"`
	InviteCode        string  `json:"invite_code"`
	CurrentPlayStreak int     `json:"current_play_streak"`
	MaxPlayStreak     int     `json:"max_play_streak"`
	LastPlayedAt      any     `json:"last_played_at"`
	VipDueAt          any     `json:"vip_due_at"`
	LastReadNoticeAt  any     `json:"last_read_notice_at"`
	CreatedAt         any     `json:"created_at"`
	UpdatedAt         any     `json:"updated_at"`
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

	profile := &ProfileData{
		ID:                user.ID,
		Grade:             user.Grade,
		Username:          user.Username,
		Nickname:          user.Nickname,
		Email:             user.Email,
		Phone:             user.Phone,
		AvatarURL:         user.AvatarURL,
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
	}

	updates := map[string]any{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if city != "" {
		updates["city"] = city
	}
	if introduction != "" {
		updates["introduction"] = introduction
	}
	if len(updates) == 0 {
		return nil
	}

	if _, err := facades.Orm().Query().Model(&models.User{}).Where("id", userID).Update(updates); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// UpdateAvatar sets the user's avatar URL directly.
func UpdateAvatar(userID, avatarURL string) error {
	if _, err := facades.Orm().Query().Model(&models.User{}).Where("id", userID).Update("avatar_url", avatarURL); err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
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

	if _, err := facades.Orm().Query().Model(&models.User{}).Where("id", userID).Update("email", email); err != nil {
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

	if _, err := facades.Orm().Query().Model(&models.User{}).Where("id", userID).Update("password", hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
