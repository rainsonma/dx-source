package models

import "github.com/goravel/framework/database/orm"

type UserLogin struct {
	orm.Timestamps
	ID       string  `gorm:"column:id;primaryKey" json:"id"`
	UserID   string  `gorm:"column:user_id" json:"user_id"`
	IP       string  `gorm:"column:ip" json:"ip"`
	Agent    *string `gorm:"column:agent" json:"agent"`
	Platform *string `gorm:"column:platform" json:"platform"`
	Country  *string `gorm:"column:country" json:"country"`
	Province *string `gorm:"column:province" json:"province"`
	City     *string `gorm:"column:city" json:"city"`
	ISP      *string `gorm:"column:isp" json:"isp"`
}

func (u *UserLogin) TableName() string {
	return "user_logins"
}
