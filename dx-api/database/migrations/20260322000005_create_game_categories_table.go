package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000005CreateGameCategoriesTable struct{}

func (r *M20260322000005CreateGameCategoriesTable) Signature() string {
	return "20260322000005_create_game_categories_table"
}

func (r *M20260322000005CreateGameCategoriesTable) Up() error {
	if !facades.Schema().HasTable("game_categories") {
		return facades.Schema().Create("game_categories", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("parent_id").Nullable()
			table.Uuid("cover_id").Nullable()
			table.Text("name").Default("")
			table.Text("alias").Nullable()
			table.Text("description").Nullable()
			table.Double("order").Default(0)
			table.Boolean("is_enabled").Default(true)
			table.TimestampsTz()
			table.Unique("name")
			table.Index("cover_id")
			table.Index("alias")
			table.Index("parent_id")
			table.Index("order")
		})
	}
	return nil
}

func (r *M20260322000005CreateGameCategoriesTable) Down() error {
	return facades.Schema().DropIfExists("game_categories")
}
