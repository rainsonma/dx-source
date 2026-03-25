package api

import (
	"fmt"
	"time"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// LevelWinnerResult holds winner data for SSE broadcast.
type LevelWinnerResult struct {
	GameLevelID string `json:"game_level_id"`
	Mode        string `json:"mode"`
	Winner      any    `json:"winner"`
}

type SoloWinner struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

type TeamWinner struct {
	SubgroupID   string       `json:"subgroup_id"`
	SubgroupName string       `json:"subgroup_name"`
	TotalScore   int          `json:"total_score"`
	Members      []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

type countRow struct {
	Count int64 `gorm:"column:count"`
}

func derefNickname(n *string) string {
	if n == nil {
		return ""
	}
	return *n
}

// CheckAndDetermineWinner is called after each level completion. It checks if
// all participants have completed the level and, if so, determines the winner.
// Uses FOR UPDATE locking for concurrency safety.
func CheckAndDetermineWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// Count participants with FOR UPDATE lock to prevent concurrent winner determination
	var participantRow countRow
	if err := tx.Raw(
		"SELECT COUNT(*) AS count FROM game_session_totals WHERE game_group_id = ? AND ended_at IS NULL FOR UPDATE",
		gameGroupID).Scan(&participantRow); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to count participants: %w", err)
	}

	// Count completed level sessions
	var completedRow countRow
	if err := tx.Raw(
		"SELECT COUNT(*) AS count FROM game_session_levels WHERE game_group_id = ? AND game_level_id = ? AND ended_at IS NOT NULL",
		gameGroupID, gameLevelID).Scan(&completedRow); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to count completed levels: %w", err)
	}

	if completedRow.Count < participantRow.Count {
		_ = tx.Rollback()
		return nil, nil // Still waiting
	}

	_ = tx.Commit()

	// All done — determine winner
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", gameGroupID).First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}

	if group.GameMode != nil && *group.GameMode == "team" {
		return determineTeamWinner(gameGroupID, gameLevelID)
	}
	return determineSoloWinner(gameGroupID, gameLevelID)
}

// DetermineWinnerForLevel is used by force-end (sessions already ended, skip participant count check).
func DetermineWinnerForLevel(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", gameGroupID).First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}

	if group.GameMode != nil && *group.GameMode == "team" {
		return determineTeamWinner(gameGroupID, gameLevelID)
	}
	return determineSoloWinner(gameGroupID, gameLevelID)
}

type soloWinnerRow struct {
	UserID   string    `gorm:"column:user_id"`
	Score    int       `gorm:"column:score"`
	EndedAt  time.Time `gorm:"column:ended_at"`
}

func determineSoloWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	var rows []soloWinnerRow
	if err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, gsl.score, gsl.ended_at
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		 ORDER BY gsl.score DESC, gsl.ended_at ASC
		 LIMIT 1`, gameGroupID, gameLevelID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query solo winner: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	winner := rows[0]

	// Get username
	var user models.User
	facades.Orm().Query().Where("id", winner.UserID).First(&user)

	// Update last_won_at
	now := time.Now()
	facades.Orm().Query().Exec(
		"UPDATE game_group_members SET last_won_at = ? WHERE game_group_id = ? AND user_id = ?",
		now, gameGroupID, winner.UserID)

	return &LevelWinnerResult{
		GameLevelID: gameLevelID,
		Mode:        "solo",
		Winner: SoloWinner{
			UserID:   winner.UserID,
			UserName: derefNickname(user.Nickname),
			Score:    winner.Score,
		},
	}, nil
}

type teamWinnerRow struct {
	SubgroupID  string    `gorm:"column:game_subgroup_id"`
	TotalScore  int       `gorm:"column:total_score"`
	LastEndedAt time.Time `gorm:"column:last_ended_at"`
}

type teamMemberRow struct {
	UserID string `gorm:"column:user_id"`
	Score  int    `gorm:"column:score"`
}

func determineTeamWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	// Find winning subgroup by sum of scores
	var winnerRows []teamWinnerRow
	if err := facades.Orm().Query().Raw(
		`SELECT gsl.game_subgroup_id, SUM(gsl.score) AS total_score, MAX(gsl.ended_at) AS last_ended_at
		 FROM game_session_levels gsl
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		   AND gsl.game_subgroup_id IS NOT NULL
		 GROUP BY gsl.game_subgroup_id
		 ORDER BY total_score DESC, last_ended_at ASC
		 LIMIT 1`, gameGroupID, gameLevelID).Scan(&winnerRows); err != nil {
		return nil, fmt.Errorf("failed to query team winner: %w", err)
	}

	if len(winnerRows) == 0 {
		return nil, nil
	}

	winnerSubgroupID := winnerRows[0].SubgroupID
	totalScore := winnerRows[0].TotalScore

	// Get subgroup name
	var subgroup models.GameSubgroup
	facades.Orm().Query().Where("id", winnerSubgroupID).First(&subgroup)

	// Get individual member scores
	var memberRows []teamMemberRow
	if err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, gsl.score
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		   AND gsl.game_subgroup_id = ?
		 ORDER BY gsl.score DESC`, gameGroupID, gameLevelID, winnerSubgroupID).Scan(&memberRows); err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}

	now := time.Now()
	var members []TeamMember
	var memberUserIDs []string
	for _, mr := range memberRows {
		var user models.User
		facades.Orm().Query().Where("id", mr.UserID).First(&user)
		members = append(members, TeamMember{UserID: mr.UserID, UserName: derefNickname(user.Nickname), Score: mr.Score})
		memberUserIDs = append(memberUserIDs, mr.UserID)
	}

	// Update last_won_at on winning subgroup
	facades.Orm().Query().Exec(
		"UPDATE game_subgroups SET last_won_at = ? WHERE id = ?",
		now, winnerSubgroupID)

	// Update last_won_at on all participating members of winning subgroup
	if len(memberUserIDs) > 0 {
		facades.Orm().Query().Exec(
			"UPDATE game_group_members SET last_won_at = ? WHERE game_group_id = ? AND user_id IN ?",
			now, gameGroupID, memberUserIDs)
	}

	return &LevelWinnerResult{
		GameLevelID: gameLevelID,
		Mode:        "team",
		Winner: TeamWinner{
			SubgroupID:   winnerSubgroupID,
			SubgroupName: subgroup.Name,
			TotalScore:   totalScore,
			Members:      members,
		},
	}, nil
}
