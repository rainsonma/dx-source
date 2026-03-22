package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000045CreateUserUnknownsTable struct{}

func (r *M20260322000045CreateUserUnknownsTable) Signature() string {
	return "20260322000045_create_user_unknowns_table"
}

func (r *M20260322000045CreateUserUnknownsTable) Up() error {
	if !facades.Schema().HasTable("user_unknowns") {
		return facades.Schema().Create("user_unknowns", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("content_item_id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.TimestampsTz()
			table.Unique("user_id", "content_item_id")
			table.Index("user_id")
			table.Index("content_item_id")
			table.Index("game_id")
			table.Index("game_level_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000045CreateUserUnknownsTable) Down() error {
	return facades.Schema().DropIfExists("user_unknowns")
}
