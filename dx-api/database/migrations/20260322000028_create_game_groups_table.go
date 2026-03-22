package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000028CreateGameGroupsTable struct{}

func (r *M20260322000028CreateGameGroupsTable) Signature() string {
	return "20260322000028_create_game_groups_table"
}

func (r *M20260322000028CreateGameGroupsTable) Up() error {
	if !facades.Schema().HasTable("game_groups") {
		return facades.Schema().Create("game_groups", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("name").Default("")
			table.Text("description").Nullable()
			table.Uuid("owner_id")
			table.Uuid("cover_id").Nullable()
			table.Uuid("current_game_id").Nullable()
			table.Text("invite_code").Default("")
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.Unique("invite_code")
			table.Index("owner_id")
			table.Index("cover_id")
			table.Index("current_game_id")
			table.Index("is_active")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000028CreateGameGroupsTable) Down() error {
	return facades.Schema().DropIfExists("game_groups")
}
