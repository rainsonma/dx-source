package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000015_CreateUserSettingsTable struct{}

func (r *M20260322000015_CreateUserSettingsTable) Signature() string {
	return "20260322000015_create_user_settings_table"
}

func (r *M20260322000015_CreateUserSettingsTable) Up() error {
	if !facades.Schema().HasTable("user_settings") {
		return facades.Schema().Create("user_settings", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("group").Default("")
			table.String("key").Default("")
			table.Text("value").Default("")
			table.String("value_type").Default("")
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000015_CreateUserSettingsTable) Down() error {
	return facades.Schema().DropIfExists("user_settings")
}
