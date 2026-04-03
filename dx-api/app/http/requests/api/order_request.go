package api

import (
	"dx-api/app/helpers"

	"github.com/goravel/framework/contracts/http"
)

type CreateMembershipOrderRequest struct {
	Grade         string `form:"grade" json:"grade"`
	PaymentMethod string `form:"paymentMethod" json:"paymentMethod"`
}

func (r *CreateMembershipOrderRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateMembershipOrderRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"grade":         "required|" + helpers.InEnum("grade"),
		"paymentMethod": "required|" + helpers.InEnum("payment_method"),
	}
}
func (r *CreateMembershipOrderRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"grade":         "trim",
		"paymentMethod": "trim",
	}
}
func (r *CreateMembershipOrderRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"grade.required":         "请选择会员等级",
		"grade.in":               "无效的会员等级",
		"paymentMethod.required": "请选择支付方式",
		"paymentMethod.in":       "无效的支付方式",
	}
}

type CreateBeansOrderRequest struct {
	Package       string `form:"package" json:"package"`
	PaymentMethod string `form:"paymentMethod" json:"paymentMethod"`
}

func (r *CreateBeansOrderRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateBeansOrderRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"package":       "required|" + helpers.InEnum("bean_package"),
		"paymentMethod": "required|" + helpers.InEnum("payment_method"),
	}
}
func (r *CreateBeansOrderRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"package":       "trim",
		"paymentMethod": "trim",
	}
}
func (r *CreateBeansOrderRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"package.required":       "请选择能量豆套餐",
		"package.in":             "无效的能量豆套餐",
		"paymentMethod.required": "请选择支付方式",
		"paymentMethod.in":       "无效的支付方式",
	}
}
