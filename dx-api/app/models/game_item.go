package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameItem struct {
	orm.SoftDeletes
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	GameID        string    `gorm:"column:game_id" json:"game_id"`
	GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
	ContentItemID string    `gorm:"column:content_item_id" json:"content_item_id"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (g *GameItem) TableName() string {
	return "game_items"
}
