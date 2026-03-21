package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000019_CreateAdmUserRolesTable struct{}

func (r *M20260322000019_CreateAdmUserRolesTable) Signature() string {
	return "20260322000019_create_adm_user_roles_table"
}

func (r *M20260322000019_CreateAdmUserRolesTable) Up() error {
	if !facades.Schema().HasTable("adm_user_roles") {
		return facades.Schema().Create("adm_user_roles", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("adm_role_id")
			table.String("adm_user_id")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000019_CreateAdmUserRolesTable) Down() error {
	return facades.Schema().DropIfExists("adm_user_roles")
}
