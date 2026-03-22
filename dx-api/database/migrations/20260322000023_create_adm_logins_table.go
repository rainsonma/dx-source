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
			table.String("id")
			table.Primary("id")
			table.String("adm_user_id")
			table.String("ip").Default("")
			table.String("agent").Nullable()
			table.String("country").Nullable()
			table.String("province").Nullable()
			table.String("city").Nullable()
			table.String("isp").Nullable()
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
