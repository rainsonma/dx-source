package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000024_CreateAdmOperatesTable struct{}

func (r *M20260322000024_CreateAdmOperatesTable) Signature() string {
	return "20260322000024_create_adm_operates_table"
}

func (r *M20260322000024_CreateAdmOperatesTable) Up() error {
	if !facades.Schema().HasTable("adm_operates") {
		return facades.Schema().Create("adm_operates", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("adm_user_id")
			table.String("path").Default("")
			table.String("method").Default("")
			table.String("ip").Default("")
			table.Text("input").Default("")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000024_CreateAdmOperatesTable) Down() error {
	return facades.Schema().DropIfExists("adm_operates")
}
