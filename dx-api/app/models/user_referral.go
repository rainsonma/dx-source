package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type UserReferral struct {
	orm.Timestamps
	ID           string           `gorm:"column:id;primaryKey" json:"id"`
	ReferrerID   string           `gorm:"column:referrer_id" json:"referrer_id"`
	InviteeID    *string          `gorm:"column:invitee_id" json:"invitee_id"`
	Status       string           `gorm:"column:status" json:"status"`
	RewardAmount float64          `gorm:"column:reward_amount" json:"reward_amount"`
	RewardedAt   *carbon.DateTime `gorm:"column:rewarded_at" json:"rewarded_at"`
}

// TableName returns the database table name.
func (u *UserReferral) TableName() string {
	return "user_referrals"
}
