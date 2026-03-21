package models

import "github.com/goravel/framework/database/orm"

type UserLogin struct {
	orm.Timestamps
	ID       string  `gorm:"column:id;primaryKey" json:"id"`
	UserID   string  `gorm:"column:user_id" json:"user_id"`
	IP       string  `gorm:"column:ip" json:"ip"`
	Agent    *string `gorm:"column:agent" json:"agent"`
	Country  *string `gorm:"column:country" json:"country"`
	Province *string `gorm:"column:province" json:"province"`
	City     *string `gorm:"column:city" json:"city"`
	ISP      *string `gorm:"column:isp" json:"isp"`
}

// TableName returns the database table name.
func (u *UserLogin) TableName() string {
	return "user_logins"
}
