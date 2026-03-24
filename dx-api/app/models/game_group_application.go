package models

import "github.com/goravel/framework/database/orm"

type GameGroupApplication struct {
	orm.Timestamps
	ID          string `gorm:"column:id;primaryKey" json:"id"`
	GameGroupID string `gorm:"column:game_group_id" json:"game_group_id"`
	UserID      string `gorm:"column:user_id" json:"user_id"`
	Status      string `gorm:"column:status" json:"status"`
}

func (g *GameGroupApplication) TableName() string {
	return "game_group_applications"
}
