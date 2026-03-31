package models

import "github.com/goravel/framework/database/orm"

type PostComment struct {
	orm.Timestamps
	ID        string  `gorm:"column:id;primaryKey" json:"id"`
	PostID    string  `gorm:"column:post_id" json:"post_id"`
	UserID    string  `gorm:"column:user_id" json:"user_id"`
	Content   string  `gorm:"column:content" json:"content"`
	ParentID  *string `gorm:"column:parent_id" json:"parent_id"`
	LikeCount int     `gorm:"column:like_count" json:"like_count"`
}

func (p *PostComment) TableName() string {
	return "post_comments"
}
