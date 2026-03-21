package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameStatsTotal struct {
	orm.Timestamps
	ID               string     `gorm:"column:id;primaryKey" json:"id"`
	UserID           string     `gorm:"column:user_id" json:"user_id"`
	GameID           string     `gorm:"column:game_id" json:"game_id"`
	TotalSessions    int        `gorm:"column:total_sessions" json:"total_sessions"`
	TotalExp         int        `gorm:"column:total_exp" json:"total_exp"`
	HighestScore     int        `gorm:"column:highest_score" json:"highest_score"`
	TotalScores      int        `gorm:"column:total_scores" json:"total_scores"`
	TotalPlayTime    int        `gorm:"column:total_play_time" json:"total_play_time"`
	FirstPlayedAt    time.Time  `gorm:"column:first_played_at" json:"first_played_at"`
	LastPlayedAt     time.Time  `gorm:"column:last_played_at" json:"last_played_at"`
	FirstCompletedAt *time.Time `gorm:"column:first_completed_at" json:"first_completed_at"`
	LastCompletedAt  *time.Time `gorm:"column:last_completed_at" json:"last_completed_at"`
	CompletionCount  int        `gorm:"column:completion_count" json:"completion_count"`
}

func (g *GameStatsTotal) TableName() string {
	return "game_stats_totals"
}
