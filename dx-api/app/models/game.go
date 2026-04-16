package models

import "github.com/goravel/framework/database/orm"

type Game struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	Name           string  `gorm:"column:name" json:"name"`
	Description    *string `gorm:"column:description" json:"description"`
	UserID         *string `gorm:"column:user_id" json:"user_id"`
	Mode           string  `gorm:"column:mode" json:"mode"`
	GameCategoryID *string `gorm:"column:game_category_id" json:"game_category_id"`
	GamePressID    *string `gorm:"column:game_press_id" json:"game_press_id"`
	Icon           *string `gorm:"column:icon" json:"icon"`
	CoverID        *string `gorm:"column:cover_id" json:"cover_id"`
	Order          float64 `gorm:"column:order" json:"order"`
	IsActive       bool    `gorm:"column:is_active" json:"is_active"`
	Status         string  `gorm:"column:status" json:"status"`
	IsSelective    bool    `gorm:"column:is_selective" json:"is_selective"`
	IsPrivate      bool    `gorm:"column:is_private" json:"is_private"`
}

func (g *Game) TableName() string {
	return "games"
}
