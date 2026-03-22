package consts

// Difficulty level values.
const (
	DifficultyA1A2 = "a1-a2"
	DifficultyB1B2 = "b1-b2"
	DifficultyC1C2 = "c1-c2"
)

// DifficultyOption represents a selectable difficulty level.
type DifficultyOption struct {
	Value string
	Label string
}

// DifficultyOptions returns all difficulty levels as an ordered slice.
func DifficultyOptions() []DifficultyOption {
	return []DifficultyOption{
		{Value: DifficultyA1A2, Label: "初级 (A1-A2)"},
		{Value: DifficultyB1B2, Label: "中级 (B1-B2)"},
		{Value: DifficultyC1C2, Label: "高级 (C1-C2)"},
	}
}
