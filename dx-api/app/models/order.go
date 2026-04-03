package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type Order struct {
	orm.Timestamps
	ID            string           `gorm:"column:id;primaryKey" json:"id"`
	UserID        string           `gorm:"column:user_id" json:"user_id"`
	Type          string           `gorm:"column:type" json:"type"`
	Product       string           `gorm:"column:product" json:"product"`
	Amount        int              `gorm:"column:amount" json:"amount"`
	Status        string           `gorm:"column:status" json:"status"`
	PaymentMethod *string          `gorm:"column:payment_method" json:"payment_method"`
	PaymentNo     *string          `gorm:"column:payment_no" json:"payment_no"`
	PaidAt        *carbon.DateTime `gorm:"column:paid_at" json:"paid_at"`
	FulfilledAt   *carbon.DateTime `gorm:"column:fulfilled_at" json:"fulfilled_at"`
	ExpiresAt     carbon.DateTime  `gorm:"column:expires_at" json:"expires_at"`
}

func (o *Order) TableName() string {
	return "orders"
}
