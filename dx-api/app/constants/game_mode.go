package constants

// Game mode values.
const (
	GameModeLSRW               = "lsrw"
	GameModeVocabBattle        = "vocab-battle"
	GameModeVocabMatch         = "vocab-match"
	GameModeVocabElimination   = "vocab-elimination"
	GameModeListeningChallenge = "listening-challenge"
)

// GameModeLabels maps each mode to its Chinese label.
var GameModeLabels = map[string]string{
	GameModeLSRW:               "听说读写",
	GameModeVocabBattle:        "词汇对轰",
	GameModeVocabMatch:         "词汇配对",
	GameModeVocabElimination:   "消消乐",
	GameModeListeningChallenge: "听力闯关",
}

// GameModeOption represents a selectable game mode.
type GameModeOption struct {
	Value string
	Label string
}

// GameModeOptions returns all game modes as an ordered slice.
func GameModeOptions() []GameModeOption {
	return []GameModeOption{
		{Value: GameModeLSRW, Label: "听说读写"},
		{Value: GameModeVocabBattle, Label: "词汇对轰"},
		{Value: GameModeVocabMatch, Label: "词汇配对"},
		{Value: GameModeVocabElimination, Label: "消消乐"},
		{Value: GameModeListeningChallenge, Label: "听力闯关"},
	}
}
