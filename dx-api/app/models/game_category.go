package models

import "github.com/goravel/framework/database/orm"

type GameCategory struct {
	orm.Timestamps
	ID          string  `gorm:"column:id;primaryKey" json:"id"`
	ParentID    *string `gorm:"column:parent_id" json:"parent_id"`
	CoverID     *string `gorm:"column:cover_id" json:"cover_id"`
	Name        string  `gorm:"column:name" json:"name"`
	Alias       *string `gorm:"column:alias" json:"alias"`
	Description *string `gorm:"column:description" json:"description"`
	Order       float64 `gorm:"column:order" json:"order"`
	IsEnabled   bool    `gorm:"column:is_enabled" json:"is_enabled"`
}

func (g *GameCategory) TableName() string {
	return "game_categories"
}
