package constants

// Game pattern values.
const (
	GamePatternListen = "listen"
	GamePatternSpeak  = "speak"
	GamePatternRead   = "read"
	GamePatternWrite  = "write"
)

// DefaultGamePattern is the default game pattern.
const DefaultGamePattern = GamePatternWrite

// GamePatternLabels maps each pattern to its Chinese label.
var GamePatternLabels = map[string]string{
	GamePatternListen: "听",
	GamePatternSpeak:  "说",
	GamePatternRead:   "读",
	GamePatternWrite:  "写",
}
