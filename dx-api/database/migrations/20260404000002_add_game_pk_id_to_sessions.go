package migrations

import "github.com/goravel/framework/facades"

type M20260404000002AddGamePkIdToSessions struct{}

func (r *M20260404000002AddGamePkIdToSessions) Signature() string {
	return "20260404000002_add_game_pk_id_to_sessions"
}

func (r *M20260404000002AddGamePkIdToSessions) Up() error {
	// Add game_pk_id to game_session_totals
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_session_totals ADD COLUMN IF NOT EXISTS game_pk_id UUID`); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_totals_game_pk_id ON game_session_totals (game_pk_id) WHERE game_pk_id IS NOT NULL`); err != nil {
		return err
	}

	// Add game_pk_id to game_session_levels
	if _, err := facades.Orm().Query().Exec(
		`ALTER TABLE game_session_levels ADD COLUMN IF NOT EXISTS game_pk_id UUID`); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_game_session_levels_game_pk_id ON game_session_levels (game_pk_id) WHERE game_pk_id IS NOT NULL`); err != nil {
		return err
	}

	// Add unique active session index for PK (prevents duplicate sessions per PK)
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_session_totals_unique_active_pk
		 ON game_session_totals (user_id, game_id, degree, COALESCE(pattern, ''), game_pk_id)
		 WHERE ended_at IS NULL AND game_pk_id IS NOT NULL`); err != nil {
		return err
	}

	return nil
}

func (r *M20260404000002AddGamePkIdToSessions) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_unique_active_pk`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_levels_game_pk_id`)
	facades.Orm().Query().Exec(`ALTER TABLE game_session_levels DROP COLUMN IF EXISTS game_pk_id`)
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_session_totals_game_pk_id`)
	facades.Orm().Query().Exec(`ALTER TABLE game_session_totals DROP COLUMN IF EXISTS game_pk_id`)
	return nil
}
