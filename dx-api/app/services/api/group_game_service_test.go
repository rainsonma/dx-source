package api

import "testing"

func TestSearchGamesForGroupExists(t *testing.T) {
	// Verify function signature: SearchGamesForGroup(query string, limit int) ([]GroupGameSearchItem, error)
	var fn func(string, int) ([]GroupGameSearchItem, error) = SearchGamesForGroup
	_ = fn
}

func TestSetGroupGameExists(t *testing.T) {
	// Verify function signature: SetGroupGame(userID, groupID, gameID, gameMode string) error
	var fn func(string, string, string, string) error = SetGroupGame
	_ = fn
}

func TestClearGroupGameExists(t *testing.T) {
	// Verify function signature: ClearGroupGame(userID, groupID string) error
	var fn func(string, string) error = ClearGroupGame
	_ = fn
}
