package models

import "github.com/goravel/framework/database/orm"

type UserBean struct {
	orm.Timestamps
	ID     string  `gorm:"column:id;primaryKey" json:"id"`
	UserID string  `gorm:"column:user_id" json:"user_id"`
	Beans  int     `gorm:"column:beans" json:"beans"`
	Origin int     `gorm:"column:origin" json:"origin"`
	Result int     `gorm:"column:result" json:"result"`
	Slug   string  `gorm:"column:slug" json:"slug"`
	Reason string  `gorm:"column:reason" json:"reason"`
	Data   *string `gorm:"column:data" json:"data"`
}

// TableName returns the database table name.
func (u *UserBean) TableName() string {
	return "user_beans"
}
