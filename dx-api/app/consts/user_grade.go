package consts

// User grade values.
const (
	UserGradeFree     = "free"
	UserGradeMonth    = "month"
	UserGradeSeason   = "season"
	UserGradeYear     = "year"
	UserGradeLifetime = "lifetime"
)

// UserGradeLabels maps each grade to its Chinese label.
var UserGradeLabels = map[string]string{
	UserGradeFree:     "免费会员",
	UserGradeMonth:    "月度会员",
	UserGradeSeason:   "季度会员",
	UserGradeYear:     "年度会员",
	UserGradeLifetime: "终身会员",
}

// UserGradePrices maps each grade to its price in CNY.
var UserGradePrices = map[string]int{
	UserGradeFree:     0,
	UserGradeMonth:    39,
	UserGradeSeason:   99,
	UserGradeYear:     309,
	UserGradeLifetime: 1999,
}

// UserGradeMonths maps each grade to the number of months it grants.
// A value of 0 means not applicable (free has no months, lifetime is permanent).
var UserGradeMonths = map[string]int{
	UserGradeFree:     0,
	UserGradeMonth:    1,
	UserGradeSeason:   3,
	UserGradeYear:     12,
	UserGradeLifetime: 0,
}
