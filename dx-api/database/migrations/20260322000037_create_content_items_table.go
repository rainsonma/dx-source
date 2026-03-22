package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000037CreateContentItemsTable struct{}

func (r *M20260322000037CreateContentItemsTable) Signature() string {
	return "20260322000037_create_content_items_table"
}

func (r *M20260322000037CreateContentItemsTable) Up() error {
	if !facades.Schema().HasTable("content_items") {
		return facades.Schema().Create("content_items", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("game_level_id")
			table.Text("content_meta_id").Nullable()
			table.Text("content").Default("")
			table.Text("content_type").Default("")
			table.Text("uk_audio_id").Nullable()
			table.Text("us_audio_id").Nullable()
			table.Text("definition").Nullable()
			table.Text("translation").Nullable()
			table.Text("explanation").Nullable()
			table.Json("items").Nullable()
			table.Json("structure").Nullable()
			table.Double("order").Default(0)
			table.Column("tags", "text[]").Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.Index("game_level_id")
			table.Index("content_meta_id")
			table.Index("uk_audio_id")
			table.Index("us_audio_id")
			table.Index("content_type")
			table.Index("order")
			table.Index("is_active")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000037CreateContentItemsTable) Down() error {
	return facades.Schema().DropIfExists("content_items")
}
