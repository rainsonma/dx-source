package helpers

import (
	"testing"
)

func TestProcessAnswer_Correct(t *testing.T) {
	state := ComboState{}
	result := ProcessAnswer(state, true)

	if result.PointsEarned != 1 {
		t.Errorf("expected 1 point, got %d", result.PointsEarned)
	}
	if result.ComboBonus != 0 {
		t.Errorf("expected 0 combo bonus, got %d", result.ComboBonus)
	}
	if result.State.Streak != 1 {
		t.Errorf("expected streak 1, got %d", result.State.Streak)
	}
	if result.State.CyclePosition != 1 {
		t.Errorf("expected cycle position 1, got %d", result.State.CyclePosition)
	}
	if result.State.TotalScore != 1 {
		t.Errorf("expected total score 1, got %d", result.State.TotalScore)
	}
	if result.State.MaxCombo != 1 {
		t.Errorf("expected max combo 1, got %d", result.State.MaxCombo)
	}
}

func TestProcessAnswer_Incorrect(t *testing.T) {
	state := ComboState{Streak: 5, CyclePosition: 5, TotalScore: 10, MaxCombo: 5}
	result := ProcessAnswer(state, false)

	if result.PointsEarned != 0 {
		t.Errorf("expected 0 points, got %d", result.PointsEarned)
	}
	if result.State.Streak != 0 {
		t.Errorf("expected streak reset to 0, got %d", result.State.Streak)
	}
	if result.State.CyclePosition != 0 {
		t.Errorf("expected cycle position reset to 0, got %d", result.State.CyclePosition)
	}
	if result.State.TotalScore != 10 {
		t.Errorf("expected total score preserved at 10, got %d", result.State.TotalScore)
	}
	if result.State.MaxCombo != 5 {
		t.Errorf("expected max combo preserved at 5, got %d", result.State.MaxCombo)
	}
}

func TestProcessAnswer_ComboThresholds(t *testing.T) {
	tests := []struct {
		name          string
		streak        int
		cyclePos      int
		wantBonus     int
		wantPoints    int
		wantCyclePos  int
	}{
		{"streak 1 — no bonus", 0, 0, 0, 1, 1},
		{"streak 2 — no bonus", 1, 1, 0, 1, 2},
		{"streak 3 — combo 3 bonus", 2, 2, 3, 4, 3},
		{"streak 4 — no bonus", 3, 3, 0, 1, 4},
		{"streak 5 — combo 5 bonus", 4, 4, 5, 6, 5},
		{"streak 6 — no bonus", 5, 5, 0, 1, 6},
		{"streak 10 — combo 10 bonus, cycle resets", 9, 9, 10, 11, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := ComboState{
				Streak:        tt.streak,
				CyclePosition: tt.cyclePos,
				TotalScore:    0,
				MaxCombo:      tt.streak,
			}
			result := ProcessAnswer(state, true)

			if result.ComboBonus != tt.wantBonus {
				t.Errorf("combo bonus: got %d, want %d", result.ComboBonus, tt.wantBonus)
			}
			if result.PointsEarned != tt.wantPoints {
				t.Errorf("points earned: got %d, want %d", result.PointsEarned, tt.wantPoints)
			}
			if result.State.CyclePosition != tt.wantCyclePos {
				t.Errorf("cycle position: got %d, want %d", result.State.CyclePosition, tt.wantCyclePos)
			}
		})
	}
}

func TestProcessAnswer_FullCycleScore(t *testing.T) {
	// Simulate 10 correct answers in a row — full cycle
	state := ComboState{}
	totalPoints := 0

	for i := 0; i < 10; i++ {
		result := ProcessAnswer(state, true)
		state = result.State
		totalPoints += result.PointsEarned
	}

	// 10 base points + 3 (at streak 3) + 5 (at streak 5) + 10 (at streak 10) = 28
	expectedTotal := 10 + 3 + 5 + 10
	if totalPoints != expectedTotal {
		t.Errorf("full cycle total: got %d, want %d", totalPoints, expectedTotal)
	}
	if state.TotalScore != expectedTotal {
		t.Errorf("state total score: got %d, want %d", state.TotalScore, expectedTotal)
	}
	if state.MaxCombo != 10 {
		t.Errorf("max combo: got %d, want 10", state.MaxCombo)
	}
	if state.CyclePosition != 0 {
		t.Errorf("cycle position should reset to 0 after 10, got %d", state.CyclePosition)
	}
	if state.Streak != 10 {
		t.Errorf("streak should be 10, got %d", state.Streak)
	}
}

func TestProcessAnswer_MaxComboPreserved(t *testing.T) {
	// Build a streak of 5, break it, build to 3 — max should stay 5
	state := ComboState{}
	for i := 0; i < 5; i++ {
		result := ProcessAnswer(state, true)
		state = result.State
	}
	if state.MaxCombo != 5 {
		t.Fatalf("expected max combo 5 after 5 correct, got %d", state.MaxCombo)
	}

	// Wrong answer breaks streak
	result := ProcessAnswer(state, false)
	state = result.State
	if state.MaxCombo != 5 {
		t.Errorf("max combo should stay 5 after wrong answer, got %d", state.MaxCombo)
	}

	// 3 more correct — max combo should stay 5
	for i := 0; i < 3; i++ {
		result := ProcessAnswer(state, true)
		state = result.State
	}
	if state.MaxCombo != 5 {
		t.Errorf("max combo should stay 5, got %d", state.MaxCombo)
	}
}

func TestProcessAnswer_SecondCycleGivesBonusesAgain(t *testing.T) {
	// First 10 correct answers, then 3 more — the 3rd in cycle 2 should get a bonus
	state := ComboState{}
	for i := 0; i < 10; i++ {
		result := ProcessAnswer(state, true)
		state = result.State
	}

	// Cycle 2: answers 11, 12, 13
	for i := 0; i < 2; i++ {
		result := ProcessAnswer(state, true)
		state = result.State
	}

	// 13th answer (cycle position 3) should get combo 3 bonus
	result := ProcessAnswer(state, true)
	if result.ComboBonus != 3 {
		t.Errorf("expected combo 3 bonus at cycle position 3 in second cycle, got %d", result.ComboBonus)
	}
	if result.PointsEarned != 4 {
		t.Errorf("expected 4 points (1+3), got %d", result.PointsEarned)
	}
}
