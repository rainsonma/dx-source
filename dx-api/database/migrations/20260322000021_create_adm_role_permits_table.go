package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000021_CreateAdmRolePermitsTable struct{}

func (r *M20260322000021_CreateAdmRolePermitsTable) Signature() string {
	return "20260322000021_create_adm_role_permits_table"
}

func (r *M20260322000021_CreateAdmRolePermitsTable) Up() error {
	if !facades.Schema().HasTable("adm_role_permits") {
		return facades.Schema().Create("adm_role_permits", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("adm_role_id")
			table.String("adm_permit_id")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000021_CreateAdmRolePermitsTable) Down() error {
	return facades.Schema().DropIfExists("adm_role_permits")
}
