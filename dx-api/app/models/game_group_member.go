package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameGroupMember struct {
	orm.Timestamps
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	GameGroupID string     `gorm:"column:game_group_id" json:"game_group_id"`
	UserID      string     `gorm:"column:user_id" json:"user_id"`
	LastWonAt   *time.Time `gorm:"column:last_won_at" json:"last_won_at"`
}

func (g *GameGroupMember) TableName() string {
	return "game_group_members"
}
