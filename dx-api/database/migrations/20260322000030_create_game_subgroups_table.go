package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000030CreateGameSubgroupsTable struct{}

func (r *M20260322000030CreateGameSubgroupsTable) Signature() string {
	return "20260322000030_create_game_subgroups_table"
}

func (r *M20260322000030CreateGameSubgroupsTable) Up() error {
	if !facades.Schema().HasTable("game_subgroups") {
		return facades.Schema().Create("game_subgroups", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_group_id")
			table.Text("name").Default("")
			table.Text("description").Nullable()
			table.Double("order").Default(0)
			table.TimestampTz("last_won_at").Nullable()
			table.TimestampsTz()
			table.Index("game_group_id")
			table.Index("order")
		})
	}
	return nil
}

func (r *M20260322000030CreateGameSubgroupsTable) Down() error {
	return facades.Schema().DropIfExists("game_subgroups")
}
