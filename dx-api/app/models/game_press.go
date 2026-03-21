package models

import "github.com/goravel/framework/database/orm"

type GamePress struct {
	orm.Timestamps
	ID      string  `gorm:"column:id;primaryKey" json:"id"`
	Name    string  `gorm:"column:name" json:"name"`
	Alias   *string `gorm:"column:alias" json:"alias"`
	CoverID *string `gorm:"column:cover_id" json:"cover_id"`
	Order   float64 `gorm:"column:order" json:"order"`
}

func (g *GamePress) TableName() string {
	return "game_presses"
}
