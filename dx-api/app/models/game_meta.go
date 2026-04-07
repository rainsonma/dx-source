package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameMeta struct {
	orm.SoftDeletes
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	GameID        string    `gorm:"column:game_id" json:"game_id"`
	GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
	ContentMetaID string    `gorm:"column:content_meta_id" json:"content_meta_id"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (g *GameMeta) TableName() string {
	return "game_metas"
}
