package models

import "github.com/goravel/framework/database/orm"

type Notice struct {
	orm.Timestamps
	ID       string  `gorm:"column:id;primaryKey" json:"id"`
	Title    string  `gorm:"column:title" json:"title"`
	Content  *string `gorm:"column:content" json:"content"`
	Icon     *string `gorm:"column:icon" json:"icon"`
	IsActive bool    `gorm:"column:is_active" json:"is_active"`
}

// TableName returns the database table name.
func (n *Notice) TableName() string {
	return "notices"
}
