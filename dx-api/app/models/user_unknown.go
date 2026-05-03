package models

import "github.com/goravel/framework/database/orm"

type UserUnknown struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	UserID         string  `gorm:"column:user_id" json:"user_id"`
	ContentItemID  *string `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	GameID         string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string  `gorm:"column:game_level_id" json:"game_level_id"`
}

// TableName returns the database table name.
func (u *UserUnknown) TableName() string {
	return "user_unknowns"
}
