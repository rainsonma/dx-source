package models

import "github.com/goravel/framework/database/orm"

type UserSetting struct {
	orm.Timestamps
	ID        string `gorm:"column:id;primaryKey" json:"id"`
	UserID    string `gorm:"column:user_id" json:"user_id"`
	Group     string `gorm:"column:group" json:"group"`
	Key       string `gorm:"column:key" json:"key"`
	Value     string `gorm:"column:value" json:"value"`
	ValueType string `gorm:"column:value_type" json:"value_type"`
}

// TableName returns the database table name.
func (u *UserSetting) TableName() string {
	return "user_settings"
}
