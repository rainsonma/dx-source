package models

import "github.com/goravel/framework/database/orm"

type ContentVocabEdit struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	ContentVocabID string  `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	EditorUserID   *string `gorm:"column:editor_user_id" json:"editor_user_id"`
	EditType       string  `gorm:"column:edit_type" json:"edit_type"`
	Before         *string `gorm:"column:before;type:jsonb" json:"before"`
	After          *string `gorm:"column:after;type:jsonb" json:"after"`
}

// TableName returns the database table name.
func (c *ContentVocabEdit) TableName() string {
	return "content_vocab_edits"
}
