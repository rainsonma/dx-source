package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type UserReview struct {
	orm.Timestamps
	ID            string           `gorm:"column:id;primaryKey" json:"id"`
	UserID        string           `gorm:"column:user_id" json:"user_id"`
	ContentItemID string           `gorm:"column:content_item_id" json:"content_item_id"`
	GameID        string           `gorm:"column:game_id" json:"game_id"`
	GameLevelID   string           `gorm:"column:game_level_id" json:"game_level_id"`
	LastReviewAt  *carbon.DateTime `gorm:"column:last_review_at" json:"last_review_at"`
	NextReviewAt  *carbon.DateTime `gorm:"column:next_review_at" json:"next_review_at"`
	ReviewCount   int              `gorm:"column:review_count" json:"review_count"`
}

// TableName returns the database table name.
func (u *UserReview) TableName() string {
	return "user_reviews"
}
