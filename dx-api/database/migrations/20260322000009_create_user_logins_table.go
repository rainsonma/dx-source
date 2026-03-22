package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000009CreateUserLoginsTable struct{}

func (r *M20260322000009CreateUserLoginsTable) Signature() string {
	return "20260322000009_create_user_logins_table"
}

func (r *M20260322000009CreateUserLoginsTable) Up() error {
	if !facades.Schema().HasTable("user_logins") {
		return facades.Schema().Create("user_logins", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Text("ip").Default("")
			table.Text("agent").Nullable()
			table.Text("country").Nullable()
			table.Text("province").Nullable()
			table.Text("city").Nullable()
			table.Text("isp").Nullable()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("ip")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000009CreateUserLoginsTable) Down() error {
	return facades.Schema().DropIfExists("user_logins")
}
