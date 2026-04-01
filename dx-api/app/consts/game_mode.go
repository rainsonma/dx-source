package consts

// Game mode values.
const (
	GameModeWordSentence       = "word-sentence"
	GameModeVocabBattle        = "vocab-battle"
	GameModeVocabMatch         = "vocab-match"
	GameModeVocabElimination   = "vocab-elimination"
	GameModeListeningChallenge = "listening-challenge"
)

// GameModeLabels maps each mode to its Chinese label.
var GameModeLabels = map[string]string{
	GameModeWordSentence:       "连词成句",
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
		{Value: GameModeWordSentence, Label: "连词成句"},
		{Value: GameModeVocabBattle, Label: "词汇对轰"},
		{Value: GameModeVocabMatch, Label: "词汇配对"},
		{Value: GameModeVocabElimination, Label: "消消乐"},
		{Value: GameModeListeningChallenge, Label: "听力闯关"},
	}
}
