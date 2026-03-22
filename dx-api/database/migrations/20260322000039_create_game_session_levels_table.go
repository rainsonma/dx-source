package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000039CreateGameSessionLevelsTable struct{}

func (r *M20260322000039CreateGameSessionLevelsTable) Signature() string {
	return "20260322000039_create_game_session_levels_table"
}

func (r *M20260322000039CreateGameSessionLevelsTable) Up() error {
	if !facades.Schema().HasTable("game_session_levels") {
		return facades.Schema().Create("game_session_levels", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("game_session_total_id")
			table.Text("game_level_id")
			table.Text("current_content_item_id").Nullable()
			table.Text("degree").Default("")
			table.Text("pattern").Nullable()
			table.TimestampTz("started_at")
			table.TimestampTz("last_played_at")
			table.TimestampTz("ended_at").Nullable()
			table.Integer("score").Default(0)
			table.Integer("exp").Default(0)
			table.Integer("max_combo").Default(0)
			table.Integer("correct_count").Default(0)
			table.Integer("wrong_count").Default(0)
			table.Integer("skip_count").Default(0)
			table.Integer("play_time").Default(0)
			table.Integer("total_items_count").Default(0)
			table.Integer("played_items_count").Default(0)
			table.TimestampsTz()
			table.Index("game_session_total_id")
			table.Index("game_level_id")
			table.Index("current_content_item_id")
		})
	}
	return nil
}

func (r *M20260322000039CreateGameSessionLevelsTable) Down() error {
	return facades.Schema().DropIfExists("game_session_levels")
}
