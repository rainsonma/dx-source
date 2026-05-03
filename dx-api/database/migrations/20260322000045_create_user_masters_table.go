package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000045CreateUserMastersTable struct{}

func (r *M20260322000045CreateUserMastersTable) Signature() string {
	return "20260322000045_create_user_masters_table"
}

func (r *M20260322000045CreateUserMastersTable) Up() error {
	if !facades.Schema().HasTable("user_masters") {
		return facades.Schema().Create("user_masters", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("content_item_id").Nullable()
			table.Uuid("content_vocab_id").Nullable()
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.TimestampTz("mastered_at").Nullable()
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("game_id")
			table.Index("game_level_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000045CreateUserMastersTable) Down() error {
	return facades.Schema().DropIfExists("user_masters")
}
