package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000020_CreateAdmUserPermitsTable struct{}

func (r *M20260322000020_CreateAdmUserPermitsTable) Signature() string {
	return "20260322000020_create_adm_user_permits_table"
}

func (r *M20260322000020_CreateAdmUserPermitsTable) Up() error {
	if !facades.Schema().HasTable("adm_user_permits") {
		return facades.Schema().Create("adm_user_permits", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("adm_user_id")
			table.String("adm_permit_id")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000020_CreateAdmUserPermitsTable) Down() error {
	return facades.Schema().DropIfExists("adm_user_permits")
}
