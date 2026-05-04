package models

import "github.com/goravel/framework/database/orm"

type GameRecord struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	UserID         string  `gorm:"column:user_id" json:"user_id"`
	GameSessionID  string  `gorm:"column:game_session_id" json:"game_session_id"`
	GameLevelID    string  `gorm:"column:game_level_id" json:"game_level_id"`
	ContentItemID  *string `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	IsCorrect      bool    `gorm:"column:is_correct" json:"is_correct"`
	SourceAnswer   string  `gorm:"column:source_answer" json:"source_answer"`
	UserAnswer     string  `gorm:"column:user_answer" json:"user_answer"`
	BaseScore      int     `gorm:"column:base_score" json:"base_score"`
	ComboScore     int     `gorm:"column:combo_score" json:"combo_score"`
	Duration       int     `gorm:"column:duration" json:"duration"`
}

func (g *GameRecord) TableName() string {
	return "game_records"
}
