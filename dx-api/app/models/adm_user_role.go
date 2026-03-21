package models

import "github.com/goravel/framework/database/orm"

type AdmUserRole struct {
	orm.Timestamps
	ID        string `gorm:"column:id;primaryKey" json:"id"`
	AdmRoleID string `gorm:"column:adm_role_id" json:"adm_role_id"`
	AdmUserID string `gorm:"column:adm_user_id" json:"adm_user_id"`
}

// TableName returns the database table name.
func (a *AdmUserRole) TableName() string {
	return "adm_user_roles"
}
