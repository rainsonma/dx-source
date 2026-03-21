package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/app/facades"
)

type M20260322000039_CreateGameSessionLevelsTable struct{}

func (r *M20260322000039_CreateGameSessionLevelsTable) Signature() string {
	return "20260322000039_create_game_session_levels_table"
}

func (r *M20260322000039_CreateGameSessionLevelsTable) Up() error {
	if !facades.Schema().HasTable("game_session_levels") {
		return facades.Schema().Create("game_session_levels", func(table schema.Blueprint) {
			table.String("id")
			table.Primary("id")
			table.String("game_session_total_id")
			table.String("game_level_id")
			table.String("current_content_item_id").Nullable()
			table.String("degree").Default("")
			table.String("pattern").Nullable()
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
		})
	}
	return nil
}

func (r *M20260322000039_CreateGameSessionLevelsTable) Down() error {
	return facades.Schema().DropIfExists("game_session_levels")
}
