package api

// RedeemCodeRequest validates a redeem code submission.
type RedeemCodeRequest struct {
	Code string `form:"code" json:"code"`
}
