package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000013_CreateUserRedeemsTable struct{}

func (r *M20260322000013_CreateUserRedeemsTable) Signature() string {
	return "20260322000013_create_user_redeems_table"
}

func (r *M20260322000013_CreateUserRedeemsTable) Up() error {
	if !facades.Schema().HasTable("user_redeems") {
		return facades.Schema().Create("user_redeems", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("code").Default("")
			table.String("grade").Default("")
			table.String("user_id").Nullable()
			table.TimestampTz("redeemed_at").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000013_CreateUserRedeemsTable) Down() error {
	return facades.Schema().DropIfExists("user_redeems")
}
