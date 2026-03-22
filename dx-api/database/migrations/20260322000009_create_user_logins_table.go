package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000009CreateUserLoginsTable struct{}

func (r *M20260322000009CreateUserLoginsTable) Signature() string {
	return "20260322000009_create_user_logins_table"
}

func (r *M20260322000009CreateUserLoginsTable) Up() error {
	if !facades.Schema().HasTable("user_logins") {
		return facades.Schema().Create("user_logins", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("ip").Default("")
			table.String("agent").Nullable()
			table.String("country").Nullable()
			table.String("province").Nullable()
			table.String("city").Nullable()
			table.String("isp").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000009CreateUserLoginsTable) Down() error {
	return facades.Schema().DropIfExists("user_logins")
}
