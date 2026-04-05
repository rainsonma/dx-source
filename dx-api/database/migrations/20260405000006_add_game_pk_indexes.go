package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260405000006AddGamePkIndexes struct{}

func (r *M20260405000006AddGamePkIndexes) Signature() string {
	return "20260405000006_add_game_pk_indexes"
}

func (r *M20260405000006AddGamePkIndexes) Up() error {
	_, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX idx_game_pks_unique_active ON game_pks (user_id, game_id) WHERE is_playing = true`)
	return err
}

func (r *M20260405000006AddGamePkIndexes) Down() error {
	_, err := facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_pks_unique_active`)
	return err
}
