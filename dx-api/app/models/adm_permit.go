package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/lib/pq"
)

type AdmPermit struct {
	orm.Timestamps
	ID          string         `gorm:"column:id;primaryKey" json:"id"`
	Slug        string         `gorm:"column:slug" json:"slug"`
	Name        string         `gorm:"column:name" json:"name"`
	HttpMethods pq.StringArray `gorm:"column:http_methods;type:text[]" json:"http_methods"`
	HttpPaths   pq.StringArray `gorm:"column:http_paths;type:text[]" json:"http_paths"`
}

// TableName returns the database table name.
func (a *AdmPermit) TableName() string {
	return "adm_permits"
}
