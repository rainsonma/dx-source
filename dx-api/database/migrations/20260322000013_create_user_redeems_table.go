package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000013CreateUserRedeemsTable struct{}

func (r *M20260322000013CreateUserRedeemsTable) Signature() string {
	return "20260322000013_create_user_redeems_table"
}

func (r *M20260322000013CreateUserRedeemsTable) Up() error {
	if !facades.Schema().HasTable("user_redeems") {
		return facades.Schema().Create("user_redeems", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("code").Default("")
			table.Text("grade").Default("")
			table.Uuid("user_id").Nullable()
			table.TimestampTz("redeemed_at").Nullable()
			table.TimestampsTz()
			table.Unique("code")
			table.Index("user_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000013CreateUserRedeemsTable) Down() error {
	return facades.Schema().DropIfExists("user_redeems")
}
