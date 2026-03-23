package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260323000001AddUniqueActiveSessionIndex struct{}

func (r *M20260323000001AddUniqueActiveSessionIndex) Signature() string {
	return "20260323000001_add_unique_active_session_index"
}

func (r *M20260323000001AddUniqueActiveSessionIndex) Up() error {
	_, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''))
		 WHERE ended_at IS NULL`)
	return err
}

func (r *M20260323000001AddUniqueActiveSessionIndex) Down() error {
	_, err := facades.Orm().Query().Exec(
		`DROP INDEX IF EXISTS idx_game_session_totals_unique_active`)
	return err
}
