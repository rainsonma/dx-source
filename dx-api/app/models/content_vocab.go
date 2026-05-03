package models

import "github.com/goravel/framework/database/orm"

type ContentVocab struct {
	orm.Timestamps
	orm.SoftDeletes
	ID           string  `gorm:"column:id;primaryKey" json:"id"`
	Content      string  `gorm:"column:content" json:"content"`
	ContentKey   string  `gorm:"column:content_key" json:"content_key"`
	UkPhonetic   *string `gorm:"column:uk_phonetic" json:"uk_phonetic"`
	UsPhonetic   *string `gorm:"column:us_phonetic" json:"us_phonetic"`
	UkAudioURL   *string `gorm:"column:uk_audio_url" json:"uk_audio_url"`
	UsAudioURL   *string `gorm:"column:us_audio_url" json:"us_audio_url"`
	Definition   *string `gorm:"column:definition;type:jsonb" json:"definition"`
	Explanation  *string `gorm:"column:explanation" json:"explanation"`
	IsVerified   bool    `gorm:"column:is_verified" json:"is_verified"`
	CreatedBy    *string `gorm:"column:created_by" json:"created_by"`
	LastEditedBy *string `gorm:"column:last_edited_by" json:"last_edited_by"`
}

// TableName returns the database table name.
func (c *ContentVocab) TableName() string {
	return "content_vocabs"
}
