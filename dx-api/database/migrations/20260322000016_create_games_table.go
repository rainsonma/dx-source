package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000016CreateGamesTable struct{}

func (r *M20260322000016CreateGamesTable) Signature() string {
	return "20260322000016_create_games_table"
}

func (r *M20260322000016CreateGamesTable) Up() error {
	if !facades.Schema().HasTable("games") {
		return facades.Schema().Create("games", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("name").Default("")
			table.Text("description").Nullable()
			table.Uuid("user_id").Nullable()
			table.Text("mode").Default("")
			table.Uuid("game_category_id").Nullable()
			table.Uuid("game_press_id").Nullable()
			table.Text("icon").Nullable()
			table.Uuid("cover_id").Nullable()
			table.Double("order").Default(0)
			table.Boolean("is_active").Default(true)
			table.Text("status").Default("")
			table.TimestampsTz()
			table.Unique("name")
			table.Index("user_id")
			table.Index("mode")
			table.Index("game_category_id")
			table.Index("game_press_id")
			table.Index("order")
			table.Index("is_active")
			table.Index("cover_id")
		})
	}
	return nil
}

func (r *M20260322000016CreateGamesTable) Down() error {
	return facades.Schema().DropIfExists("games")
}
