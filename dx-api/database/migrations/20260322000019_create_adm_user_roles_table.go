package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000019CreateAdmUserRolesTable struct{}

func (r *M20260322000019CreateAdmUserRolesTable) Signature() string {
	return "20260322000019_create_adm_user_roles_table"
}

func (r *M20260322000019CreateAdmUserRolesTable) Up() error {
	if !facades.Schema().HasTable("adm_user_roles") {
		return facades.Schema().Create("adm_user_roles", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("adm_role_id")
			table.Text("adm_user_id")
			table.TimestampsTz()
			table.Unique("adm_role_id", "adm_user_id")
		})
	}
	return nil
}

func (r *M20260322000019CreateAdmUserRolesTable) Down() error {
	return facades.Schema().DropIfExists("adm_user_roles")
}
