package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// LevelWinnerResult holds winner data for SSE broadcast.
type LevelWinnerResult struct {
	GameLevelID  string       `json:"game_level_id"`
	Mode         string       `json:"mode"`
	Winner       any          `json:"winner"`
	Participants []SoloWinner `json:"participants"`
	Teams        []TeamWinner `json:"teams,omitempty"`
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

	// Lock participant rows to prevent concurrent winner determination
	type idRow struct {
		ID string `gorm:"column:id"`
	}
	var lockedRows []idRow
	if err := tx.Raw(
		"SELECT id FROM game_session_totals WHERE game_group_id = ? AND ended_at IS NULL FOR UPDATE",
		gameGroupID).Scan(&lockedRows); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to lock participants: %w", err)
	}
	participantCount := int64(len(lockedRows))

	// Count completed level sessions
	var completedRow countRow
	if err := tx.Raw(
		"SELECT COUNT(*) AS count FROM game_session_levels WHERE game_group_id = ? AND game_level_id = ? AND ended_at IS NOT NULL",
		gameGroupID, gameLevelID).Scan(&completedRow); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to count completed levels: %w", err)
	}

	fmt.Printf("[GROUP] Winner check: participants=%d completed=%d\n", participantCount, completedRow.Count)
	if completedRow.Count < participantCount {
		_ = tx.Rollback()
		return nil, nil // Still waiting
	}

	_ = tx.Commit()

	// All done — determine winner
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", gameGroupID).First(&group); err != nil || group.ID == "" {
		return nil, ErrGroupNotFound
	}

	if group.GameMode != nil && *group.GameMode == consts.GameModeTeam {
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

	if group.GameMode != nil && *group.GameMode == consts.GameModeTeam {
		return determineTeamWinner(gameGroupID, gameLevelID)
	}
	return determineSoloWinner(gameGroupID, gameLevelID)
}

type soloWinnerRow struct {
	UserID   string    `gorm:"column:user_id"`
	Nickname *string   `gorm:"column:nickname"`
	Score    int       `gorm:"column:score"`
	EndedAt  time.Time `gorm:"column:ended_at"`
}

func determineSoloWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	var rows []soloWinnerRow
	if err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, u.nickname, gsl.score, gsl.ended_at
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 JOIN users u ON u.id = gst.user_id
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		 ORDER BY gsl.score DESC, gsl.ended_at ASC`, gameGroupID, gameLevelID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query solo participants: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	// Build participants list
	participants := make([]SoloWinner, len(rows))
	for i, r := range rows {
		participants[i] = SoloWinner{
			UserID:   r.UserID,
			UserName: derefNickname(r.Nickname),
			Score:    r.Score,
		}
	}

	// Winner is first participant
	winner := participants[0]

	// Update last_won_at for winner only
	now := time.Now()
	facades.Orm().Query().Exec(
		"UPDATE game_group_members SET last_won_at = ? WHERE game_group_id = ? AND user_id = ?",
		now, gameGroupID, winner.UserID)

	return &LevelWinnerResult{
		GameLevelID:  gameLevelID,
		Mode:         consts.GameModeSolo,
		Winner:       winner,
		Participants: participants,
	}, nil
}

type teamWinnerRow struct {
	SubgroupID  string    `gorm:"column:game_subgroup_id"`
	TotalScore  int       `gorm:"column:total_score"`
	LastEndedAt time.Time `gorm:"column:last_ended_at"`
}

func determineTeamWinner(gameGroupID, gameLevelID string) (*LevelWinnerResult, error) {
	// Fetch ALL subgroups ranked by sum of scores
	var teamRows []teamWinnerRow
	if err := facades.Orm().Query().Raw(
		`SELECT gsl.game_subgroup_id, SUM(gsl.score) AS total_score, MAX(gsl.ended_at) AS last_ended_at
		 FROM game_session_levels gsl
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		   AND gsl.game_subgroup_id IS NOT NULL
		 GROUP BY gsl.game_subgroup_id
		 ORDER BY total_score DESC, last_ended_at ASC`, gameGroupID, gameLevelID).Scan(&teamRows); err != nil {
		return nil, fmt.Errorf("failed to query team results: %w", err)
	}

	if len(teamRows) == 0 {
		return nil, nil
	}

	// Fetch ALL member scores across all subgroups, JOIN users for nicknames
	var memberRows []struct {
		UserID     string  `gorm:"column:user_id"`
		Nickname   *string `gorm:"column:nickname"`
		Score      int     `gorm:"column:score"`
		SubgroupID string  `gorm:"column:game_subgroup_id"`
	}
	if err := facades.Orm().Query().Raw(
		`SELECT gst.user_id, u.nickname, gsl.score, gsl.game_subgroup_id
		 FROM game_session_levels gsl
		 JOIN game_session_totals gst ON gst.id = gsl.game_session_total_id
		 JOIN users u ON u.id = gst.user_id
		 WHERE gsl.game_group_id = ? AND gsl.game_level_id = ? AND gsl.ended_at IS NOT NULL
		   AND gsl.game_subgroup_id IS NOT NULL
		 ORDER BY gsl.score DESC`, gameGroupID, gameLevelID).Scan(&memberRows); err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}

	// Group members by subgroup
	membersBySubgroup := make(map[string][]TeamMember)
	var allParticipants []SoloWinner
	for _, mr := range memberRows {
		name := derefNickname(mr.Nickname)
		membersBySubgroup[mr.SubgroupID] = append(membersBySubgroup[mr.SubgroupID], TeamMember{
			UserID: mr.UserID, UserName: name, Score: mr.Score,
		})
		allParticipants = append(allParticipants, SoloWinner{
			UserID: mr.UserID, UserName: name, Score: mr.Score,
		})
	}

	// Fetch subgroup names
	subgroupIDs := make([]string, len(teamRows))
	for i, tr := range teamRows {
		subgroupIDs[i] = tr.SubgroupID
	}
	var subgroups []models.GameSubgroup
	facades.Orm().Query().Where("id IN ?", subgroupIDs).Find(&subgroups)
	subgroupNames := make(map[string]string)
	for _, sg := range subgroups {
		subgroupNames[sg.ID] = sg.Name
	}

	// Build teams slice
	teams := make([]TeamWinner, len(teamRows))
	for i, tr := range teamRows {
		teams[i] = TeamWinner{
			SubgroupID:   tr.SubgroupID,
			SubgroupName: subgroupNames[tr.SubgroupID],
			TotalScore:   tr.TotalScore,
			Members:      membersBySubgroup[tr.SubgroupID],
		}
	}

	// Winner is first team
	winnerTeam := teams[0]

	// Update last_won_at on winning subgroup only
	now := time.Now()
	facades.Orm().Query().Exec(
		"UPDATE game_subgroups SET last_won_at = ? WHERE id = ?",
		now, winnerTeam.SubgroupID)

	// Update last_won_at on winning team's members only
	winnerMemberIDs := make([]string, len(winnerTeam.Members))
	for i, m := range winnerTeam.Members {
		winnerMemberIDs[i] = m.UserID
	}
	if len(winnerMemberIDs) > 0 {
		facades.Orm().Query().Exec(
			"UPDATE game_group_members SET last_won_at = ? WHERE game_group_id = ? AND user_id IN ?",
			now, gameGroupID, winnerMemberIDs)
	}

	return &LevelWinnerResult{
		GameLevelID:  gameLevelID,
		Mode:         consts.GameModeTeam,
		Winner:       winnerTeam,
		Participants: allParticipants,
		Teams:        teams,
	}, nil
}
