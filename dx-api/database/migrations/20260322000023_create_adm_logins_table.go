package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000023CreateAdmLoginsTable struct{}

func (r *M20260322000023CreateAdmLoginsTable) Signature() string {
	return "20260322000023_create_adm_logins_table"
}

func (r *M20260322000023CreateAdmLoginsTable) Up() error {
	if !facades.Schema().HasTable("adm_logins") {
		return facades.Schema().Create("adm_logins", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("adm_user_id")
			table.Text("ip").Default("")
			table.Text("agent").Nullable()
			table.Text("country").Nullable()
			table.Text("province").Nullable()
			table.Text("city").Nullable()
			table.Text("isp").Nullable()
			table.TimestampsTz()
			table.Index("adm_user_id")
			table.Index("ip")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000023CreateAdmLoginsTable) Down() error {
	return facades.Schema().DropIfExists("adm_logins")
}
