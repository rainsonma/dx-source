package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/lib/pq"
)

type Post struct {
	orm.Timestamps
	ID           string         `gorm:"column:id;primaryKey" json:"id"`
	UserID       string         `gorm:"column:user_id" json:"user_id"`
	Content      string         `gorm:"column:content" json:"content"`
	ImageID      *string        `gorm:"column:image_id" json:"image_id"`
	Tags         pq.StringArray `gorm:"column:tags;type:text[]" json:"tags"`
	LikeCount    int            `gorm:"column:like_count" json:"like_count"`
	CommentCount int            `gorm:"column:comment_count" json:"comment_count"`
	ShareCount   int            `gorm:"column:share_count" json:"share_count"`
	IsActive     bool           `gorm:"column:is_active" json:"is_active"`
}

// TableName returns the database table name.
func (p *Post) TableName() string {
	return "posts"
}
