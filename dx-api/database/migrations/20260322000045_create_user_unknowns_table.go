package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000045CreateUserUnknownsTable struct{}

func (r *M20260322000045CreateUserUnknownsTable) Signature() string {
	return "20260322000045_create_user_unknowns_table"
}

func (r *M20260322000045CreateUserUnknownsTable) Up() error {
	if !facades.Schema().HasTable("user_unknowns") {
		return facades.Schema().Create("user_unknowns", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("content_item_id")
			table.String("game_id")
			table.String("game_level_id")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000045CreateUserUnknownsTable) Down() error {
	return facades.Schema().DropIfExists("user_unknowns")
}
