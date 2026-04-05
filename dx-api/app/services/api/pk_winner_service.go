package api

import (
	"time"

	"github.com/goravel/framework/facades"
)

// ForceEndPkLoser sets ended_at on the loser's session.
func ForceEndPkLoser(pkID, winnerUserID string) error {
	now := time.Now()
	_, err := facades.Orm().Query().Exec(
		`UPDATE game_sessions SET ended_at = ?, updated_at = now()
		 WHERE game_pk_id = ? AND user_id != ? AND ended_at IS NULL`,
		now, pkID, winnerUserID,
	)
	return err
}
