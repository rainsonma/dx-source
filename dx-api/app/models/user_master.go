package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type UserMaster struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string           `gorm:"column:id;primaryKey" json:"id"`
	UserID         string           `gorm:"column:user_id" json:"user_id"`
	ContentItemID  *string          `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string          `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	GameID         string           `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string           `gorm:"column:game_level_id" json:"game_level_id"`
	MasteredAt     *carbon.DateTime `gorm:"column:mastered_at" json:"mastered_at"`
}

// TableName returns the database table name.
func (u *UserMaster) TableName() string {
	return "user_masters"
}
