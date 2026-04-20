package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// DashboardData aggregates user stats for the hall dashboard.
type DashboardData struct {
	Profile      DashboardProfile  `json:"profile"`
	MasterStats  MasterStats       `json:"masterStats"`
	ReviewStats  ReviewStats       `json:"reviewStats"`
	Sessions     []SessionProgress `json:"sessions"`
	TodayAnswers int               `json:"todayAnswers"`
	Greeting     consts.Greeting   `json:"greeting"`
}

// DashboardProfile is the user profile subset shown on the dashboard.
type DashboardProfile struct {
	ID                string  `json:"id"`
	Username          string  `json:"username"`
	Nickname          *string `json:"nickname"`
	Grade             string  `json:"grade"`
	Level             int     `json:"level"`
	Exp               int     `json:"exp"`
	Beans             int     `json:"beans"`
	AvatarURL         *string `json:"avatarUrl"`
	CurrentPlayStreak int     `json:"currentPlayStreak"`
	InviteCode        string  `json:"inviteCode"`
	LastReadNoticeAt  any     `json:"lastReadNoticeAt"`
	CreatedAt         any     `json:"createdAt"`
}

// MasterStats holds mastered content statistics.
type MasterStats struct {
	Total     int `json:"total"`
	ThisWeek  int `json:"thisWeek"`
	ThisMonth int `json:"thisMonth"`
}

// ReviewStats holds review queue statistics.
type ReviewStats struct {
	Pending       int `json:"pending"`
	Overdue       int `json:"overdue"`
	ReviewedToday int `json:"reviewedToday"`
}

// SessionProgress represents a recent game session with progress.
type SessionProgress struct {
	GameID          string    `json:"gameId"`
	GameName        string    `json:"gameName"`
	GameMode        string    `json:"gameMode"`
	CompletedLevels int       `json:"completedLevels"`
	TotalLevels     int       `json:"totalLevels"`
	Score           int       `json:"score"`
	Exp             int       `json:"exp"`
	LastPlayedAt    time.Time `json:"lastPlayedAt"`
}

// HeatmapData contains daily activity counts for a given year.
type HeatmapData struct {
	Year        int          `json:"year"`
	Days        []HeatmapDay `json:"days"`
	AccountYear int          `json:"accountYear"`
}

// HeatmapDay represents the answer count for a single day.
type HeatmapDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GetDashboard returns aggregated dashboard data for a user.
func GetDashboard(userID string) (*DashboardData, error) {
	query := facades.Orm().Query()

	// Fetch user profile
	var user models.User
	if err := query.Where("id", userID).First(&user); err != nil || user.ID == "" {
		return nil, ErrUserNotFound
	}

	level, err := consts.GetLevel(user.Exp)
	if err != nil {
		return nil, fmt.Errorf("failed to compute user level: %w", err)
	}

	profile := DashboardProfile{
		ID:                user.ID,
		Username:          user.Username,
		Nickname:          user.Nickname,
		Grade:             user.Grade,
		Level:             level,
		Exp:               user.Exp,
		Beans:             user.Beans,
		AvatarURL:         user.AvatarURL,
		CurrentPlayStreak: user.CurrentPlayStreak,
		InviteCode:        user.InviteCode,
		LastReadNoticeAt:  user.LastReadNoticeAt,
		CreatedAt:         user.CreatedAt,
	}

	// Master stats
	masterStats, err := getMasterStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get master stats: %w", err)
	}

	// Review stats
	reviewStats, err := getReviewStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get review stats: %w", err)
	}

	// Recent session progress
	sessions, err := getSessionProgress(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session progress: %w", err)
	}

	// Today's answer count
	todayAnswers, err := getTodayAnswerCount(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get today answer count: %w", err)
	}

	return &DashboardData{
		Profile:      profile,
		MasterStats:  *masterStats,
		ReviewStats:  *reviewStats,
		Sessions:     sessions,
		TodayAnswers: todayAnswers,
		Greeting:     consts.PickGreeting(time.Now()),
	}, nil
}

