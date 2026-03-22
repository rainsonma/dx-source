package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000043CreateGameReportsTable struct{}

func (r *M20260322000043CreateGameReportsTable) Signature() string {
	return "20260322000043_create_game_reports_table"
}

func (r *M20260322000043CreateGameReportsTable) Up() error {
	if !facades.Schema().HasTable("game_reports") {
		return facades.Schema().Create("game_reports", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_item_id")
			table.Text("reason").Default("")
			table.Text("note").Nullable()
			table.Integer("count").Default(0)
			table.TimestampsTz()
			table.Unique("user_id", "content_item_id", "reason")
			table.Index("user_id")
			table.Index("game_id")
			table.Index("content_item_id")
		})
	}
	return nil
}

func (r *M20260322000043CreateGameReportsTable) Down() error {
	return facades.Schema().DropIfExists("game_reports")
}
