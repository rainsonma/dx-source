package models

import "github.com/goravel/framework/database/orm"

type GamePk struct {
	orm.Timestamps
	ID              string  `gorm:"column:id;primaryKey" json:"id"`
	UserID          string  `gorm:"column:user_id" json:"user_id"`
	OpponentID      string  `gorm:"column:opponent_id" json:"opponent_id"`
	GameID          string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID     string  `gorm:"column:game_level_id" json:"game_level_id"`
	Degree          string  `gorm:"column:degree" json:"degree"`
	Pattern         *string `gorm:"column:pattern" json:"pattern"`
	RobotDifficulty string  `gorm:"column:robot_difficulty" json:"robot_difficulty"`
	IsPlaying        bool    `gorm:"column:is_playing" json:"is_playing"`
	LastWinnerID     *string `gorm:"column:last_winner_id" json:"last_winner_id"`
	PkType           string  `gorm:"column:pk_type" json:"pk_type"`
	InvitationStatus *string `gorm:"column:invitation_status" json:"invitation_status"`
}

func (g *GamePk) TableName() string {
	return "game_pks"
}
