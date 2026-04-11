package consts

// Referral status values.
const (
	ReferralStatusPending  = "pending"
	ReferralStatusPaid     = "paid"
	ReferralStatusRewarded = "rewarded"
)

// ReferralStatusLabels maps each referral status to its Chinese label.
var ReferralStatusLabels = map[string]string{
	ReferralStatusPending:  "待激活",
	ReferralStatusPaid:     "已付费",
	ReferralStatusRewarded: "已发放",
}
