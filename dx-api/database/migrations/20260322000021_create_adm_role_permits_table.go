package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000021CreateAdmRolePermitsTable struct{}

func (r *M20260322000021CreateAdmRolePermitsTable) Signature() string {
	return "20260322000021_create_adm_role_permits_table"
}

func (r *M20260322000021CreateAdmRolePermitsTable) Up() error {
	if !facades.Schema().HasTable("adm_role_permits") {
		return facades.Schema().Create("adm_role_permits", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("adm_role_id")
			table.Uuid("adm_permit_id")
			table.TimestampsTz()
			table.Unique("adm_role_id", "adm_permit_id")
		})
	}
	return nil
}

func (r *M20260322000021CreateAdmRolePermitsTable) Down() error {
	return facades.Schema().DropIfExists("adm_role_permits")
}
