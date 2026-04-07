package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/lib/pq"
)

type GameLevel struct {
	orm.Timestamps
	orm.SoftDeletes
	ID           string         `gorm:"column:id;primaryKey" json:"id"`
	GameID       string         `gorm:"column:game_id" json:"game_id"`
	Name         string         `gorm:"column:name" json:"name"`
	Description  *string        `gorm:"column:description" json:"description"`
	Order        float64        `gorm:"column:order" json:"order"`
	PassingScore int            `gorm:"column:passing_score" json:"passing_score"`
	Degrees      pq.StringArray `gorm:"column:degrees;type:text[]" json:"degrees"`
	IsActive     bool           `gorm:"column:is_active" json:"is_active"`
}

func (g *GameLevel) TableName() string {
	return "game_levels"
}
