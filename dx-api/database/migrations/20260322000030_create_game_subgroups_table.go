package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000030CreateGameSubgroupsTable struct{}

func (r *M20260322000030CreateGameSubgroupsTable) Signature() string {
	return "20260322000030_create_game_subgroups_table"
}

func (r *M20260322000030CreateGameSubgroupsTable) Up() error {
	if !facades.Schema().HasTable("game_subgroups") {
		return facades.Schema().Create("game_subgroups", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("game_group_id")
			table.String("name").Default("")
			table.Text("description").Nullable()
			table.Double("order").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000030CreateGameSubgroupsTable) Down() error {
	return facades.Schema().DropIfExists("game_subgroups")
}
