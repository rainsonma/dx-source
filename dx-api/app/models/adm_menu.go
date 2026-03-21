package models

import "github.com/goravel/framework/database/orm"

type AdmMenu struct {
	orm.Timestamps
	ID       string  `gorm:"column:id;primaryKey" json:"id"`
	ParentID *string `gorm:"column:parent_id" json:"parent_id"`
	Name     string  `gorm:"column:name" json:"name"`
	Alias    *string `gorm:"column:alias" json:"alias"`
	Icon     *string `gorm:"column:icon" json:"icon"`
	Uri      *string `gorm:"column:uri" json:"uri"`
	Order    float64 `gorm:"column:order" json:"order"`
}

// TableName returns the database table name.
func (a *AdmMenu) TableName() string {
	return "adm_menus"
}
