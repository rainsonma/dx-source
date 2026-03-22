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

// GetLeaderboard returns a ranked list by type (exp|playtime) and period (all|day|week|month).
func GetLeaderboard(lbType, period, userID string) (*LeaderboardResult, error) {
	if period == "all" {
		if lbType == "exp" {
			return getAllTimeExp(userID)
		}
		return getAllTimePlayTime(userID)
	}
	if lbType == "exp" {
		return getWindowedExp(period, userID)
	}
	return getWindowedPlayTime(period, userID)
}

// getAllTimeExp ranks active users by total exp.
func getAllTimeExp(userID string) (*LeaderboardResult, error) {
	var rows []leaderboardRow
	if err := facades.Orm().Query().Raw(`
		WITH ranked AS (
			SELECT u.id, u.username, u.nickname, u.avatar_id, u.exp AS value,
			       RANK() OVER (ORDER BY u.exp DESC)::int AS rank
			FROM users u
			WHERE u.is_active = true AND u.exp > 0
		)
		SELECT * FROM ranked
		WHERE rank <= 100 OR id = ?
		ORDER BY rank
	`, safeUID(userID)).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query all-time exp leaderboard: %w", err)
	}

	return buildLeaderboardResult(rows, userID), nil
}

// getAllTimePlayTime ranks active users by total play time across all games.
func getAllTimePlayTime(userID string) (*LeaderboardResult, error) {
	var rows []leaderboardRow
	if err := facades.Orm().Query().Raw(`
		WITH ranked AS (
			SELECT u.id, u.username, u.nickname, u.avatar_id,
			       COALESCE(SUM(g.total_play_time), 0)::int AS value,
			       RANK() OVER (ORDER BY COALESCE(SUM(g.total_play_time), 0) DESC)::int AS rank
			FROM users u
			INNER JOIN game_stats_totals g ON g.user_id = u.id
			WHERE u.is_active = true
			GROUP BY u.id, u.username, u.nickname, u.avatar_id
			HAVING COALESCE(SUM(g.total_play_time), 0) > 0
		)
		SELECT * FROM ranked
		WHERE rank <= 100 OR id = ?
		ORDER BY rank
	`, safeUID(userID)).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query all-time playtime leaderboard: %w", err)
	}

	return buildLeaderboardResult(rows, userID), nil
}

// getWindowedExp ranks active users by exp earned within a time window.
func getWindowedExp(period, userID string) (*LeaderboardResult, error) {
	var rows []leaderboardRow
	if err := facades.Orm().Query().Raw(fmt.Sprintf(`
		WITH ranked AS (
			SELECT u.id, u.username, u.nickname, u.avatar_id,
			       COALESCE(SUM(s.exp), 0)::int AS value,
			       RANK() OVER (ORDER BY COALESCE(SUM(s.exp), 0) DESC)::int AS rank
			FROM users u
			INNER JOIN game_session_totals s ON s.user_id = u.id
			  AND s.last_played_at >= %s
			  AND s.last_played_at < NOW()
			WHERE u.is_active = true
			GROUP BY u.id, u.username, u.nickname, u.avatar_id
			HAVING COALESCE(SUM(s.exp), 0) > 0
		)
		SELECT * FROM ranked
		WHERE rank <= 100 OR id = ?
		ORDER BY rank
	`, windowStartSQL(period)), safeUID(userID)).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query windowed exp leaderboard: %w", err)
	}

	return buildLeaderboardResult(rows, userID), nil
}

// getWindowedPlayTime ranks active users by play time within a time window.
func getWindowedPlayTime(period, userID string) (*LeaderboardResult, error) {
	var rows []leaderboardRow
	if err := facades.Orm().Query().Raw(fmt.Sprintf(`
		WITH ranked AS (
			SELECT u.id, u.username, u.nickname, u.avatar_id,
			       COALESCE(SUM(s.play_time), 0)::int AS value,
			       RANK() OVER (ORDER BY COALESCE(SUM(s.play_time), 0) DESC)::int AS rank
			FROM users u
			INNER JOIN game_session_totals s ON s.user_id = u.id
			  AND s.last_played_at >= %s
			  AND s.last_played_at < NOW()
			WHERE u.is_active = true
			GROUP BY u.id, u.username, u.nickname, u.avatar_id
			HAVING COALESCE(SUM(s.play_time), 0) > 0
		)
		SELECT * FROM ranked
		WHERE rank <= 100 OR id = ?
		ORDER BY rank
	`, windowStartSQL(period)), safeUID(userID)).Scan(&rows); err != nil {
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

// safeUID returns a non-matchable placeholder if userID is empty.
func safeUID(userID string) string {
	if userID == "" {
		return "___none___"
	}
	return userID
}

// buildLeaderboardResult splits rows into top-100 entries and current user's rank.
func buildLeaderboardResult(rows []leaderboardRow, userID string) *LeaderboardResult {
	entries := make([]LeaderboardEntry, 0, len(rows))
	var myRank *LeaderboardEntry

	for _, r := range rows {
		var avatarURL *string
		if r.AvatarID != nil && *r.AvatarID != "" {
			url := helpers.ImageServeURL(*r.AvatarID)
			avatarURL = &url
		}

		entry := LeaderboardEntry{
			ID:        r.ID,
			Username:  r.Username,
			Nickname:  r.Nickname,
			AvatarURL: avatarURL,
			Value:     r.Value,
			Rank:      r.Rank,
		}
		if r.Rank <= 100 {
			entries = append(entries, entry)
		}
		if userID != "" && r.ID == userID {
			e := entry
			myRank = &e
		}
	}

	return &LeaderboardResult{
		Entries: entries,
		MyRank:  myRank,
	}
}
