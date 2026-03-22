package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000036CreateContentMetasTable struct{}

func (r *M20260322000036CreateContentMetasTable) Signature() string {
	return "20260322000036_create_content_metas_table"
}

func (r *M20260322000036CreateContentMetasTable) Up() error {
	if !facades.Schema().HasTable("content_metas") {
		return facades.Schema().Create("content_metas", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("game_level_id")
			table.String("source_from").Default("")
			table.String("source_type").Default("")
			table.Text("source_data").Default("")
			table.Text("translation").Nullable()
			table.Boolean("is_break_done").Default(false)
			table.Double("order").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000036CreateContentMetasTable) Down() error {
	return facades.Schema().DropIfExists("content_metas")
}
