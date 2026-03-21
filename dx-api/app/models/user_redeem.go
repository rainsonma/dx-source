package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type UserRedeem struct {
	orm.Timestamps
	ID         string           `gorm:"column:id;primaryKey" json:"id"`
	Code       string           `gorm:"column:code" json:"code"`
	Grade      string           `gorm:"column:grade" json:"grade"`
	UserID     *string          `gorm:"column:user_id" json:"user_id"`
	RedeemedAt *carbon.DateTime `gorm:"column:redeemed_at" json:"redeemed_at"`
}

// TableName returns the database table name.
func (u *UserRedeem) TableName() string {
	return "user_redeems"
}