// getMasterStats returns mastered item counts (total, this week, this month).
func getMasterStats(userID string) (*MasterStats, error) {
	now := time.Now()
	weekStart := mondayOfWeek(now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type countRow struct {
		Total     int `gorm:"column:total"`
		ThisWeek  int `gorm:"column:this_week"`
		ThisMonth int `gorm:"column:this_month"`
	}

	var row countRow
	if err := facades.Orm().Query().Raw(`
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE created_at >= ?) AS this_week,
			COUNT(*) FILTER (WHERE created_at >= ?) AS this_month
		FROM user_masters
		WHERE user_id = ?
	`, weekStart, monthStart, userID).Scan(&row); err != nil {
		return nil, fmt.Errorf("failed to query master stats: %w", err)
	}

	return &MasterStats{
		Total:     row.Total,
		ThisWeek:  row.ThisWeek,
		ThisMonth: row.ThisMonth,
	}, nil
}

// getReviewStats returns review queue counts (pending, overdue, reviewed today).
func getReviewStats(userID string) (*ReviewStats, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type countRow struct {
		Pending       int `gorm:"column:pending"`
		Overdue       int `gorm:"column:overdue"`
		ReviewedToday int `gorm:"column:reviewed_today"`
	}

	var row countRow
	if err := facades.Orm().Query().Raw(`
		SELECT
			COUNT(*) FILTER (WHERE next_review_at <= ?) AS pending,
			COUNT(*) FILTER (WHERE next_review_at < ?) AS overdue,
			COUNT(*) FILTER (WHERE last_review_at >= ?) AS reviewed_today
		FROM user_reviews
		WHERE user_id = ?
	`, now, todayStart, todayStart, userID).Scan(&row); err != nil {
		return nil, fmt.Errorf("failed to query review stats: %w", err)
	}

	return &ReviewStats{
		Pending:       row.Pending,
		Overdue:       row.Overdue,
		ReviewedToday: row.ReviewedToday,
	}, nil
}

// getSessionProgress returns recent game session progress entries.
func getSessionProgress(userID string) ([]SessionProgress, error) {
	type SessionProgressItem struct {
		GameID          string    `gorm:"column:game_id"`
		GameName        string    `gorm:"column:game_name"`
		GameMode        string    `gorm:"column:game_mode"`
		CompletedLevels int       `gorm:"column:completed_levels"`
		TotalLevels     int       `gorm:"column:total_levels"`
		Score           int       `gorm:"column:score"`
		Exp             int       `gorm:"column:exp"`
		LastPlayedAt    time.Time `gorm:"column:last_played_at"`
	}

	var rows []SessionProgressItem
	if err := facades.Orm().Query().Raw(`
		SELECT
		  s.game_id,
		  g.name AS game_name,
		  g.mode AS game_mode,
		  COUNT(DISTINCT s.game_level_id) FILTER (WHERE s.ended_at IS NOT NULL)::int AS completed_levels,
		  (SELECT COUNT(*)::int FROM game_levels gl WHERE gl.game_id = s.game_id AND gl.is_active = true) AS total_levels,
		  COALESCE(SUM(s.score), 0)::int AS score,
		  COALESCE(SUM(s.exp), 0)::int AS exp,
		  MAX(s.last_played_at) AS last_played_at
		FROM game_sessions s
		INNER JOIN games g ON g.id = s.game_id
		WHERE s.user_id = ? AND s.game_group_id IS NULL AND s.game_pk_id IS NULL
		GROUP BY s.game_id, g.name, g.mode
		ORDER BY MAX(s.last_played_at) DESC
		LIMIT 20
	`, userID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query session progress: %w", err)
	}

	results := make([]SessionProgress, 0, len(rows))
	for _, r := range rows {
		results = append(results, SessionProgress{
			GameID:          r.GameID,
			GameName:        r.GameName,
			GameMode:        r.GameMode,
			CompletedLevels: r.CompletedLevels,
			TotalLevels:     r.TotalLevels,
			Score:           r.Score,
			Exp:             r.Exp,
			LastPlayedAt:    r.LastPlayedAt,
		})
	}

	return results, nil
}

// getTodayAnswerCount returns the number of answers recorded today.
func getTodayAnswerCount(userID string) (int, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type countRow struct {
		Count int `gorm:"column:count"`
	}

	var row countRow
	if err := facades.Orm().Query().Raw(`
		SELECT COUNT(*) AS count
		FROM game_records
		WHERE user_id = ? AND created_at >= ?
	`, userID, todayStart).Scan(&row); err != nil {
		return 0, fmt.Errorf("failed to count today answers: %w", err)
	}

	return row.Count, nil
}

// GetHeatmap returns daily answer counts for a given year.
func GetHeatmap(userID string, year int) (*HeatmapData, error) {
	// Determine account year from user's created_at
	var user models.User
	if err := facades.Orm().Query().Select("id", "created_at").Where("id", userID).First(&user); err != nil || user.ID == "" {
		return nil, ErrUserNotFound
	}

	accountYear := user.CreatedAt.Year()
	currentYear := time.Now().Year()

	// Clamp year to valid range
	if year < accountYear {
		year = accountYear
	}
	if year > currentYear {
		year = currentYear
	}

	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	type dayRow struct {
		Date  string `gorm:"column:date"`
		Count int    `gorm:"column:count"`
	}

	var rows []dayRow
	if err := facades.Orm().Query().Raw(`
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS date, COUNT(*)::int AS count
		FROM game_records
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY TO_CHAR(created_at, 'YYYY-MM-DD')
		ORDER BY date
	`, userID, yearStart, yearEnd).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query heatmap data: %w", err)
	}

	days := make([]HeatmapDay, 0, len(rows))
	for _, r := range rows {
		days = append(days, HeatmapDay{
			Date:  r.Date,
			Count: r.Count,
		})
	}

	return &HeatmapData{
		Year:        year,
		Days:        days,
		AccountYear: accountYear,
	}, nil
}

// mondayOfWeek returns the Monday 00:00 of the current week.
func mondayOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}
