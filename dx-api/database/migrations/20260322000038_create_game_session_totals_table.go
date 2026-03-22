package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000038CreateGameSessionTotalsTable struct{}

func (r *M20260322000038CreateGameSessionTotalsTable) Signature() string {
	return "20260322000038_create_game_session_totals_table"
}

func (r *M20260322000038CreateGameSessionTotalsTable) Up() error {
	if !facades.Schema().HasTable("game_session_totals") {
		return facades.Schema().Create("game_session_totals", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("game_id")
			table.Text("degree").Default("")
			table.Text("pattern").Nullable()
			table.Uuid("current_level_id").Nullable()
			table.Uuid("current_content_item_id").Nullable()
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
			table.Integer("total_levels_count").Default(0)
			table.Integer("played_levels_count").Default(0)
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("game_id")
			table.Index("current_level_id")
			table.Index("current_content_item_id")
			table.Index("user_id", "game_id", "degree", "pattern", "ended_at")
			table.Index("started_at")
		})
	}
	return nil
}

func (r *M20260322000038CreateGameSessionTotalsTable) Down() error {
	return facades.Schema().DropIfExists("game_session_totals")
}
