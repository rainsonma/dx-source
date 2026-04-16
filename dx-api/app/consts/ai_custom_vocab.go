package consts

// Vocab mode limits.
const (
	MaxMetasPerLevel      = 20
	MaxLevelsPerGame      = 20
	VocabMatchCount       = 5
	VocabEliminationCount = 8
	VocabBattleCount      = 20
)

// VocabGenerateCount returns how many pairs to generate based on game mode.
func VocabGenerateCount(mode string) int {
	switch mode {
	case GameModeVocabMatch:
		return VocabMatchCount
	case GameModeVocabElimination:
		return VocabEliminationCount
	case GameModeVocabBattle:
		return VocabBattleCount
	default:
		return VocabMatchCount
	}
}

// IsVocabMode returns true if the mode is one of the three vocab game modes.
func IsVocabMode(mode string) bool {
	return mode == GameModeVocabBattle || mode == GameModeVocabMatch || mode == GameModeVocabElimination
}

// VocabBatchSize returns the required batch size for the given vocab mode.
// Returns 0 for modes with no batch constraint (vocab-battle).
func VocabBatchSize(mode string) int {
	switch mode {
	case GameModeVocabMatch:
		return VocabMatchCount
	case GameModeVocabElimination:
		return VocabEliminationCount
	default:
		return 0
	}
}
