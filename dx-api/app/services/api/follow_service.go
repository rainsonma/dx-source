package api

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"

	"dx-api/app/models"
)

type ToggleFollowResult struct {
	Followed bool `json:"followed"`
}

func ToggleFollow(userID, targetUserID string) (*ToggleFollowResult, error) {
	if userID == targetUserID {
		return nil, ErrSelfFollow
	}

	query := facades.Orm().Query()

	var target models.User
	if err := query.Where("id", targetUserID).First(&target); err != nil || target.ID == "" {
		return nil, ErrUserNotFound
	}

	var existing models.UserFollow
	if err := query.Where("follower_id", userID).Where("following_id", targetUserID).First(&existing); err == nil && existing.ID != "" {
		if _, err := query.Exec("DELETE FROM user_follows WHERE id = ?", existing.ID); err != nil {
			return nil, fmt.Errorf("failed to unfollow: %w", err)
		}
		return &ToggleFollowResult{Followed: false}, nil
	}

	follow := models.UserFollow{
		ID:          uuid.Must(uuid.NewV7()).String(),
		FollowerID:  userID,
		FollowingID: targetUserID,
	}
	if err := query.Create(&follow); err != nil {
		return nil, fmt.Errorf("failed to follow: %w", err)
	}

	return &ToggleFollowResult{Followed: true}, nil
}
