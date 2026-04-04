package api

import (
	"fmt"
	"time"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

type PkWinnerResult struct {
	GameLevelID  string     `json:"game_level_id"`
	Winner       PkWinner   `json:"winner"`
	Participants []PkWinner `json:"participants"`
}

type PkWinner struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

// DeterminePkWinner compares the two players' level scores for a given PK match and level.
func DeterminePkWinner(pkID, gameLevelID string) (*PkWinnerResult, error) {
	type scoreRow struct {
		UserID   string     `gorm:"column:user_id"`
		Nickname *string    `gorm:"column:nickname"`
		Score    int        `gorm:"column:score"`
		EndedAt  *time.Time `gorm:"column:ended_at"`
	}

	var rows []scoreRow
	if err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, u.nickname, gsl.score, gsl.ended_at
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 JOIN users u ON u.id = gst.user_id
		 WHERE gsl.game_pk_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		 ORDER BY gsl.score DESC, gsl.ended_at ASC`,
		pkID, gameLevelID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query pk scores: %w", err)
	}

	if len(rows) < 2 {
		return nil, nil
	}

	participants := make([]PkWinner, len(rows))
	for i, r := range rows {
		name := ""
		if r.Nickname != nil {
			name = *r.Nickname
		}
		participants[i] = PkWinner{
			UserID:   r.UserID,
			UserName: name,
			Score:    r.Score,
		}
	}

	winner := participants[0]

	// Update last_winner_id on game_pks
	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).Update("last_winner_id", winner.UserID)

	return &PkWinnerResult{
		GameLevelID:  gameLevelID,
		Winner:       winner,
		Participants: participants,
	}, nil
}
