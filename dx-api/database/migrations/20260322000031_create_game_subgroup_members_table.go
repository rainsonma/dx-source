package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000031CreateGameSubgroupMembersTable struct{}

func (r *M20260322000031CreateGameSubgroupMembersTable) Signature() string {
	return "20260322000031_create_game_subgroup_members_table"
}

func (r *M20260322000031CreateGameSubgroupMembersTable) Up() error {
	if !facades.Schema().HasTable("game_subgroup_members") {
		return facades.Schema().Create("game_subgroup_members", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("game_subgroup_id")
			table.Text("user_id")
			table.Text("role").Default("")
			table.TimestampsTz()
			table.Unique("game_subgroup_id", "user_id")
			table.Index("game_subgroup_id")
			table.Index("user_id")
		})
	}
	return nil
}

func (r *M20260322000031CreateGameSubgroupMembersTable) Down() error {
	return facades.Schema().DropIfExists("game_subgroup_members")
}
