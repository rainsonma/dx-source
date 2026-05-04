package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000047CreateUserUnknownsTable struct{}

func (r *M20260322000047CreateUserUnknownsTable) Signature() string {
	return "20260322000047_create_user_unknowns_table"
}

func (r *M20260322000047CreateUserUnknownsTable) Up() error {
	if !facades.Schema().HasTable("user_unknowns") {
		return facades.Schema().Create("user_unknowns", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("content_item_id").Nullable()
			table.Uuid("content_vocab_id").Nullable()
			table.Uuid("game_id")
			table.Uuid("game_level_id")
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

func (r *M20260322000047CreateUserUnknownsTable) Down() error {
	return facades.Schema().DropIfExists("user_unknowns")
}
