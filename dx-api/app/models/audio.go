package models

import "github.com/goravel/framework/database/orm"

type Audio struct {
	orm.Timestamps
	ID        string  `gorm:"column:id;primaryKey" json:"id"`
	AdmUserID *string `gorm:"column:adm_user_id" json:"adm_user_id"`
	UserID    *string `gorm:"column:user_id" json:"user_id"`
	Url       string  `gorm:"column:url" json:"url"`
	Name      string  `gorm:"column:name" json:"name"`
	Mime      string  `gorm:"column:mime" json:"mime"`
	Size      int     `gorm:"column:size" json:"size"`
	Duration  int     `gorm:"column:duration" json:"duration"`
	Role      string  `gorm:"column:role" json:"role"`
}

// TableName returns the database table name.
func (a *Audio) TableName() string {
	return "audios"
}
