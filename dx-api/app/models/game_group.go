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
	InviteCode    string  `gorm:"column:invite_code" json:"invite_code"`
	IsActive      bool    `gorm:"column:is_active" json:"is_active"`
}

func (g *GameGroup) TableName() string {
	return "game_groups"
}
