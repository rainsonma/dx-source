package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameGroup struct {
	orm.Timestamps
	ID               string     `gorm:"column:id;primaryKey" json:"id"`
	Name             string     `gorm:"column:name" json:"name"`
	Description      *string    `gorm:"column:description" json:"description"`
	OwnerID          string     `gorm:"column:owner_id" json:"owner_id"`
	CoverURL         *string    `gorm:"column:cover_url" json:"cover_url"`
	CurrentGameID    *string    `gorm:"column:current_game_id" json:"current_game_id"`
	StartGameLevelID *string    `gorm:"column:start_game_level_id" json:"start_game_level_id"`
	InviteCode       string     `gorm:"column:invite_code" json:"invite_code"`
	InviteQrcodeURL  *string    `gorm:"column:invite_qrcode_url" json:"invite_qrcode_url"`
	DismissedAt      *time.Time `gorm:"column:dismissed_at" json:"dismissed_at"`
	MemberCount      int        `gorm:"column:member_count" json:"member_count"`
	GameMode         *string    `gorm:"column:game_mode" json:"game_mode"`
	IsPlaying        bool       `gorm:"column:is_playing" json:"is_playing"`
}

func (g *GameGroup) TableName() string {
	return "game_groups"
}
