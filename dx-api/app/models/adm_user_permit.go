package models

import "github.com/goravel/framework/database/orm"

type AdmUserPermit struct {
	orm.Timestamps
	ID          string `gorm:"column:id;primaryKey" json:"id"`
	AdmUserID   string `gorm:"column:adm_user_id" json:"adm_user_id"`
	AdmPermitID string `gorm:"column:adm_permit_id" json:"adm_permit_id"`
}

// TableName returns the database table name.
func (a *AdmUserPermit) TableName() string {
	return "adm_user_permits"
}
