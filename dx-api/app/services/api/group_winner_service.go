package api

import (
	"time"

	"github.com/goravel/framework/facades"
)

// ForceEndGroupLosers sets ended_at on all sessions for this group level except the winner.
func ForceEndGroupLosers(gameGroupID, gameLevelID, winnerUserID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`UPDATE game_sessions SET ended_at = ?, updated_at = now()
		 WHERE game_group_id = ? AND game_level_id = ? AND user_id != ? AND ended_at IS NULL`,
		now, gameGroupID, gameLevelID, winnerUserID,
	)
	return err
}

// ForceEndGroupLosersExceptTeam ends sessions for all players not in the winning team.
func ForceEndGroupLosersExceptTeam(gameGroupID, gameLevelID, winnerSubgroupID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`UPDATE game_sessions SET ended_at = ?, updated_at = now()
		 WHERE game_group_id = ? AND game_level_id = ? AND (game_subgroup_id IS NULL OR game_subgroup_id != ?) AND ended_at IS NULL`,
		now, gameGroupID, gameLevelID, winnerSubgroupID,
	)
	return err
}
