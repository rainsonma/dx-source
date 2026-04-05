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
			table.Uuid("start_game_level_id").Nullable()
			table.Text("invite_code").Default("")
			table.Uuid("invite_qrcode_id").Nullable()
			table.TimestampTz("dismissed_at").Nullable()
			table.Integer("member_count").Default(0)
			table.Text("game_mode").Nullable()
			table.Boolean("is_playing").Default(false)
			table.TimestampsTz()
			table.Unique("invite_code")
			table.Index("owner_id")
			table.Index("cover_id")
			table.Index("current_game_id")
			table.Index("dismissed_at")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000028CreateGameGroupsTable) Down() error {
	return facades.Schema().DropIfExists("game_groups")
}
