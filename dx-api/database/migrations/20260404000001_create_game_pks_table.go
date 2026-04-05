package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260404000001CreateGamePksTable struct{}

func (r *M20260404000001CreateGamePksTable) Signature() string {
	return "20260404000001_create_game_pks_table"
}

func (r *M20260404000001CreateGamePksTable) Up() error {
	if !facades.Schema().HasTable("game_pks") {
		if err := facades.Schema().Create("game_pks", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("opponent_id")
			table.Uuid("game_id")
			table.Text("degree").Default("")
			table.Text("pattern").Nullable()
			table.Text("robot_difficulty").Default("normal")
			table.Uuid("current_level_id").Nullable()
			table.Boolean("is_playing").Default(false)
			table.Uuid("last_winner_id").Nullable()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("opponent_id")
			table.Index("game_id")
			table.Index("is_playing")
		}); err != nil {
			return err
		}
	}
	// Only one active PK per user per game
	if _, err := facades.Orm().Query().Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_pks_unique_active
		 ON game_pks (user_id, game_id)
		 WHERE is_playing = true`); err != nil {
		return err
	}
	return nil
}

func (r *M20260404000001CreateGamePksTable) Down() error {
	return facades.Schema().DropIfExists("game_pks")
}
