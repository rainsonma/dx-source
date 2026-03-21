package models

import "github.com/goravel/framework/database/orm"

type Setting struct {
	orm.Timestamps
	ID           string  `gorm:"column:id;primaryKey" json:"id"`
	Group        string  `gorm:"column:group" json:"group"`
	Label        *string `gorm:"column:label" json:"label"`
	Key          string  `gorm:"column:key" json:"key"`
	Value        string  `gorm:"column:value" json:"value"`
	ValueType    string  `gorm:"column:value_type" json:"value_type"`
	ValueFrom    string  `gorm:"column:value_from" json:"value_from"`
	ValueOptions string  `gorm:"column:value_options;type:jsonb" json:"value_options"`
	Description  string  `gorm:"column:description" json:"description"`
	Order        float64 `gorm:"column:order" json:"order"`
	IsEnabled    bool    `gorm:"column:is_enabled" json:"is_enabled"`
}

// TableName returns the database table name.
func (s *Setting) TableName() string {
	return "settings"
}
