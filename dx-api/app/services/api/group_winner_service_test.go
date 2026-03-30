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
