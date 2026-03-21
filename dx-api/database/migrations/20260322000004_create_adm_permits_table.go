package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000004_CreateAdmPermitsTable struct{}

func (r *M20260322000004_CreateAdmPermitsTable) Signature() string {
	return "20260322000004_create_adm_permits_table"
}

func (r *M20260322000004_CreateAdmPermitsTable) Up() error {
	if !facades.Schema().HasTable("adm_permits") {
		return facades.Schema().Create("adm_permits", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("slug").Default("")
			table.String("name").Default("")
			table.Column("http_methods", "text[]").Nullable()
			table.Column("http_paths", "text[]").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000004_CreateAdmPermitsTable) Down() error {
	return facades.Schema().DropIfExists("adm_permits")
}
