package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameSession struct {
	orm.Timestamps
	ID                   string     `gorm:"column:id;primaryKey" json:"id"`
	UserID               string     `gorm:"column:user_id" json:"user_id"`
	GameID               string     `gorm:"column:game_id" json:"game_id"`
	GameLevelID          string     `gorm:"column:game_level_id" json:"game_level_id"`
	Degree               string     `gorm:"column:degree" json:"degree"`
	Pattern              *string    `gorm:"column:pattern" json:"pattern"`
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
	TotalItemsCount      int        `gorm:"column:total_items_count" json:"total_items_count"`
	PlayedItemsCount     int        `gorm:"column:played_items_count" json:"played_items_count"`
	GameGroupID          *string    `gorm:"column:game_group_id" json:"game_group_id"`
	GameSubgroupID       *string    `gorm:"column:game_subgroup_id" json:"game_subgroup_id"`
	GamePkID             *string    `gorm:"column:game_pk_id" json:"game_pk_id"`
}

func (g *GameSession) TableName() string {
	return "game_sessions"
}
