package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameSessionTotal struct {
	orm.Timestamps
	ID                   string     `gorm:"column:id;primaryKey" json:"id"`
	UserID               string     `gorm:"column:user_id" json:"user_id"`
	GameID               string     `gorm:"column:game_id" json:"game_id"`
	Degree               string     `gorm:"column:degree" json:"degree"`
	Pattern              *string    `gorm:"column:pattern" json:"pattern"`
	CurrentLevelID       *string    `gorm:"column:current_level_id" json:"current_level_id"`
	CurrentContentItemID *string    `gorm:"column:current_content_item_id" json:"current_content_item_id"`
	StartedAt            time.Time  `gorm:"column:started_at" json:"started_at"`
	LastPlayedAt         time.Time  `gorm:"column:last_played_at" json:"last_played_at"`
	EndedAt              *time.Time `gorm:"column:ended_at" json:"ended_at"`
	Score                int        `gorm:"column:score" json:"score"`
	Exp                  int        `gorm:"column:exp" json:"exp"`
	MaxCombo             int        `gorm:"column:max_combo" json:"max_combo"`
	CorrectCount         int        `gorm:"column:correct_count" json:"correct_count"`
	WrongCount           int        `gorm:"column:wrong_count" json:"wrong_count"`
	SkipCount            int        `gorm:"column:skip_count" json:"skip_count"`
	PlayTime             int        `gorm:"column:play_time" json:"play_time"`
	TotalLevelsCount     int        `gorm:"column:total_levels_count" json:"total_levels_count"`
	PlayedLevelsCount    int        `gorm:"column:played_levels_count" json:"played_levels_count"`
}

func (g *GameSessionTotal) TableName() string {
	return "game_session_totals"
}
