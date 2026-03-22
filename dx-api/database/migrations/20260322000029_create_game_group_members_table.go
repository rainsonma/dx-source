package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000029CreateGameGroupMembersTable struct{}

func (r *M20260322000029CreateGameGroupMembersTable) Signature() string {
	return "20260322000029_create_game_group_members_table"
}

func (r *M20260322000029CreateGameGroupMembersTable) Up() error {
	if !facades.Schema().HasTable("game_group_members") {
		return facades.Schema().Create("game_group_members", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("game_group_id")
			table.String("user_id")
			table.String("role").Default("")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000029CreateGameGroupMembersTable) Down() error {
	return facades.Schema().DropIfExists("game_group_members")
}
