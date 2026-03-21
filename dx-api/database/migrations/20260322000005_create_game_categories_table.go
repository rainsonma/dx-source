package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000005_CreateGameCategoriesTable struct{}

func (r *M20260322000005_CreateGameCategoriesTable) Signature() string {
	return "20260322000005_create_game_categories_table"
}

func (r *M20260322000005_CreateGameCategoriesTable) Up() error {
	if !facades.Schema().HasTable("game_categories") {
		return facades.Schema().Create("game_categories", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("parent_id").Nullable()
			table.String("cover_id").Nullable()
			table.String("name").Default("")
			table.String("alias").Nullable()
			table.String("description").Nullable()
			table.Double("order").Default(0)
			table.Boolean("is_enabled").Default(true)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000005_CreateGameCategoriesTable) Down() error {
	return facades.Schema().DropIfExists("game_categories")
}
