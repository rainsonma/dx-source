package consts

const (
	PkDifficultyEasy   = "easy"
	PkDifficultyNormal = "normal"
	PkDifficultyHard   = "hard"
)

// PkDifficultyParams defines the robot behavior for a difficulty level.
type PkDifficultyParams struct {
	AccuracyMin   float64
	AccuracyMax   float64
	MinDelayMs    int
	MaxDelayMs    int
	ComboBreakPct float64
}

// PkDifficulties maps difficulty slugs to their robot behavior parameters.
var PkDifficulties = map[string]PkDifficultyParams{
	PkDifficultyEasy:   {AccuracyMin: 0.50, AccuracyMax: 0.70, MinDelayMs: 7000, MaxDelayMs: 12000, ComboBreakPct: 0.50},
	PkDifficultyNormal: {AccuracyMin: 0.70, AccuracyMax: 0.85, MinDelayMs: 5000, MaxDelayMs: 8000, ComboBreakPct: 0.30},
	PkDifficultyHard:   {AccuracyMin: 0.85, AccuracyMax: 0.95, MinDelayMs: 3000, MaxDelayMs: 5000, ComboBreakPct: 0.10},
}

// PkDifficultySlugs lists valid difficulty slugs for validation.
var PkDifficultySlugs = []string{PkDifficultyEasy, PkDifficultyNormal, PkDifficultyHard}
