package models

import "github.com/goravel/framework/database/orm"

type ContentMeta struct {
	orm.Timestamps
	orm.SoftDeletes
	ID          string  `gorm:"column:id;primaryKey" json:"id"`
	GameLevelID string  `gorm:"column:game_level_id" json:"game_level_id"`
	SourceFrom  string  `gorm:"column:source_from" json:"source_from"`
	SourceType  string  `gorm:"column:source_type" json:"source_type"`
	SourceData  string  `gorm:"column:source_data" json:"source_data"`
	Translation *string `gorm:"column:translation" json:"translation"`
	IsBreakDone bool    `gorm:"column:is_break_done" json:"is_break_done"`
	Order       float64 `gorm:"column:order" json:"order"`
}

// TableName returns the database table name.
func (c *ContentMeta) TableName() string {
	return "content_metas"
}
