package models

import "github.com/goravel/framework/database/orm"

type AdmLogin struct {
	orm.Timestamps
	ID        string  `gorm:"column:id;primaryKey" json:"id"`
	AdmUserID string  `gorm:"column:adm_user_id" json:"adm_user_id"`
	Ip        string  `gorm:"column:ip" json:"ip"`
	Agent     *string `gorm:"column:agent" json:"agent"`
	Country   *string `gorm:"column:country" json:"country"`
	Province  *string `gorm:"column:province" json:"province"`
	City      *string `gorm:"column:city" json:"city"`
	Isp       *string `gorm:"column:isp" json:"isp"`
}

// TableName returns the database table name.
func (a *AdmLogin) TableName() string {
	return "adm_logins"
}
