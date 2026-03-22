package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000028CreateGameGroupsTable struct{}

func (r *M20260322000028CreateGameGroupsTable) Signature() string {
	return "20260322000028_create_game_groups_table"
}

func (r *M20260322000028CreateGameGroupsTable) Up() error {
	if !facades.Schema().HasTable("game_groups") {
		return facades.Schema().Create("game_groups", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("name").Default("")
			table.Text("description").Nullable()
			table.String("owner_id")
			table.String("cover_id").Nullable()
			table.String("current_game_id").Nullable()
			table.String("invite_code").Default("")
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000028CreateGameGroupsTable) Down() error {
	return facades.Schema().DropIfExists("game_groups")
}
