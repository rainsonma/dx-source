package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000020CreateAdmUserPermitsTable struct{}

func (r *M20260322000020CreateAdmUserPermitsTable) Signature() string {
	return "20260322000020_create_adm_user_permits_table"
}

func (r *M20260322000020CreateAdmUserPermitsTable) Up() error {
	if !facades.Schema().HasTable("adm_user_permits") {
		return facades.Schema().Create("adm_user_permits", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("adm_user_id")
			table.Text("adm_permit_id")
			table.TimestampsTz()
			table.Unique("adm_user_id", "adm_permit_id")
		})
	}
	return nil
}

func (r *M20260322000020CreateAdmUserPermitsTable) Down() error {
	return facades.Schema().DropIfExists("adm_user_permits")
}
