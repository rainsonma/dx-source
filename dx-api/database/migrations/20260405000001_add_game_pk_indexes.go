package migrations

import "github.com/goravel/framework/facades"

type M20260405000001AddGamePkIndexes struct{}

func (r *M20260405000001AddGamePkIndexes) Signature() string {
	return "20260405000001_add_game_pk_indexes"
}

func (r *M20260405000001AddGamePkIndexes) Up() error {
	// Only one active PK per user per game
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_pks_unique_active
		 ON game_pks (user_id, game_id)
		 WHERE is_playing = true`); err != nil {
		return err
	}
	return nil
}

func (r *M20260405000001AddGamePkIndexes) Down() error {
	facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_pks_unique_active`)
	return nil
}
