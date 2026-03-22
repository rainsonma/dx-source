package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000015CreateUserSettingsTable struct{}

func (r *M20260322000015CreateUserSettingsTable) Signature() string {
	return "20260322000015_create_user_settings_table"
}

func (r *M20260322000015CreateUserSettingsTable) Up() error {
	if !facades.Schema().HasTable("user_settings") {
		return facades.Schema().Create("user_settings", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Text("group").Default("")
			table.Text("key").Default("")
			table.Text("value").Default("")
			table.Text("value_type").Default("")
			table.TimestampsTz()
			table.Index("user_id")
		})
	}
	return nil
}

func (r *M20260322000015CreateUserSettingsTable) Down() error {
	return facades.Schema().DropIfExists("user_settings")
}
