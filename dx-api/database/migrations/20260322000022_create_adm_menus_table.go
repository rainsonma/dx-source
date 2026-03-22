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
			table.String("id")
			table.Primary("id")
			table.String("parent_id").Nullable()
			table.String("name").Default("")
			table.String("alias").Nullable()
			table.String("icon").Nullable()
			table.String("uri").Nullable()
			table.Double("order").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000022CreateAdmMenusTable) Down() error {
	return facades.Schema().DropIfExists("adm_menus")
}
