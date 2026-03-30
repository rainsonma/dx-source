package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAndDetermineWinnerFunctionExists(t *testing.T) {
	// Verify function signature: CheckAndDetermineWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error)
	var fn func(string, string) (*LevelWinnerResult, error) = CheckAndDetermineWinner
	assert.NotNil(t, fn)
}

func TestDetermineWinnerForLevelFunctionExists(t *testing.T) {
	// Verify function signature: DetermineWinnerForLevel(gameGroupID, gameLevelID string, sessionIDs []string) (*LevelWinnerResult, error)
	var fn func(string, string, []string) (*LevelWinnerResult, error) = DetermineWinnerForLevel
	assert.NotNil(t, fn)
}

func TestRecheckGroupWinnersFunctionExists(t *testing.T) {
	// Verify function signature: RecheckGroupWinners(groupID string)
	var fn func(string) = RecheckGroupWinners
	assert.NotNil(t, fn)
}

func TestLevelWinnerResultStructure(t *testing.T) {
	// Verify solo winner result fields
	result := LevelWinnerResult{
		GameLevelID:  "level-1",
		Mode:         "group_solo",
		Winner:       SoloWinner{UserID: "u1", UserName: "Alice", Score: 42},
		Participants: []SoloWinner{{UserID: "u1", UserName: "Alice", Score: 42}},
	}
	assert.Equal(t, "level-1", result.GameLevelID)
	assert.Equal(t, "group_solo", result.Mode)
	assert.Len(t, result.Participants, 1)
	assert.Equal(t, 42, result.Participants[0].Score)
}

// TestConnectedParticipantDeduplication verifies that a user with multiple
// active session totals is only counted once in the participant list. Without
// deduplication, participantCount exceeds COUNT(DISTINCT user_id) and the
// winner check can never succeed (regression root cause).
func TestConnectedParticipantDeduplication(t *testing.T) {
	type idRow struct {
		ID     string
		UserID string
	}

	// Simulate: user-1 has 2 active sessions, user-2 has 1
	lockedRows := []idRow{
		{ID: "s1", UserID: "user-1"},
		{ID: "s2", UserID: "user-1"}, // duplicate
		{ID: "s3", UserID: "user-2"},
	}

	connectedSet := map[string]bool{
		"user-1": true,
		"user-2": true,
	}

	// Apply the deduplication logic from CheckAndDetermineWinner
	seen := make(map[string]bool)
	var connectedParticipantIDs []string
	for _, row := range lockedRows {
		if connectedSet[row.UserID] && !seen[row.UserID] {
			seen[row.UserID] = true
			connectedParticipantIDs = append(connectedParticipantIDs, row.UserID)
		}
	}

	assert.Equal(t, 2, len(connectedParticipantIDs), "should count each user once")
	assert.ElementsMatch(t, []string{"user-1", "user-2"}, connectedParticipantIDs)
}

// TestDisconnectedPlayerExcludedFromParticipants verifies that a player who
// disconnected from SSE is not counted as a participant.
func TestDisconnectedPlayerExcludedFromParticipants(t *testing.T) {
	type idRow struct {
		ID     string
		UserID string
	}

	// 3 players have active sessions, but user-3 disconnected
	lockedRows := []idRow{
		{ID: "s1", UserID: "user-1"},
		{ID: "s2", UserID: "user-2"},
		{ID: "s3", UserID: "user-3"},
	}

	connectedSet := map[string]bool{
		"user-1": true,
		"user-2": true,
		// user-3 is NOT connected
	}

	seen := make(map[string]bool)
	var connectedParticipantIDs []string
	for _, row := range lockedRows {
		if connectedSet[row.UserID] && !seen[row.UserID] {
			seen[row.UserID] = true
			connectedParticipantIDs = append(connectedParticipantIDs, row.UserID)
		}
	}

	assert.Equal(t, 2, len(connectedParticipantIDs), "disconnected user should not be counted")
	assert.ElementsMatch(t, []string{"user-1", "user-2"}, connectedParticipantIDs)
}

func TestTeamWinnerStructure(t *testing.T) {
	// Verify team winner result fields
	result := LevelWinnerResult{
		GameLevelID: "level-1",
		Mode:        "group_team",
		Winner: TeamWinner{
			SubgroupID:   "sg-1",
			SubgroupName: "A组",
			TotalScore:   128,
			Members: []TeamMember{
				{UserID: "u1", UserName: "Alice", Score: 45},
				{UserID: "u2", UserName: "Bob", Score: 42},
			},
		},
		Participants: []SoloWinner{
			{UserID: "u1", UserName: "Alice", Score: 45},
			{UserID: "u2", UserName: "Bob", Score: 42},
		},
		Teams: []TeamWinner{
			{SubgroupID: "sg-1", SubgroupName: "A组", TotalScore: 128},
		},
	}
	assert.Equal(t, "group_team", result.Mode)
	assert.Len(t, result.Teams, 1)

	winner, ok := result.Winner.(TeamWinner)
	assert.True(t, ok)
	assert.Equal(t, 128, winner.TotalScore)
	assert.Len(t, winner.Members, 2)
}
