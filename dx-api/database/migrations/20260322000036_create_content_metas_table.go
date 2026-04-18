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
			table.Text("source_from").Default("")
			table.Text("source_type").Default("")
			table.Text("source_data").Default("")
			table.Text("translation").Nullable()
			table.Boolean("is_break_done").Default(false)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("source_from")
			table.Index("source_type")
			table.Index("created_at")

			// Supports the per-user dedup SELECT in SaveMetadataBatch:
			//   WHERE cm.deleted_at IS NULL
			//     AND cm.source_data IN ?
			//     AND cm.source_type IN ?
			//   AND <join to games via user_id>
			//
			// Column order is (source_data, source_type) because source_data has
			// ~millions of distinct values on our dataset while source_type only has
			// 2 ('sentence' | 'vocab'). Leading with the more-selective column
			// narrows the B-tree scan aggressively before the second-column filter
			// runs — ~2-5x faster than the reverse order on the 1.22M-row table.
			table.Index("source_data", "source_type", "deleted_at").Name("idx_content_metas_dedup_lookup")
		})
	}
	return nil
}

func (r *M20260322000036CreateContentMetasTable) Down() error {
	return facades.Schema().DropIfExists("content_metas")
}
