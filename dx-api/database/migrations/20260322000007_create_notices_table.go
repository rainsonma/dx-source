package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000007CreateNoticesTable struct{}

func (r *M20260322000007CreateNoticesTable) Signature() string {
	return "20260322000007_create_notices_table"
}

func (r *M20260322000007CreateNoticesTable) Up() error {
	if !facades.Schema().HasTable("notices") {
		return facades.Schema().Create("notices", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("title").Default("")
			table.Text("content").Nullable()
			table.String("icon").Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000007CreateNoticesTable) Down() error {
	return facades.Schema().DropIfExists("notices")
}
