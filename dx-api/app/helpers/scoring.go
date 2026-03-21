package helpers

import "dx-api/app/constants"

// ComboState tracks the player's combo streak within a level.
type ComboState struct {
	Streak        int
	CyclePosition int
	TotalScore    int
	MaxCombo      int
}

// ProcessAnswerResult holds the scoring outcome of a single answer.
type ProcessAnswerResult struct {
	State       ComboState
	PointsEarned int
	ComboBonus   int
}

// ProcessAnswer computes scoring for a single answer based on current combo state.
func ProcessAnswer(state ComboState, isCorrect bool) ProcessAnswerResult {
	if !isCorrect {
		return ProcessAnswerResult{
			State: ComboState{
				Streak:        0,
				CyclePosition: 0,
				TotalScore:    state.TotalScore,
				MaxCombo:      state.MaxCombo,
			},
		}
	}

	points := constants.CorrectAnswer
	bonus := 0
	newStreak := state.Streak + 1
	newCyclePosition := state.CyclePosition + 1

	for _, threshold := range constants.ComboThresholds {
		if newCyclePosition == threshold.Streak {
			bonus += threshold.Bonus
		}
	}

	if newCyclePosition >= constants.ComboCycleLength {
		newCyclePosition = 0
	}

	points += bonus

	newMaxCombo := state.MaxCombo
	if newStreak > newMaxCombo {
		newMaxCombo = newStreak
	}

	return ProcessAnswerResult{
		State: ComboState{
			Streak:        newStreak,
			CyclePosition: newCyclePosition,
			TotalScore:    state.TotalScore + points,
			MaxCombo:      newMaxCombo,
		},
		PointsEarned: points,
		ComboBonus:   bonus,
	}
}
