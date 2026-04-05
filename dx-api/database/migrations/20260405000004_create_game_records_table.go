package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260405000004CreateGameRecordsTable struct{}

func (r *M20260405000004CreateGameRecordsTable) Signature() string {
	return "20260405000004_create_game_records_table"
}

func (r *M20260405000004CreateGameRecordsTable) Up() error {
	return facades.Schema().Create("game_records", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("user_id")
		table.Uuid("game_session_id")
		table.Uuid("game_level_id")
		table.Uuid("content_item_id")
		table.Boolean("is_correct").Default(false)
		table.Text("source_answer").Default("")
		table.Text("user_answer").Default("")
		table.Integer("base_score").Default(0)
		table.Integer("combo_score").Default(0)
		table.Integer("duration").Default(0)
		table.TimestampsTz()
		table.Unique("game_session_id", "content_item_id")
		table.Index("user_id")
		table.Index("game_session_id")
		table.Index("game_level_id")
		table.Index("content_item_id")
		table.Index("is_correct")
	})
}

func (r *M20260405000004CreateGameRecordsTable) Down() error {
	return facades.Schema().DropIfExists("game_records")
}
