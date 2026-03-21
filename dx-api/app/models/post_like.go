package models

import "github.com/goravel/framework/support/carbon"

type PostLike struct {
	ID        string           `gorm:"column:id;primaryKey" json:"id"`
	PostID    string           `gorm:"column:post_id" json:"post_id"`
	UserID    string           `gorm:"column:user_id" json:"user_id"`
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
}

// TableName returns the database table name.
func (p *PostLike) TableName() string {
	return "post_likes"
}
