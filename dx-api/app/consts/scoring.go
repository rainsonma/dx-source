package consts

// Scoring consts for game score calculation.
const (
	CorrectAnswer        = 1
	Combo3Bonus          = 3
	Combo5Bonus          = 5
	Combo10Bonus         = 10
	ComboCycleLength     = 10
	LevelCompleteExp     = 10
	ExpAccuracyThreshold = 0.6
)

// ComboThreshold defines the streak count and its associated bonus.
type ComboThreshold struct {
	Streak int
	Bonus  int
}

// ComboThresholds lists the combo bonus tiers in ascending order.
var ComboThresholds = []ComboThreshold{
	{Streak: 3, Bonus: Combo3Bonus},
	{Streak: 5, Bonus: Combo5Bonus},
	{Streak: 10, Bonus: Combo10Bonus},
}
