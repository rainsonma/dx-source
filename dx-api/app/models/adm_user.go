package models

import "github.com/goravel/framework/database/orm"

type AdmUser struct {
	orm.Timestamps
	ID       string  `gorm:"column:id;primaryKey" json:"id"`
	Username string  `gorm:"column:username" json:"username"`
	Nickname *string `gorm:"column:nickname" json:"nickname"`
	Password string  `gorm:"column:password" json:"-"`
	AvatarID *string `gorm:"column:avatar_id" json:"avatar_id"`
	IsActive bool    `gorm:"column:is_active" json:"is_active"`
}

// TableName returns the database table name.
func (a *AdmUser) TableName() string {
	return "adm_users"
}
