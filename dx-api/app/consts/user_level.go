package consts

import (
	"fmt"
	"math"
)

// Level progression consts.
const (
	MaxLevel = 100
	// introExp is the flat EXP cost for Lv.0 → Lv.1, separate from the exponential curve.
	introExp   = 100
	baseExp    = 100
	multiplier = 1.05
)

// UserLevel holds a level number and the cumulative exp required to reach it.
type UserLevel struct {
	Level       int
	ExpRequired int
}

// userLevels is the precomputed level table.
var userLevels []UserLevel

func init() {
	userLevels = generateLevels()
}

// generateLevels builds the full level progression table.
func generateLevels() []UserLevel {
	levels := make([]UserLevel, 0, MaxLevel+1)
	levels = append(levels, UserLevel{Level: 0, ExpRequired: 0})
	levels = append(levels, UserLevel{Level: 1, ExpRequired: introExp})

	cumulative := introExp
	for i := 2; i <= MaxLevel; i++ {
		cumulative += int(math.Floor(baseExp * math.Pow(multiplier, float64(i-2))))
		levels = append(levels, UserLevel{Level: i, ExpRequired: cumulative})
	}

	return levels
}

// GetLevel returns the level for the given cumulative exp.
func GetLevel(exp int) (int, error) {
	if exp < 0 {
		return 0, fmt.Errorf("failed to get level: exp must be non-negative, got %d", exp)
	}
	for i := len(userLevels) - 1; i >= 0; i-- {
		if exp >= userLevels[i].ExpRequired {
			return userLevels[i].Level, nil
		}
	}
	// Unreachable: userLevels[0].ExpRequired is always 0 and exp is non-negative.
	return 0, nil
}

// GetExpForLevel returns the cumulative exp required to reach the given level.
func GetExpForLevel(level int) (int, error) {
	if level < 0 || level > MaxLevel {
		return 0, fmt.Errorf("failed to get exp for level: level must be between 0 and %d, got %d", MaxLevel, level)
	}
	return userLevels[level].ExpRequired, nil
}
