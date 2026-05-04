package models

import "github.com/goravel/framework/database/orm"

type GameVocab struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	GameID         string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string  `gorm:"column:game_level_id" json:"game_level_id"`
	ContentVocabID string  `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	Order          float64 `gorm:"column:order" json:"order"`
}

// TableName returns the database table name.
func (g *GameVocab) TableName() string {
	return "game_vocabs"
}
