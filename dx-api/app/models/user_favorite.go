package models

import "github.com/goravel/framework/support/carbon"

type UserFavorite struct {
	ID        string           `gorm:"column:id;primaryKey" json:"id"`
	UserID    string           `gorm:"column:user_id" json:"user_id"`
	GameID    string           `gorm:"column:game_id" json:"game_id"`
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
}

// TableName returns the database table name.
func (u *UserFavorite) TableName() string {
	return "user_favorites"
}
