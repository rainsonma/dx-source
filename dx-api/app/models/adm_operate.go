package models

import "github.com/goravel/framework/database/orm"

type AdmOperate struct {
	orm.Timestamps
	ID        string `gorm:"column:id;primaryKey" json:"id"`
	AdmUserID string `gorm:"column:adm_user_id" json:"adm_user_id"`
	Path      string `gorm:"column:path" json:"path"`
	Method    string `gorm:"column:method" json:"method"`
	Ip        string `gorm:"column:ip" json:"ip"`
	Input     string `gorm:"column:input" json:"input"`
}

// TableName returns the database table name.
func (a *AdmOperate) TableName() string {
	return "adm_operates"
}
