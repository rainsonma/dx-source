package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000003CreateAdmRolesTable struct{}

func (r *M20260322000003CreateAdmRolesTable) Signature() string {
	return "20260322000003_create_adm_roles_table"
}

func (r *M20260322000003CreateAdmRolesTable) Up() error {
	if !facades.Schema().HasTable("adm_roles") {
		return facades.Schema().Create("adm_roles", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("slug").Default("")
			table.String("name").Default("")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000003CreateAdmRolesTable) Down() error {
	return facades.Schema().DropIfExists("adm_roles")
}
