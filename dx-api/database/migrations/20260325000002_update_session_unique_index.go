package migrations

import "github.com/goravel/framework/facades"

type M20260325000002UpdateSessionUniqueIndex struct{}

func (r *M20260325000002UpdateSessionUniqueIndex) Signature() string {
	return "20260325000002_update_session_unique_index"
}

func (r *M20260325000002UpdateSessionUniqueIndex) Up() error {
	if _, err := facades.Orm().Query().Exec(
		`DROP INDEX IF EXISTS idx_game_session_totals_unique_active`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_regular
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''))
		 WHERE ended_at IS NULL AND game_group_id IS NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_group
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''), game_group_id)
		 WHERE ended_at IS NULL AND game_group_id IS NOT NULL`); err != nil {
		return err
	}

	return nil
}

func (r *M20260325000002UpdateSessionUniqueIndex) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_group`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_regular`)
	facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''))
		 WHERE ended_at IS NULL`)
	return nil
}
