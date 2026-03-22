package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000024CreateAdmOperatesTable struct{}

func (r *M20260322000024CreateAdmOperatesTable) Signature() string {
	return "20260322000024_create_adm_operates_table"
}

func (r *M20260322000024CreateAdmOperatesTable) Up() error {
	if !facades.Schema().HasTable("adm_operates") {
		return facades.Schema().Create("adm_operates", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("adm_user_id")
			table.Text("path").Default("")
			table.Text("method").Default("")
			table.Text("ip").Default("")
			table.Text("input").Default("")
			table.TimestampsTz()
			table.Index("adm_user_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000024CreateAdmOperatesTable) Down() error {
	return facades.Schema().DropIfExists("adm_operates")
}
