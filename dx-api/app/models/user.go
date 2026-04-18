package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type User struct {
	orm.Timestamps
	ID                string           `gorm:"column:id;primaryKey" json:"id"`
	Grade             string           `gorm:"column:grade" json:"grade"`
	Username          string           `gorm:"column:username" json:"username"`
	Nickname          *string          `gorm:"column:nickname" json:"nickname"`
	Email             *string          `gorm:"column:email" json:"email"`
	Phone             *string          `gorm:"column:phone" json:"phone"`
	Password          string           `gorm:"column:password" json:"-"`
	AvatarID          *string          `gorm:"column:avatar_id" json:"avatar_id"`
	City              *string          `gorm:"column:city" json:"city"`
	Introduction      *string          `gorm:"column:introduction" json:"introduction"`
	IsActive          bool             `gorm:"column:is_active" json:"is_active"`
	IsMock            bool             `gorm:"column:is_mock" json:"is_mock"`
	Beans             int              `gorm:"column:beans" json:"beans"`
	GrantedBeans      int              `gorm:"column:granted_beans" json:"granted_beans"`
	Exp               int              `gorm:"column:exp" json:"exp"`
	InviteCode        string           `gorm:"column:invite_code" json:"invite_code"`
	CurrentPlayStreak int              `gorm:"column:current_play_streak" json:"current_play_streak"`
	MaxPlayStreak     int              `gorm:"column:max_play_streak" json:"max_play_streak"`
	LastPlayedAt      *carbon.DateTime `gorm:"column:last_played_at" json:"last_played_at"`
	VipDueAt          *carbon.DateTime `gorm:"column:vip_due_at" json:"vip_due_at"`
	LastReadNoticeAt  *carbon.DateTime `gorm:"column:last_read_notice_at" json:"last_read_notice_at"`
	OpenID            *string          `gorm:"column:openid"  json:"-"`
	UnionID           *string          `gorm:"column:unionid" json:"-"`
}

// TableName returns the database table name.
func (u *User) TableName() string {
	return "users"
}
