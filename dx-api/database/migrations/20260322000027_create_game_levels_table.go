package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000027CreateGameLevelsTable struct{}

func (r *M20260322000027CreateGameLevelsTable) Signature() string {
	return "20260322000027_create_game_levels_table"
}

func (r *M20260322000027CreateGameLevelsTable) Up() error {
	if !facades.Schema().HasTable("game_levels") {
		return facades.Schema().Create("game_levels", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Text("name").Default("")
			table.Text("description").Nullable()
			table.Double("order").Default(0)
			table.Integer("passing_score").Default(0)
			table.Column("degrees", "text[]").Nullable()
			table.Boolean("is_active").Default(true)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("game_id")
			table.Index("order")
			table.Index("is_active")
		})
	}
	return nil
}

func (r *M20260322000027CreateGameLevelsTable) Down() error {
	return facades.Schema().DropIfExists("game_levels")
}
