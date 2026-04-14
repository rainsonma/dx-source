package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/lib/pq"
)

type ContentItem struct {
	orm.Timestamps
	orm.SoftDeletes
	ID            string         `gorm:"column:id;primaryKey" json:"id"`
	ContentMetaID *string        `gorm:"column:content_meta_id" json:"content_meta_id"`
	Content       string         `gorm:"column:content" json:"content"`
	ContentType   string         `gorm:"column:content_type" json:"content_type"`
	UkAudioID     *string        `gorm:"column:uk_audio_id" json:"uk_audio_id"`
	UsAudioID     *string        `gorm:"column:us_audio_id" json:"us_audio_id"`
	Definition    *string        `gorm:"column:definition" json:"definition"`
	Translation   *string        `gorm:"column:translation" json:"translation"`
	Explanation   *string        `gorm:"column:explanation" json:"explanation"`
	Items         *string        `gorm:"column:items;type:jsonb" json:"items"`
	Structure     *string        `gorm:"column:structure;type:jsonb" json:"structure"`
	Tags          pq.StringArray `gorm:"column:tags;type:text[]" json:"tags"`
}

// TableName returns the database table name.
func (c *ContentItem) TableName() string {
	return "content_items"
}
