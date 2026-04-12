package api

import (
	"context"
	"fmt"

	"github.com/goravel/framework/facades"

	"dx-api/app/models"
	"dx-api/app/realtime"
)

type VerifyOnlineResult struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	IsOnline bool   `json:"is_online"`
	IsVip    bool   `json:"is_vip"`
}

// VerifyUserOnline checks if a user exists, is online, and has active VIP.
func VerifyUserOnline(callerID, username string) (*VerifyOnlineResult, error) {
	var user models.User
	if err := facades.Orm().Query().
		Select("id", "username", "nickname", "grade", "vip_due_at").
		Where("username", username).
		First(&user); err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return nil, ErrUserNotFound
	}

	if user.ID == callerID {
		return nil, ErrCannotChallengeSelf
	}

	isOnline, _ := realtime.DefaultHub().Presence().IsPresent(context.Background(), realtime.UserTopic(user.ID), user.ID)
	if !isOnline {
		return &VerifyOnlineResult{
			UserID:   user.ID,
			Nickname: nickname(user),
			IsOnline: false,
			IsVip:    false,
		}, nil
	}

	vipActive := checkVipActive(user)

	return &VerifyOnlineResult{
		UserID:   user.ID,
		Nickname: nickname(user),
		IsOnline: true,
		IsVip:    vipActive,
	}, nil
}

func nickname(user models.User) string {
	if user.Nickname != nil && *user.Nickname != "" {
		return *user.Nickname
	}
	return user.Username
}
