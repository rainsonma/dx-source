package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260324000002CreateGameGroupApplicationsTable struct{}

func (r *M20260324000002CreateGameGroupApplicationsTable) Signature() string {
	return "20260324000002_create_game_group_applications_table"
}

func (r *M20260324000002CreateGameGroupApplicationsTable) Up() error {
	if !facades.Schema().HasTable("game_group_applications") {
		return facades.Schema().Create("game_group_applications", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_group_id")
			table.Uuid("user_id")
			table.Text("status").Default("pending")
			table.TimestampsTz()
			table.Index("game_group_id", "user_id", "status")
			table.Index("game_group_id", "status")
		})
	}
	return nil
}

func (r *M20260324000002CreateGameGroupApplicationsTable) Down() error {
	return facades.Schema().DropIfExists("game_group_applications")
}
