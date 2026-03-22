package api

import (
	"fmt"
	"time"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/database/orm"
)

// UpsertGameStats creates game stats if they don't exist, or increments totalSessions.
func UpsertGameStats(userID, gameID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`INSERT INTO game_stats_totals (id, user_id, game_id, total_sessions, first_played_at, last_played_at, created_at, updated_at)
		 VALUES (?, ?, ?, 1, ?, ?, now(), now())
		 ON CONFLICT (user_id, game_id) DO UPDATE SET
		   total_sessions = game_stats_totals.total_sessions + 1,
		   last_played_at = EXCLUDED.last_played_at,
		   updated_at = now()`,
		newID(), userID, gameID, now, now,
	)
	return err
}

// UpdateGameStatsAfterSession updates completion stats after a session ends.
func UpdateGameStatsAfterSession(userID, gameID string, allCompleted bool) error {
	if !allCompleted {
		return nil
	}
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		"UPDATE game_stats_totals SET completion_count = completion_count + 1, last_completed_at = ?, updated_at = now() WHERE user_id = ? AND game_id = ?",
		now, userID, gameID,
	)
	return err
}

// MarkGameFirstCompletion sets firstCompletedAt if this is the first completion.
func MarkGameFirstCompletion(userID, gameID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Model(&models.GameStatsTotal{}).
		Where("user_id", userID).Where("game_id", gameID).
		Where("first_completed_at IS NULL").
		Update("first_completed_at", now)
	return err
}

// UpsertLevelStats creates level stats if they don't exist, or updates lastPlayedAt.
func UpsertLevelStats(userID, gameLevelID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`INSERT INTO game_stats_levels (id, user_id, game_level_id, first_played_at, last_played_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, now(), now())
		 ON CONFLICT (user_id, game_level_id) DO UPDATE SET
		   last_played_at = EXCLUDED.last_played_at,
		   updated_at = now()`,
		newID(), userID, gameLevelID, now, now,
	)
	return err
}

// completeLevelStatsInTx records a level completion within a transaction.
func completeLevelStatsInTx(tx orm.Query, userID, gameLevelID string, score, playTimeSeconds int) error {
	now := time.Now()

	var existing models.GameStatsLevel
	if err := tx.Where("user_id", userID).Where("game_level_id", gameLevelID).
		First(&existing); err != nil || existing.ID == "" {
		return fmt.Errorf("level stats not found for user %s, level %s", userID, gameLevelID)
	}

	// Build dynamic SET clause for conditional fields
	setClause := "completion_count = completion_count + 1, last_completed_at = ?, total_play_time = total_play_time + ?, total_scores = total_scores + ?, updated_at = now()"
	args := []any{now, playTimeSeconds, score}

	if existing.FirstCompletedAt == nil {
		setClause += ", first_completed_at = ?"
		args = append(args, now)
	}
	if score > existing.HighestScore {
		setClause += ", highest_score = ?"
		args = append(args, score)
	}

	args = append(args, existing.ID)
	_, err := tx.Exec(
		fmt.Sprintf("UPDATE game_stats_levels SET %s WHERE id = ?", setClause),
		args...,
	)
	return err
}

// updateGameStatsOnLevelCompleteInTx updates game stats when a level completes (within tx).
func updateGameStatsOnLevelCompleteInTx(tx orm.Query, userID, gameID string, levelScore, playTimeSeconds, expEarned int) error {
	var stats models.GameStatsTotal
	if err := tx.Where("user_id", userID).Where("game_id", gameID).
		First(&stats); err != nil || stats.ID == "" {
		return fmt.Errorf("game stats not found for user %s, game %s", userID, gameID)
	}

	setClause := "total_play_time = total_play_time + ?, total_scores = total_scores + ?, total_exp = total_exp + ?, updated_at = now()"
	args := []any{playTimeSeconds, levelScore, expEarned}

	if levelScore > stats.HighestScore {
		setClause += ", highest_score = ?"
		args = append(args, levelScore)
	}

	args = append(args, stats.ID)
	_, err := tx.Exec(
		fmt.Sprintf("UPDATE game_stats_totals SET %s WHERE id = ?", setClause),
		args...,
	)
	return err
}
