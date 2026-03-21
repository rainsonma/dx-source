package models

import "github.com/goravel/framework/database/orm"

type ContentSeek struct {
	orm.Timestamps
	ID          string `gorm:"column:id;primaryKey" json:"id"`
	UserID      string `gorm:"column:user_id" json:"user_id"`
	CourseName  string `gorm:"column:course_name" json:"course_name"`
	Description string `gorm:"column:description" json:"description"`
	DiskUrl     string `gorm:"column:disk_url" json:"disk_url"`
	Count       int    `gorm:"column:count" json:"count"`
}

// TableName returns the database table name.
func (c *ContentSeek) TableName() string {
	return "content_seeks"
}
