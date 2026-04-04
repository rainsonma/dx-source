package consts

import (
	"testing"
)

func TestGetLevel(t *testing.T) {
	tests := []struct {
		name  string
		exp   int
		want  int
		isErr bool
	}{
		{"new user 0 exp", 0, 0, false},
		{"just below Lv.1", 99, 0, false},
		{"exactly Lv.1", 100, 1, false},
		{"just below Lv.2", 199, 1, false},
		{"exactly Lv.2", 200, 2, false},
		{"just below Lv.3", 304, 2, false},
		{"exactly Lv.3", 305, 3, false},
		{"negative exp", -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLevel(tt.exp)
			if tt.isErr {
				if err == nil {
					t.Fatalf("GetLevel(%d) expected error, got level %d", tt.exp, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetLevel(%d) unexpected error: %v", tt.exp, err)
			}
			if got != tt.want {
				t.Errorf("GetLevel(%d) = %d, want %d", tt.exp, got, tt.want)
			}
		})
	}
}

func TestGetExpForLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  int
		isErr bool
	}{
		{"level 0", 0, 0, false},
		{"level 1", 1, 100, false},
		{"level 2", 2, 200, false},
		{"level 3", 3, 305, false},
		{"max level", MaxLevel, 248531, false},
		{"below min", -1, 0, true},
		{"above max", MaxLevel + 1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetExpForLevel(tt.level)
			if tt.isErr {
				if err == nil {
					t.Fatalf("GetExpForLevel(%d) expected error, got %d", tt.level, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetExpForLevel(%d) unexpected error: %v", tt.level, err)
			}
			if got != tt.want {
				t.Errorf("GetExpForLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}

func TestLevelTableBoundaries(t *testing.T) {
	// Table has 101 entries (Lv.0 through Lv.100)
	if len(userLevels) != MaxLevel+1 {
		t.Errorf("userLevels length = %d, want %d", len(userLevels), MaxLevel+1)
	}

	// First entry is Lv.0 at 0 EXP
	if userLevels[0].Level != 0 || userLevels[0].ExpRequired != 0 {
		t.Errorf("userLevels[0] = %+v, want {Level:0 ExpRequired:0}", userLevels[0])
	}

	// Second entry is Lv.1 at 100 EXP
	if userLevels[1].Level != 1 || userLevels[1].ExpRequired != 100 {
		t.Errorf("userLevels[1] = %+v, want {Level:1 ExpRequired:100}", userLevels[1])
	}

	// EXP is strictly increasing
	for i := 1; i < len(userLevels); i++ {
		if userLevels[i].ExpRequired <= userLevels[i-1].ExpRequired {
			t.Errorf("EXP not increasing at level %d: %d <= %d",
				userLevels[i].Level, userLevels[i].ExpRequired, userLevels[i-1].ExpRequired)
		}
	}
}
