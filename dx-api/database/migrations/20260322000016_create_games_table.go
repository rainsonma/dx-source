package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000016CreateGamesTable struct{}

func (r *M20260322000016CreateGamesTable) Signature() string {
	return "20260322000016_create_games_table"
}

func (r *M20260322000016CreateGamesTable) Up() error {
	if !facades.Schema().HasTable("games") {
		return facades.Schema().Create("games", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("name").Default("")
			table.Text("description").Nullable()
			table.String("user_id").Nullable()
			table.String("mode").Default("")
			table.String("game_category_id").Nullable()
			table.String("game_press_id").Nullable()
			table.String("icon").Nullable()
			table.String("cover_id").Nullable()
			table.Double("order").Default(0)
			table.Boolean("is_active").Default(true)
			table.String("status").Default("")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000016CreateGamesTable) Down() error {
	return facades.Schema().DropIfExists("games")
}
