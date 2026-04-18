package api

import "github.com/goravel/framework/contracts/http"

type WechatMiniAuthRequest struct {
	Code string `form:"code" json:"code"`
}

func (r *WechatMiniAuthRequest) Authorize(ctx http.Context) error { return nil }
func (r *WechatMiniAuthRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"code": "required",
	}
}
func (r *WechatMiniAuthRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{"code": "trim"}
}
func (r *WechatMiniAuthRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.required": "缺少 wx.login code",
	}
}
