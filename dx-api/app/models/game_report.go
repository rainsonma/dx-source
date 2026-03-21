package models

import "github.com/goravel/framework/database/orm"

type GameReport struct {
	orm.Timestamps
	ID            string  `gorm:"column:id;primaryKey" json:"id"`
	UserID        string  `gorm:"column:user_id" json:"user_id"`
	GameID        string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID   string  `gorm:"column:game_level_id" json:"game_level_id"`
	ContentItemID string  `gorm:"column:content_item_id" json:"content_item_id"`
	Reason        string  `gorm:"column:reason" json:"reason"`
	Note          *string `gorm:"column:note" json:"note"`
	Count         int     `gorm:"column:count" json:"count"`
}

func (g *GameReport) TableName() string {
	return "game_reports"
}
