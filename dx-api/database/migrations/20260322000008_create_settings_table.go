package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000008CreateSettingsTable struct{}

func (r *M20260322000008CreateSettingsTable) Signature() string {
	return "20260322000008_create_settings_table"
}

func (r *M20260322000008CreateSettingsTable) Up() error {
	if !facades.Schema().HasTable("settings") {
		return facades.Schema().Create("settings", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("group").Default("")
			table.Text("label").Nullable()
			table.Text("key").Default("")
			table.Text("value").Default("")
			table.Text("value_type").Default("")
			table.Text("value_from").Default("")
			table.Json("value_options").Default("{}")
			table.Text("description").Default("")
			table.Double("order").Default(0)
			table.Boolean("is_enabled").Default(true)
			table.TimestampsTz()
			table.Index("order")
		})
	}
	return nil
}

func (r *M20260322000008CreateSettingsTable) Down() error {
	return facades.Schema().DropIfExists("settings")
}
