package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000044CreateUserMastersTable struct{}

func (r *M20260322000044CreateUserMastersTable) Signature() string {
	return "20260322000044_create_user_masters_table"
}

func (r *M20260322000044CreateUserMastersTable) Up() error {
	if !facades.Schema().HasTable("user_masters") {
		return facades.Schema().Create("user_masters", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("content_item_id")
			table.String("game_id")
			table.String("game_level_id")
			table.TimestampTz("mastered_at").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000044CreateUserMastersTable) Down() error {
	return facades.Schema().DropIfExists("user_masters")
}
