package api

import (
	"fmt"

	"dx-api/app/helpers"

	"github.com/goravel/framework/facades"
)

// LeaderboardEntry represents a single user row in the leaderboard.
type LeaderboardEntry struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatarUrl"`
	Value     int     `json:"value"`
	Rank      int     `json:"rank"`
}

// LeaderboardResult contains the top entries and the current user's rank.
type LeaderboardResult struct {
	Entries []LeaderboardEntry `json:"entries"`
	MyRank  *LeaderboardEntry  `json:"myRank"`
}

// leaderboardRow is the raw scan target for leaderboard queries.
type leaderboardRow struct {
	ID       string  `gorm:"column:id"`
	Username string  `gorm:"column:username"`
	Nickname *string `gorm:"column:nickname"`
	AvatarID *string `gorm:"column:avatar_id"`
	Value    int     `gorm:"column:value"`
	Rank     int     `gorm:"column:rank"`
}

// GetLeaderboard returns a ranked list by type (exp|playtime) and period (day|week|month).
func GetLeaderboard(lbType, period, userID string) (*LeaderboardResult, error) {
	if lbType == "exp" {
		return getWindowedExp(period, userID)
	}
	return getWindowedPlayTime(period, userID)
}

// getWindowedExp ranks active users by exp earned within a time window from game_sessions.
func getWindowedExp(period, userID string) (*LeaderboardResult, error) {
	var rows []leaderboardRow
	if err := facades.Orm().Query().Raw(fmt.Sprintf(`
		WITH ranked AS (
			SELECT u.id, u.username, u.nickname, u.avatar_id,
			       COALESCE(SUM(s.exp), 0)::int AS value,
			       RANK() OVER (ORDER BY COALESCE(SUM(s.exp), 0) DESC)::int AS rank
			FROM users u
			INNER JOIN game_sessions s ON s.user_id = u.id
			  AND s.last_played_at >= %s
			  AND s.last_played_at < NOW()
			WHERE u.is_active = true
			GROUP BY u.id, u.username, u.nickname, u.avatar_id
			HAVING COALESCE(SUM(s.exp), 0) > 0
		)
		SELECT * FROM ranked WHERE rank <= 100 ORDER BY rank
	`, windowStartSQL(period))).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query windowed exp leaderboard: %w", err)
	}

	return buildLeaderboardResult(rows, userID), nil
}

// getWindowedPlayTime ranks active users by play time within a time window from game_sessions.
func getWindowedPlayTime(period, userID string) (*LeaderboardResult, error) {
	var rows []leaderboardRow
	if err := facades.Orm().Query().Raw(fmt.Sprintf(`
		WITH ranked AS (
			SELECT u.id, u.username, u.nickname, u.avatar_id,
			       COALESCE(SUM(s.play_time), 0)::int AS value,
			       RANK() OVER (ORDER BY COALESCE(SUM(s.play_time), 0) DESC)::int AS rank
			FROM users u
			INNER JOIN game_sessions s ON s.user_id = u.id
			  AND s.last_played_at >= %s
			  AND s.last_played_at < NOW()
			WHERE u.is_active = true
			GROUP BY u.id, u.username, u.nickname, u.avatar_id
			HAVING COALESCE(SUM(s.play_time), 0) > 0
		)
		SELECT * FROM ranked WHERE rank <= 100 ORDER BY rank
	`, windowStartSQL(period))).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query windowed playtime leaderboard: %w", err)
	}

	return buildLeaderboardResult(rows, userID), nil
}

// windowStartSQL returns a PostgreSQL expression for the start of the period.
func windowStartSQL(period string) string {
	switch period {
	case "day":
		return "DATE_TRUNC('day', NOW())"
	case "week":
		return "DATE_TRUNC('week', NOW())"
	case "month":
		return "DATE_TRUNC('month', NOW())"
	default:
		return "DATE_TRUNC('day', NOW())"
	}
}

// buildLeaderboardResult converts raw rows into entries and finds the current user's rank.
func buildLeaderboardResult(rows []leaderboardRow, userID string) *LeaderboardResult {
	entries := make([]LeaderboardEntry, 0, len(rows))

	for _, r := range rows {
		var avatarURL *string
		if r.AvatarID != nil && *r.AvatarID != "" {
			url := helpers.ImageServeURL(*r.AvatarID)
			avatarURL = &url
		}

		entries = append(entries, LeaderboardEntry{
			ID:        r.ID,
			Username:  r.Username,
			Nickname:  r.Nickname,
			AvatarURL: avatarURL,
			Value:     r.Value,
			Rank:      r.Rank,
		})
	}

	var myRank *LeaderboardEntry
	for i := range entries {
		if entries[i].ID == userID {
			myRank = &entries[i]
			break
		}
	}

	return &LeaderboardResult{
		Entries: entries,
		MyRank:  myRank,
	}
}
