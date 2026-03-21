package models

import "github.com/goravel/framework/database/orm"

type AdmRole struct {
	orm.Timestamps
	ID   string `gorm:"column:id;primaryKey" json:"id"`
	Slug string `gorm:"column:slug" json:"slug"`
	Name string `gorm:"column:name" json:"name"`
}

// TableName returns the database table name.
func (a *AdmRole) TableName() string {
	return "adm_roles"
}
