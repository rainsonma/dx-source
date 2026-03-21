package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000040_CreateGameRecordsTable struct{}

func (r *M20260322000040_CreateGameRecordsTable) Signature() string {
	return "20260322000040_create_game_records_table"
}

func (r *M20260322000040_CreateGameRecordsTable) Up() error {
	if !facades.Schema().HasTable("game_records") {
		return facades.Schema().Create("game_records", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("user_id")
			table.String("game_session_total_id")
			table.String("game_session_level_id")
			table.String("game_level_id")
			table.String("content_item_id")
			table.Boolean("is_correct").Default(false)
			table.Text("source_answer").Default("")
			table.Text("user_answer").Default("")
			table.Integer("base_score").Default(0)
			table.Integer("combo_score").Default(0)
			table.Integer("duration").Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000040_CreateGameRecordsTable) Down() error {
	return facades.Schema().DropIfExists("game_records")
}
