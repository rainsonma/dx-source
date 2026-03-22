package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000022CreateAdmMenusTable struct{}

func (r *M20260322000022CreateAdmMenusTable) Signature() string {
	return "20260322000022_create_adm_menus_table"
}

func (r *M20260322000022CreateAdmMenusTable) Up() error {
	if !facades.Schema().HasTable("adm_menus") {
		return facades.Schema().Create("adm_menus", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("parent_id").Nullable()
			table.Text("name").Default("")
			table.Text("alias").Nullable()
			table.Text("icon").Nullable()
			table.Text("uri").Nullable()
			table.Double("order").Default(0)
			table.TimestampsTz()
			table.Index("parent_id", "order")
		})
	}
	return nil
}

func (r *M20260322000022CreateAdmMenusTable) Down() error {
	return facades.Schema().DropIfExists("adm_menus")
}
