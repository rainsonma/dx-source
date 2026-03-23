package api

import "github.com/goravel/framework/contracts/http"

type RedeemCodeRequest struct {
	Code string `form:"code" json:"code"`
}

func (r *RedeemCodeRequest) Authorize(ctx http.Context) error { return nil }
func (r *RedeemCodeRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"code": "required|len:19",
	}
}
func (r *RedeemCodeRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"code": "trim|upper",
	}
}
func (r *RedeemCodeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.required": "请输入兑换码",
		"code.len":      "兑换码格式不正确",
	}
}
