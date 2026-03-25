package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameSubgroup struct {
	orm.Timestamps
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	GameGroupID string     `gorm:"column:game_group_id" json:"game_group_id"`
	Name        string     `gorm:"column:name" json:"name"`
	Description *string    `gorm:"column:description" json:"description"`
	Order       float64    `gorm:"column:order" json:"order"`
	LastWonAt   *time.Time `gorm:"column:last_won_at" json:"last_won_at"`
}

func (g *GameSubgroup) TableName() string {
	return "game_subgroups"
}
