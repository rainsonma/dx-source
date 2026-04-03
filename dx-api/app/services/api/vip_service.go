package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// IsVipActive checks whether the user has an active VIP membership.
// Returns true if grade is "lifetime", or grade is not "free" and vipDueAt > now.
func IsVipActive(userID string) (bool, error) {
	var user models.User
	if err := facades.Orm().Query().Select("id", "grade", "vip_due_at").
		Where("id", userID).First(&user); err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return false, ErrUserNotFound
	}
	return checkVipActive(user), nil
}

// checkVipActive applies VIP logic to a loaded user.
func checkVipActive(user models.User) bool {
	if user.Grade == consts.UserGradeLifetime {
		return true
	}
	if user.Grade == consts.UserGradeFree {
		return false
	}
	return user.VipDueAt != nil && user.VipDueAt.StdTime().After(time.Now())
}

// isFirstLevel checks whether the given levelID is the first active level of a game.
func isFirstLevel(query orm.Query, gameID, levelID string) (bool, error) {
	var first models.GameLevel
	if err := query.Where("game_id", gameID).Where("is_active", true).
		Order("\"order\" asc").First(&first); err != nil || first.ID == "" {
		return false, ErrNoGameLevels
	}
	return first.ID == levelID, nil
}

// requireVipForLevel checks VIP status if the level is not the first level.
// Returns nil if the user can access the level, ErrVipRequired otherwise.
func requireVipForLevel(userID, gameID, levelID string) error {
	query := facades.Orm().Query()
	first, err := isFirstLevel(query, gameID, levelID)
	if err != nil {
		return err
	}
	if first {
		return nil
	}
	vip, err := IsVipActive(userID)
	if err != nil {
		return fmt.Errorf("failed to check VIP status: %w", err)
	}
	if !vip {
		return ErrVipRequired
	}
	return nil
}

// requireVip checks VIP status unconditionally.
// Returns nil if the user is VIP, ErrVipRequired otherwise.
func requireVip(userID string) error {
	vip, err := IsVipActive(userID)
	if err != nil {
		return fmt.Errorf("failed to check VIP status: %w", err)
	}
	if !vip {
		return ErrVipRequired
	}
	return nil
}
