package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000003CreateAdmRolesTable struct{}

func (r *M20260322000003CreateAdmRolesTable) Signature() string {
	return "20260322000003_create_adm_roles_table"
}

func (r *M20260322000003CreateAdmRolesTable) Up() error {
	if !facades.Schema().HasTable("adm_roles") {
		return facades.Schema().Create("adm_roles", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("slug").Default("")
			table.Text("name").Default("")
			table.TimestampsTz()
			table.Unique("slug")
			table.Unique("name")
		})
	}
	return nil
}

func (r *M20260322000003CreateAdmRolesTable) Down() error {
	return facades.Schema().DropIfExists("adm_roles")
}
