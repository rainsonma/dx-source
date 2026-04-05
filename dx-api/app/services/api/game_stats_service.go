package api

import (
	"github.com/goravel/framework/facades"
)

// GameStatsData represents per-game stats for a user.
type GameStatsData struct {
	HighestScore    int    `json:"highestScore"`
	TotalSessions   int    `json:"totalSessions"`
	TotalScores     int    `json:"totalScores"`
	TotalExp        int    `json:"totalExp"`
	TotalPlayTime   int    `json:"totalPlayTime"`
	CompletionCount int    `json:"completionCount"`
	FirstCompleted  *int64 `json:"firstCompleted"`
}

// GetGameStats returns the user's stats for a specific game, derived from game_sessions.
func GetGameStats(userID, gameID string) (*GameStatsData, error) {
	var result struct {
		TotalSessions   int    `gorm:"column:total_sessions"`
		HighestScore    int    `gorm:"column:highest_score"`
		TotalScores     int    `gorm:"column:total_scores"`
		TotalExp        int    `gorm:"column:total_exp"`
		TotalPlayTime   int    `gorm:"column:total_play_time"`
		CompletionCount int    `gorm:"column:completion_count"`
		FirstCompleted  *int64 `gorm:"column:first_completed"`
	}

	err := facades.Orm().Query().Raw(
		`SELECT
			COUNT(*)::int AS total_sessions,
			COALESCE(MAX(score), 0)::int AS highest_score,
			COALESCE(SUM(score), 0)::int AS total_scores,
			COALESCE(SUM(exp), 0)::int AS total_exp,
			COALESCE(SUM(play_time), 0)::int AS total_play_time,
			COUNT(*)::int AS completion_count,
			EXTRACT(EPOCH FROM MIN(ended_at))::bigint AS first_completed
		FROM game_sessions
		WHERE user_id = ? AND game_id = ? AND ended_at IS NOT NULL`,
		userID, gameID,
	).Scan(&result)
	if err != nil {
		return nil, err
	}

	return &GameStatsData{
		HighestScore:    result.HighestScore,
		TotalSessions:   result.TotalSessions,
		TotalScores:     result.TotalScores,
		TotalExp:        result.TotalExp,
		TotalPlayTime:   result.TotalPlayTime,
		CompletionCount: result.CompletionCount,
		FirstCompleted:  result.FirstCompleted,
	}, nil
}
