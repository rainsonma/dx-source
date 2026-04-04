package migrations

import "github.com/goravel/framework/facades"

type M20260325000002AddSessionIndexes struct{}

func (r *M20260325000002AddSessionIndexes) Signature() string {
	return "20260325000002_add_session_indexes"
}

func (r *M20260325000002AddSessionIndexes) Up() error {
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_regular
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''))
		 WHERE ended_at IS NULL AND game_group_id IS NULL AND game_pk_id IS NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_group
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''), game_group_id)
		 WHERE ended_at IS NULL AND game_group_id IS NOT NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_totals_group
		 ON game_session_totals (game_group_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_levels_group_level
		 ON game_session_levels (game_group_id, game_level_id)
		 WHERE game_group_id IS NOT NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_pk
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''), game_pk_id)
		 WHERE ended_at IS NULL AND game_pk_id IS NOT NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_totals_pk
		 ON game_session_totals (game_pk_id)
		 WHERE game_pk_id IS NOT NULL`); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_levels_pk_level
		 ON game_session_levels (game_pk_id, game_level_id)
		 WHERE game_pk_id IS NOT NULL`); err != nil {
		return err
	}

	return nil
}

func (r *M20260325000002AddSessionIndexes) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_levels_pk_level`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_pk`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_pk`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_levels_group_level`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_group`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_group`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_regular`)
	return nil
}
