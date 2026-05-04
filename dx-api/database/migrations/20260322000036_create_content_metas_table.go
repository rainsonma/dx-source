package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000036CreateContentMetasTable struct{}

func (r *M20260322000036CreateContentMetasTable) Signature() string {
	return "20260322000036_create_content_metas_table"
}

func (r *M20260322000036CreateContentMetasTable) Up() error {
	if !facades.Schema().HasTable("content_metas") {
		return facades.Schema().Create("content_metas", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Text("source_from").Default("")
			table.Text("source_type").Default("")
			table.Text("source_data").Default("")
			table.Text("translation").Nullable()
			table.Text("speaker").Nullable()
			table.Boolean("is_break_done").Default(false)
			table.Double("order").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("source_from")
			table.Index("source_type")
			table.Index("created_at")
			table.Index("game_id")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_content_metas_level_order")
		})
	}
	return nil
}

func (r *M20260322000036CreateContentMetasTable) Down() error {
	return facades.Schema().DropIfExists("content_metas")
}
