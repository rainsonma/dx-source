package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000006CreateGamePressesTable struct{}

func (r *M20260322000006CreateGamePressesTable) Signature() string {
	return "20260322000006_create_game_presses_table"
}

func (r *M20260322000006CreateGamePressesTable) Up() error {
	if !facades.Schema().HasTable("game_presses") {
		return facades.Schema().Create("game_presses", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("name").Default("")
			table.String("alias").Nullable()
			table.String("cover_id").Nullable()
			table.Double("order").Default(0)
			table.TimestampsTz()
			table.Unique("name")
			table.Index("cover_id")
			table.Index("order")
		})
	}
	return nil
}

func (r *M20260322000006CreateGamePressesTable) Down() error {
	return facades.Schema().DropIfExists("game_presses")
}
