package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260405000007AddGamePksRaw struct{}

func (r *M20260405000007AddGamePksRaw) Signature() string {
	return "20260405000007_add_game_pks_raw"
}

func (r *M20260405000007AddGamePksRaw) Up() error {
	_, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX idx_game_pks_unique_active ON game_pks (user_id, game_id) WHERE is_playing = true`)
	return err
}

func (r *M20260405000007AddGamePksRaw) Down() error {
	_, err := facades.Orm().Query().Exec(`DROP INDEX IF EXISTS idx_game_pks_unique_active`)
	return err
}
