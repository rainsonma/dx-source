package migrations

import "github.com/goravel/framework/facades"

type M20260324000003CreateCustomIndexes struct{}

func (r *M20260324000003CreateCustomIndexes) Signature() string {
	return "20260324000003_create_custom_indexes"
}

func (r *M20260324000003CreateCustomIndexes) Up() error {
	// Unique active session: one active session per user per game+degree+pattern
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''))
		 WHERE ended_at IS NULL`); err != nil {
		return err
	}

	// Partial index for group queries on session totals
	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_totals_group
		 ON game_session_totals (game_group_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	// Partial index for group winner determination on session levels
	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_levels_group_level
		 ON game_session_levels (game_group_id, game_level_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	return nil
}

func (r *M20260324000003CreateCustomIndexes) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_levels_group_level`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_group`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active`)
	return nil
}
