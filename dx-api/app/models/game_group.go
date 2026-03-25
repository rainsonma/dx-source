package models

import "github.com/goravel/framework/database/orm"

type GameGroup struct {
	orm.Timestamps
	ID            string  `gorm:"column:id;primaryKey" json:"id"`
	Name          string  `gorm:"column:name" json:"name"`
	Description   *string `gorm:"column:description" json:"description"`
	OwnerID       string  `gorm:"column:owner_id" json:"owner_id"`
	CoverID       *string `gorm:"column:cover_id" json:"cover_id"`
	CurrentGameID *string `gorm:"column:current_game_id" json:"current_game_id"`
	InviteCode       string  `gorm:"column:invite_code" json:"invite_code"`
	InviteQrcodeID   *string `gorm:"column:invite_qrcode_id" json:"invite_qrcode_id"`
	IsActive      bool    `gorm:"column:is_active" json:"is_active"`
	MemberCount   int     `gorm:"column:member_count" json:"member_count"`
	GameMode      *string `gorm:"column:game_mode" json:"game_mode"`
}

func (g *GameGroup) TableName() string {
	return "game_groups"
}
