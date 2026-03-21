package models

import "github.com/goravel/framework/database/orm"

type AdmRolePermit struct {
	orm.Timestamps
	ID          string `gorm:"column:id;primaryKey" json:"id"`
	AdmRoleID   string `gorm:"column:adm_role_id" json:"adm_role_id"`
	AdmPermitID string `gorm:"column:adm_permit_id" json:"adm_permit_id"`
}

// TableName returns the database table name.
func (a *AdmRolePermit) TableName() string {
	return "adm_role_permits"
}
