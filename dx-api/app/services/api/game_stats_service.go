package api

import (
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// GameStatsData represents per-game stats for a user.
type GameStatsData struct {
	HighestScore    int  `json:"highestScore"`
	TotalSessions   int  `json:"totalSessions"`
	TotalScores     int  `json:"totalScores"`
	TotalExp        int  `json:"totalExp"`
	TotalPlayTime   int  `json:"totalPlayTime"`
	CompletionCount int  `json:"completionCount"`
	FirstCompleted  bool `json:"firstCompleted"`
}

// GetGameStats returns the user's stats for a specific game, or nil if none.
func GetGameStats(userID, gameID string) (*GameStatsData, error) {
	var stats models.GameStatsTotal
	if err := facades.Orm().Query().
		Where("user_id", userID).Where("game_id", gameID).
		First(&stats); err != nil || stats.ID == "" {
		return nil, nil
	}

	return &GameStatsData{
		HighestScore:    stats.HighestScore,
		TotalSessions:   stats.TotalSessions,
		TotalScores:     stats.TotalScores,
		TotalExp:        stats.TotalExp,
		TotalPlayTime:   stats.TotalPlayTime,
		CompletionCount: stats.CompletionCount,
		FirstCompleted:  stats.FirstCompletedAt != nil,
	}, nil
}
