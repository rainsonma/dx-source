package models

import "github.com/goravel/framework/database/orm"

type Feedback struct {
	orm.Timestamps
	ID          string `gorm:"column:id;primaryKey" json:"id"`
	UserID      string `gorm:"column:user_id" json:"user_id"`
	Type        string `gorm:"column:type" json:"type"`
	Description string `gorm:"column:description" json:"description"`
	Count       int    `gorm:"column:count" json:"count"`
}

// TableName returns the database table name.
func (f *Feedback) TableName() string {
	return "feedbacks"
}
