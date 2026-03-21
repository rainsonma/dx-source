package models

import "github.com/goravel/framework/support/carbon"

type UserFollow struct {
	ID          string           `gorm:"column:id;primaryKey" json:"id"`
	FollowerID  string           `gorm:"column:follower_id" json:"follower_id"`
	FollowingID string           `gorm:"column:following_id" json:"following_id"`
	CreatedAt   *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
}

// TableName returns the database table name.
func (u *UserFollow) TableName() string {
	return "user_follows"
}
